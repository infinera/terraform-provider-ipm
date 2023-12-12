package host

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
	_ datasource.DataSource              = &HostsDataSource{}
	_ datasource.DataSourceWithConfigure = &HostsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewHostsDataSource() datasource.DataSource {
	return &HostsDataSource{}
}

// coffeesDataSource is the data source implementation.
type HostsDataSource struct {
	client *ipm_pf.Client
}

type HostsDataSourceData struct {
	Id     types.String          `tfsdk:"id"`
	Hosts  types.List            `tfsdk:"hosts"`
}

// Metadata returns the data source type name.
func (r *HostsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosts"
}

func (d *HostsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Host ID",
				Optional: true,
			},
			"hosts":schema.ListAttribute{
				Computed: true,
				ElementType: HostObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *HostsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *HostsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := HostsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "HostsDataSource: get Hosts", map[string]interface{}{"host id": query.Id.ValueString()})

	var body []byte
	var err error
	if query.Id.IsNull() || strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/hosts?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/hosts/"+query.Id.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"HostsDataSource: read ##: Error get  HostsDataSource",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "HostsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HostsDataSource: read ##: Error Get NetworkResource",
			"Get:Could not get HostsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	switch data.(type) {
		case []interface{}:
			if len(data.([]interface{})) > 0 {
				query.Hosts = types.ListValueMust(HostObjectType(), HostObjectsValue(data.([]interface{})))
		}
		case map[string]interface{}:
			hosts := []interface{}{data}
			query.Hosts = types.ListValueMust(HostObjectType(), HostObjectsValue(hosts))
		default:
			// it's something else
	}
	
	tflog.Debug(ctx, "HostsDataSource: Hosts ", map[string]interface{}{"Hosts": query.Hosts})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "HostsDataSource: get ", map[string]interface{}{"Hosts": query})
}

func HostDataSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the Host",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href of the host",
			Computed:    true,
		},
		//Config    NodeConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "name",
					Computed: true,
				},
				"managed_by": schema.StringAttribute{
					Description: "managed_by",
					Computed: true,
				},
				"location": schema.SingleNestedAttribute{
					Description: "location",
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"latitude": schema.Int64Attribute{
							Description: "latitude",
							Computed: true,
						},
						"longitude": schema.Int64Attribute{
							Description: "longitude",
							Computed: true,
						},
					},
				},
				"labels": schema.MapAttribute{
					Description: "labels",
					Computed: true,
					ElementType: types.StringType,
				},
				"selector": schema.SingleNestedAttribute{
					Description: "selector",
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"module_selector_by_module_id": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_id",
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"module_id": schema.StringAttribute{
									Description: "module_id",
									Computed: true,
								},
							},
						},
						"module_selector_by_module_name": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_name",
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"module_name": schema.StringAttribute{
									Description: "module_name",
									Computed: true,
								},
							},
						},
						"module_selector_by_module_mac": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_mac",
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"module_mac": schema.StringAttribute{
									Description: "module_mac",
									Computed: true,
								},
							},
						},
						"module_selector_by_module_serial_number": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_serial_number",
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"module_serial_number": schema.StringAttribute{
									Description: "module_serial_number",
									Computed: true,
								},
							},
						},
						"host_selector_by_host_chassis_id": schema.SingleNestedAttribute{
							Description: "host_selector_by_host_chassis_id",
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"chassis_id": schema.StringAttribute{
									Description: "host_name",
									Computed: true,
								},
								"chassis_id_subtype": schema.StringAttribute{
									Description: "chassis_id_subtype",
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: HostStateAttributeType(),
		},
		//Ports      types.List `tfsdk:"ports"`
		"ports":schema.ListAttribute{
			Computed: true,
			ElementType: HostPortObjectType(),
		},
	}
}

