package provider

import (
	"context"
	"os"

	"terraform-provider-ipm/internal/ipm_pf"
	//"terraform-provider-ipm/internal/provider/internal/common"
	network "terraform-provider-ipm/internal/provider/internal/networkservice"
	networkconnection "terraform-provider-ipm/internal/provider/internal/networkconnectionservice"
	transportcapacity "terraform-provider-ipm/internal/provider/internal/transportcapacityservice"
	host "terraform-provider-ipm/internal/provider/internal/hostservice"
	module "terraform-provider-ipm/internal/provider/internal/moduleservice"
	ndu "terraform-provider-ipm/internal/provider/internal/nduservice"
	event "terraform-provider-ipm/internal/provider/internal/eventservice"
	mqttServer "terraform-provider-ipm/internal/provider/internal/mqttserverservice"
	actions "terraform-provider-ipm/internal/provider/internal/actionservice"


	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &XRProvider{}

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &XRProvider{}
}

// provider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type XRProvider struct {
	version string
}

// providerData can be used to store data from the Terraform configuration.
type XRProviderModel struct {
	Username types.String `tfsdk:"username"`
	Host     types.String `tfsdk:"host"`
	Password types.String `tfsdk:"password"`
}

func (p *XRProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ipm"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *XRProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with XR",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for IPM API. May also be provided via ipm_host environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for IPM API. May also be provided via ipm_username environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for IPM API. May also be provided via ipm_password environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *XRProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config XRProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	tflog.Debug(ctx, "XRProvider: Configure", map[string]interface{}{"config": config})

	if resp.Diagnostics.HasError() {
		return
	}

	// User must provide a user to the provider
	if config.Host.IsUnknown() {
		// Cannot connect to client with an unknown Host value
		resp.Diagnostics.AddAttributeError(
			path.Root("Host"),
			"Unknown client Host",
			"The provider cannot create the IPM API client as there is an unknown configuration value for the API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ipm_HOST environment variable.",
		)
	}

	// User must provide a user to the provider
	if config.Username.IsUnknown() {
		// Cannot connect to client with an unknown  Username value
		resp.Diagnostics.AddAttributeError(
			path.Root("Username"),
			"Unknown client Username",
			"The provider cannot create the IPM API client as there is an unknown configuration value for the API Username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ipm_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		// Cannot connect to client with an unknown Password value
		resp.Diagnostics.AddAttributeError(
			path.Root("Password"),
			"Unknown client Password",
			"The provider cannot create the IPM API client as there is an unknown configuration value for the API Password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ipm_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("IPM_HOST")
	username := os.Getenv("IPM_USERNAME")
	password := os.Getenv("IPM_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing IPM API Host",
			"The provider cannot create the IPM API client as there is a missing or empty value for the IPM API host. "+
				"Set the host value in the configuration or use the ipm_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing IPM API Username",
			"The provider cannot create the IPM API client as there is a missing or empty value for the IPM API username. "+
				"Set the username value in the configuration or use the ipm_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing IPM API Password",
			"The provider cannot create the IPM API client as there is a missing or empty value for the IPM API password. "+
				"Set the password value in the configuration or use the ipm_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "ipm_host", host)
	ctx = tflog.SetField(ctx, "ipm_username", username)
	ctx = tflog.SetField(ctx, "ipm_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "ipm_password")

	tflog.Debug(ctx, "Creating XR client")

	// Create a new ipm client and set it to the provider client
	client, err := ipm_pf.NewClient(&host, &username, &password)

	if err != nil {
		resp.Diagnostics.AddError(
			"provider: Unable to create client",
			"Unable to create ipm client:\n\n"+err.Error(),
		)
		return
	}

	// Make the XR client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

// listen to wss 
/*
tflog.Debug(ctx, "****provider: WSS connection to " + os.Getenv("IPM_HOST") + os.Getenv("IPM_WSS"))
wssClient, error := common.NewWebSocketClient(os.Getenv("IPM_HOST"), os.Getenv("IPM_WSS"));
	if error != nil {
		resp.Diagnostics.AddError(
			"provider: Unable to connect to IPM WSS: " + os.Getenv("IPM_WSS"),
			"Unable to create ipm client:\n\n"+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "provider: ipm - Connecting WSS connection")
	wsClient := wssClient.Connect()
	if wsClient == nil {
		tflog.Debug(ctx, "provider: ipm - FAIL WSS connection to " + os.Getenv("IPM_HOST") + os.Getenv("IPM_WSS"))
	}*/
	//common.WSSConnect3()

	tflog.Debug(ctx, "provider: ipm - successful connection request")
}

// DataSources defines the data sources implemented in the provider.
func (p *XRProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		network.NewNetworksDataSource,
		network.NewFoundNetworksDataSource,
		network.NewHubModuleDataSource,
		network.NewLeafModulesDataSource,
		network.NewReachableModulesDataSource,
		networkconnection.NewNetworkConnectionsDataSource,
		networkconnection.NewFoundNetworkConnectionsDataSource,
		networkconnection.NewACsDataSource,
		networkconnection.NewLCsDataSource,
		networkconnection.NewNCEndpointsDataSource,
		host.NewHostsDataSource,
		host.NewHostPortsDataSource,
		transportcapacity.NewCapacityLinksDataSource,
		transportcapacity.NewTCEndpointsDataSource,
		transportcapacity.NewTransportCapacitiesDataSource,
		transportcapacity.NewFoundTransportCapacitiesDataSource,
		module.NewACsDataSource,
		module.NewLCsDataSource,
		module.NewLinePTPsDataSource,
		module.NewCarriersDataSource,
		module.NewDSCGsDataSource,
		module.NewDSCsDataSource,
		module.NewEClientsDataSource,
		module.NewModulesDataSource,
		module.NewODUsDataSource,
		module.NewOTUsDataSource,
		ndu.NewCarriersDataSource,
		ndu.NewEClientsDataSource,
		ndu.NewEDFAsDataSource,
		ndu.NewLCsDataSource,
		ndu.NewLinePTPsDataSource,
		ndu.NewNDUsDataSource,
		ndu.NewOTUsDataSource,
		ndu.NewPolPTPsDataSource,
		ndu.NewPortsDataSource,
		ndu.NewTOMsDataSource,
		ndu.NewTribPTPsDataSource,
		ndu.NewVOAsDataSource,
		ndu.NewXRsDataSource,
		event.NewEventsDataSource,
		event.NewFoundEventsDataSource,
		mqttServer.NewMQTTServerDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *XRProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		network.NewNetworkResource,
		network.NewHubModuleResource,
		network.NewLeafModuleResource,
		networkconnection.NewACResource,
		networkconnection.NewLCResource,
		networkconnection.NewNetworkConnectionResource,
		networkconnection.NewNCEndpointResource,
		transportcapacity.NewTransportCapacityResource,
		transportcapacity.NewTCCapacityLinkResource,
		host.NewHostResource,
		host.NewHostPortResource,
		module.NewACResource,
		module.NewCarrierResource,
		module.NewDSCGResource,
		module.NewDSCResource,
		module.NewEClientResource,
		module.NewLCResource,
		module.NewLinePTPResource,
		module.NewModuleResource,
		module.NewODUResource,
		module.NewOTUResource,
		ndu.NewCarrierResource,
		ndu.NewEClientResource,
		ndu.NewLinePTPResource,
		ndu.NewLEDsResource,
		ndu.NewLCResource,
		ndu.NewFanUnitResource,
		ndu.NewEDFAResource,
		ndu.NewVOAResource,
		ndu.NewNDUResource,
		ndu.NewOTUResource,
		ndu.NewPEMResource,
		ndu.NewPolPTPResource,
		ndu.NewPortResource,
		ndu.NewTOMResource,
		ndu.NewTribPTPResource,
		ndu.NewXRResource,
		event.NewEventResource,
		mqttServer.NewMQTTResource,
		actions.NewActionsResource,
	}
}
