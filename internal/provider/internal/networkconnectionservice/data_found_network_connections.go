package networkconnection

import (
	"context"
	"encoding/json"
	//"go/types"
	"strconv"
	"strings"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &FoundNetworkConnectionsDataSource{}
	_ datasource.DataSourceWithConfigure = &FoundNetworkConnectionsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewFoundNetworkConnectionsDataSource() datasource.DataSource {
	return &FoundNetworkConnectionsDataSource{}
}

// coffeesDataSource is the data source implementation.
type FoundNetworkConnectionsDataSource struct {
	client *ipm_pf.Client
}

type FoundNetworkConnectionsDataSourceData struct {
	Id          types.String      `tfsdk:"id"`
	Href        types.String      `tfsdk:"href"`
	Capacity    types.Int64       `tfsdk:"capacity"`
	ServiceMode types.String      `tfsdk:"service_mode"`
	EndpointSelectors []common.IfSelector `tfsdk:"endpoint_selectors"`
	NCs         types.List        `tfsdk:"ncs"`
}

// Metadata returns the data source type name.
func (r *FoundNetworkConnectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_found_network_connections"
}

func (d *FoundNetworkConnectionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "NC ID",
				Optional:    true,
			},
			"href": schema.StringAttribute{
				Description: "Href of the Network Connection",
				Optional:    true,
			},
			"service_mode": schema.StringAttribute{
				Description: "service_mode",
				Optional:    true,
			},
			"capacity": schema.Int64Attribute{
				Description: "capacity",
				Optional:    true,
			},
			"endpoint_selectors": schema.ListNestedAttribute{
				Description: "List of NC Endpoints's selectors",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:  map[string]schema.Attribute{
							"module_if_selector_by_module_id": schema.SingleNestedAttribute{
								Description: "module_if_selector_by_module_id",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"module_id": schema.StringAttribute{
										Description: "module_id",
										Optional:    true,
									},
									"module_client_if_aid": schema.StringAttribute{
										Description: "module_client_if_aid",
										Optional:    true,
									},
								},
							},
							"module_if_selector_by_module_name": schema.SingleNestedAttribute{
								Description: "module_if_selector_by_module_name",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"module_name": schema.StringAttribute{
										Description: "module_name",
										Optional:    true,
									},
									"module_client_if_aid": schema.StringAttribute{
										Description: "module_client_if_aid",
										Optional:    true,
									},
								},
							},
							"module_if_selector_by_module_mac": schema.SingleNestedAttribute{
								Description: "module_if_selector_by_module_mac",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"module_mac": schema.StringAttribute{
										Description: "module_mac",
										Optional:    true,
									},
									"module_client_if_aid": schema.StringAttribute{
										Description: "module_client_if_aid",
										Optional:    true,
									},
								},
							},
							"module_if_selector_by_module_serial_number": schema.SingleNestedAttribute{
								Description: "module_if_selector_by_module_serial_number",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"module_serial_number": schema.StringAttribute{
										Description: "module_serial_number",
										Optional:    true,
									},
									"module_client_if_aid": schema.StringAttribute{
										Description: "module_client_if_aid",
										Optional:    true,
									},
								},
							},
							"host_port_selector_by_name": schema.SingleNestedAttribute{
								Description: "HostPort_port_selector_by_chassis_id",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"host_name": schema.StringAttribute{
										Description: "HostPort_name",
										Optional:    true,
									},
									"host_port_name": schema.StringAttribute{
										Description: "hostPortName",
										Optional:    true,
									},
								},
							},
							"host_port_selector_by_port_id": schema.SingleNestedAttribute{
								Description: "HostPort_port_selector_by_chassis_id",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"chassis_id_subtype": schema.StringAttribute{
										Description: "chassis_id_subtype",
										Optional:    true,
									},
									"chassis_id": schema.StringAttribute{
										Description: "chassis_id",
										Optional:    true,
									},
									"port_id_subtype": schema.StringAttribute{
										Description: "port_id_subtype",
										Optional:    true,
									},
									"port_id": schema.StringAttribute{
										Description: "port_id",
										Optional:    true,
									},
								},
							},
							"host_port_selector_by_sys_name": schema.SingleNestedAttribute{
								Description: "host_port_selector_by_sys_name",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"sys_name": schema.StringAttribute{
										Description: "sys_name",
										Optional:    true,
									},
									"port_id_subtype": schema.StringAttribute{
										Description: "port_id_subtype",
										Optional:    true,
									},
									"port_id": schema.StringAttribute{
										Description: "port_id",
										Optional:    true,
									},
								},
							},
							"host_port_selector_by_port_source_mac": schema.SingleNestedAttribute{
								Description: "host_port_selector_by_port_source_mac",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"port_source_mac": schema.StringAttribute{
										Description: "port_source_mac",
										Optional:    true,
									},
								},
							},
						},
					},
			},
			"ncs":schema.ListAttribute{
				Computed: true,
				ElementType: NetworkConnectionObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *FoundNetworkConnectionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *FoundNetworkConnectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := FoundNetworkConnectionsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "FoundNetworkConnectionsDataSource: get FoundNetworkConnections", map[string]interface{}{"queryNetworks": query})
  queryString := "?content=expanded"
	if !query.Id.IsNull() {
		if strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") != 0 {
			queryString = "/" + query.Id.ValueString() + queryString
		}
	} else if !query.Href.IsNull() {
		queryString = queryString + "&q={\"href\":\"" + query.Href.ValueString() + "\"}"
	} else {
		queryString = queryString + "&q={\"$and\":[{\"config.serviceMode\":\""+ query.ServiceMode.ValueString() + "\"}"
		selectorStr :=""
		for _, endpointSelector := range query.EndpointSelectors {
			selectorStr = selectorStr + ",{\"endpoints\":{\"$elemMatch\":{\"config.capacity\":"+ strconv.Itoa((int)(query.Capacity.ValueInt64())) + ",\"config.selector.moduleIfSelectorByModuleName.moduleName\":\"" + endpointSelector.ModuleIfSelectorByModuleName.ModuleName.ValueString() +"\",\"config.selector.moduleIfSelectorByModuleName.moduleClientIfAid\":\"" + endpointSelector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString() + "\"}}}"
		}
		selectorStr = selectorStr + "]}"
		queryString = queryString + selectorStr
	}
	tflog.Debug(ctx, "FoundNetworksDataSource: get Network", map[string]interface{}{"queryString": "/network-connections" + queryString})
	body, err := d.client.ExecuteIPMHttpCommand("GET", "/network-connections" + queryString, nil)
	if err != nil {
		diags.AddError(
			"FoundNetworkConnectionsDataSource: read ##: Error Update NetworkConnectionResource",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "FoundNetworkConnectionsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"FoundNetworkConnectionsDataSource: read ##: Error Get NetworkConnectionResource",
			"Get:Could not get FoundNetworkConnectionsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	switch data.(type) {
		case []interface{}:
			if len(data.([]interface{})) > 0 {
				query.NCs = types.ListValueMust(NetworkConnectionObjectType(), NetworkConnectionObjectsValue(data.([]interface{})))
		}
		case map[string]interface{}:
			ncs := []interface{}{data}
			query.NCs = types.ListValueMust(NetworkConnectionObjectType(), NetworkConnectionObjectsValue(ncs))
		default:
			// it's something else
	}
	
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "FoundNetworkConnectionsDataSource: get ", map[string]interface{}{"FoundNetworkConnections": query})
}

