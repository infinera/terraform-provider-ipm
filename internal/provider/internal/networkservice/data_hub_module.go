package network

import (
	"context"
	"encoding/json"

	"terraform-provider-ipm/internal/ipm_pf"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &HubModuleDataSource{}
	_ datasource.DataSourceWithConfigure = &HubModuleDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewHubModuleDataSource() datasource.DataSource {
	return &HubModuleDataSource{}
}

// coffeesDataSource is the data source implementation.
type HubModuleDataSource struct {
	client *ipm_pf.Client
}

type HubModuleDataSourceData struct {
	NetworkId     types.String          `tfsdk:"network_id"`
	Module        types.Object    `tfsdk:"module"`
}

// Metadata returns the data source type name.
func (r *HubModuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hub_module"
}

func (d *HubModuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"network_id": schema.StringAttribute{
				Description: "Network ID",
				Optional:    true,
			},
			"module": schema.ObjectAttribute{
				Description: "module",
				Computed:    true,
				AttributeTypes: NWModuleAttributeType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *HubModuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *HubModuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	query := HubModuleDataSourceData{}
	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	if query.NetworkId.IsNull() {
			diags.AddError(
				"Error Get hub module",
				"HubModuleDataSource: Could not get hub module for a network. Network ID is not specified",
			)
			return
	}

	tflog.Debug(ctx, "HubModuleDataSource: get HubModule", map[string]interface{}{"queryHubModule": query})

	body, err := d.client.ExecuteIPMHttpCommand("GET", "/xr-networks/"+query.NetworkId.ValueString()+"/hubModule?content=expanded", nil)
	if err != nil {
		diags.AddError(
			"Error: read ##: Error Get Leaf Modules",
			"HubModuleDataSource: Could not get, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "HubModuleDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HubModuleDataSource: read ##: Error Get HubModuleResource",
			"Get:Could not get HubModuleDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "HubModuleDataSource: get ", map[string]interface{}{"HubModule": data})
	query.Module = types.ObjectNull(NWModuleAttributeType())
	switch data.(type) {
		case []interface{}:
			if len(data.([]interface{})) > 0 {
				for _,k := range data.([]interface{}) {
					module := k.(map[string]interface{})
					query.Module = types.ObjectValueMust(NWModuleAttributeType(), NWModuleAttributeValue(module))
				}
		}
		case map[string]interface{}:
			query.Module = types.ObjectValueMust(NWModuleAttributeType(), NWModuleAttributeValue(data.(map[string]interface{})))
		default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "HubModuleDataSource: get ", map[string]interface{}{"HubModule": query})
}

func HubModuleDataSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"network_id": schema.StringAttribute{
			Description: "Numeric identifier of the Constellation Network.",
			Optional:    true,
		},
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the network module",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href of the network module",
			Computed:    true,
		},
		//Config    NodeConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"selector": schema.SingleNestedAttribute{
					Description: "selector",
					Computed:    true,
					Attributes: map[string]schema.Attribute{
						"module_selector_by_module_id": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_id",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"module_id": schema.StringAttribute{
									Description: "module_id",
									Computed:    true,
								},
							},
						},
						"module_selector_by_module_name": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_name",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"module_name": schema.StringAttribute{
									Description: "module_name",
									Computed:    true,
								},
							},
						},
						"module_selector_by_module_mac": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_mac",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"module_mac": schema.StringAttribute{
									Description: "module_mac",
									Computed:    true,
								},
							},
						},
						"module_selector_by_module_serial_number": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_serial_number",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"module_serial_number": schema.StringAttribute{
									Description: "module_serial_number",
									Computed:    true,
								},
							},
						},
						"host_port_selector_by_name": schema.SingleNestedAttribute{
							Description: "host_port_selector_by_name",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"host_name": schema.StringAttribute{
									Description: "host_name",
									Computed:    true,
								},
								"host_port_name": schema.StringAttribute{
									Description: "host_port_name",
									Computed:    true,
								},
							},
						},
						"host_port_selector_by_port_id": schema.SingleNestedAttribute{
							Description: "host_port_selector_by_port_id",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"chassis_id_subtype": schema.StringAttribute{
									Description: "chassis_id_subtype",
									Computed:    true,
								},
								"chassis_id": schema.StringAttribute{
									Description: "chassis_id",
									Computed:    true,
								},
								"port_id_subtype": schema.StringAttribute{
									Description: "port_id_subtype",
									Computed:    true,
								},
								"port_id": schema.StringAttribute{
									Description: "port_id",
									Computed:    true,
								},
							},
						},
						"host_port_selector_by_sys_name": schema.SingleNestedAttribute{
							Description: "host_port_selector_by_sys_name",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"sysname": schema.StringAttribute{
									Description: "sysname",
									Computed:    true,
								},
								"port_id_subtype": schema.StringAttribute{
									Description: "port_id_subtype",
									Computed:    true,
								},
								"port_id": schema.StringAttribute{
									Description: "port_id",
									Computed:    true,
								},
							},
						},
						"host_port_selector_by_port_source_mac": schema.SingleNestedAttribute{
							Description: "host_port_selector_by_port_source_mac",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"port_source_mac": schema.StringAttribute{
									Description: "port_source_mac",
									Computed:    true,
								},
							},
						},
					},
				},
				"module": schema.SingleNestedAttribute{
					Description: "module",
					Computed:    true,

					Attributes: map[string]schema.Attribute{
						"planned_capacity": schema.StringAttribute{
							Description: "planned_capacity",
							Computed:    true,
						},
						"traffic_mode": schema.StringAttribute{
							Description: "traffic_mode",
							Computed:    true,
						},
						"fiber_connection_mode": schema.StringAttribute{
							Description: "fiber_connection_mode",
							Computed:    true,
						},
						"fec_iterations": schema.StringAttribute{
							Description: "fec_iterations",
							Computed:    true,
						},
						"requested_nominal_psd_offset": schema.StringAttribute{
							Description: "requested_nominal_psd_offset",
							Computed:    true,
						},
						"tx_clp_target": schema.Int64Attribute{
							Description: "tx_clp_target",
							Computed:    true,
						},
						"max_dscs": schema.Int64Attribute{
							Description: "maxDSCs",
							Computed:    true,
						},
						"max_tx_dscs": schema.Int64Attribute{
							Description: "maxTxDSCs",
							Computed:    true,
						},
						"allowed_rx_cdscs": schema.ListAttribute{
							Description: "allowed_rx_cdscs",
							Computed:    true,
							ElementType: types.Int64Type,
						},
						"allowed_tx_cdscs": schema.ListAttribute{
							Description: "allowed_tx_cdscs",
							Computed:    true,
							ElementType: types.Int64Type,
						},
					},
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: NWModuleStateAttributeType(),
		},
	}
}

