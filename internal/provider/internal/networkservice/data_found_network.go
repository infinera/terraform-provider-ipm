package network

import (
	"context"
	"encoding/json"
	"strings"

	"terraform-provider-ipm/internal/ipm_pf"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &FoundNetworksDataSource{}
	_ datasource.DataSourceWithConfigure = &FoundNetworksDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewFoundNetworksDataSource() datasource.DataSource {
	return &FoundNetworksDataSource{}
}

// coffeesDataSource is the data source implementation.
type FoundNetworksDataSource struct {
	client *ipm_pf.Client
}

type FoundNetworksDataSourceData struct {
	Id                types.String                  `tfsdk:"id"`
	Href              types.String                  `tfsdk:"href"`
	Name                   types.String `tfsdk:"name"`
	ConstellationFrequency types.Int64  `tfsdk:"constellation_frequency"`
	Modulation             types.String `tfsdk:"modulation"`
	TcMode                 types.Bool   `tfsdk:"tc_mode"`
	Topology               types.String `tfsdk:"topology"`
	HubSelector       *ModuleSelector                `tfsdk:"hub_selector"`
	Networks          types.List                    `tfsdk:"networks"`
}

// Metadata returns the data source type name.
func (r *FoundNetworksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_found_networks"
}

func (d *FoundNetworksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network ID",
				Optional:    true,
			},
			"href": schema.StringAttribute{
				Description: "Network href",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Network Name",
				Optional:    true,
			},
			"constellation_frequency": schema.Int64Attribute{
				Description: "constellation_frequency",
				Optional:    true,
			},
			"modulation": schema.StringAttribute{
				Description: "modulation",
				Optional:    true,
			},
			"tc_mode": schema.BoolAttribute{
				Description: "tc_mode",
				Optional:    true,
			},
			"topology": schema.StringAttribute{
				Description: "topology",
				Optional:    true,
			},
			"hub_selector": schema.SingleNestedAttribute{
				Description: "selector",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"module_selector_by_module_id": schema.SingleNestedAttribute{
						Description: "module_selector_by_module_id",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"module_id": schema.StringAttribute{
								Description: "module_id",
								Optional:    true,
							},
						},
					},
					"module_selector_by_module_name": schema.SingleNestedAttribute{
						Description: "module_selector_by_module_name",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"module_name": schema.StringAttribute{
								Description: "module_name",
								Optional:    true,
							},
						},
					},
					"module_selector_by_module_mac": schema.SingleNestedAttribute{
						Description: "module_selector_by_module_mac",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"module_mac": schema.StringAttribute{
								Description: "module_mac",
								Optional:    true,
							},
						},
					},
					"module_selector_by_module_serial_number": schema.SingleNestedAttribute{
						Description: "module_selector_by_module_serial_number",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"module_serial_number": schema.StringAttribute{
								Description: "module_serial_number",
								Optional:    true,
							},
						},
					},
					"host_port_selector_by_name": schema.SingleNestedAttribute{
						Description: "host_port_selector_by_name",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"host_name": schema.StringAttribute{
								Description: "host_name",
								Optional:    true,
							},
							"host_port_name": schema.StringAttribute{
								Description: "host_port_name",
								Optional:    true,
							},
						},
					},
					"host_port_selector_by_port_id": schema.SingleNestedAttribute{
						Description: "host_port_selector_by_port_id",
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
							"sysname": schema.StringAttribute{
								Description: "sysname",
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
			"networks":schema.ListAttribute{
				Computed: true,
				ElementType: NetworkObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *FoundNetworksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *FoundNetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := FoundNetworksDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "FoundNetworksDataSource: get Network", map[string]interface{}{"network id": query.Id.ValueString()})
	queryString := "?content=expanded"
	if !query.Id.IsNull() {
		if strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") != 0 {
			queryString = "/" + query.Id.ValueString() + queryString
		}
	} else if !query.Href.IsNull() {
		queryString = queryString + "&q={\"href\":\"" + query.Href.ValueString() + "\"}"
	} else {
		queryString = queryString + "&q={\"hubModule.state.module.moduleName\":\"" + query.HubSelector.ModuleSelectorByModuleName.ModuleName.ValueString() + "\"}"
	}
	tflog.Debug(ctx, "FoundNetworksDataSource: get Network", map[string]interface{}{"queryString": "/xr-networks" + queryString})
	body, err := d.client.ExecuteIPMHttpCommand("GET", "/xr-networks"+queryString, nil)
	if err != nil {
		diags.AddError(
			"FoundNetworksDataSource: read ##: Error Get Network",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "FoundNetworksDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"FoundNetworksDataSource: read ##: Error Get NetworkResource",
			"Get:Could not get FoundNetworksDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	query.Networks = types.ListNull(NetworkObjectType())
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.Networks = types.ListValueMust(NetworkObjectType(), NetworksValue(data.([]interface{})))
	}
	case map[string]interface{}:
		network := []interface{}{data}
		query.Networks = types.ListValueMust(NetworkObjectType(), NetworksValue(network))
	default:
		// it's something else
	}
	tflog.Debug(ctx, "FoundNetworksDataSource: Network ", map[string]interface{}{"Networks": query.Networks})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "FoundNetworksDataSource: get ", map[string]interface{}{"Network": query})
}

