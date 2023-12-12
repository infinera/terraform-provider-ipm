package transportcapacity

import (
	"context"
	"encoding/json"
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
	_ datasource.DataSource              = &TransportCapacitiesDataSource{}
	_ datasource.DataSourceWithConfigure = &TransportCapacitiesDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewTransportCapacitiesDataSource() datasource.DataSource {
	return &TransportCapacitiesDataSource{}
}

// coffeesDataSource is the data source implementation.
type TransportCapacitiesDataSource struct {
	client *ipm_pf.Client
}

type TransportCapacitiesDataSourceData struct {
	Id                   types.String                       `tfsdk:"id"`
	TransportCapacities  []TransportCapacityResourceData    `tfsdk:"transport_capacities"`
}

// Metadata returns the data source type name.
func (r *TransportCapacitiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transport_capacities"
}

func (d *TransportCapacitiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of transport capacities",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "TransportCapacity ID",
				Optional: true,
			},
			"transport_capacities": schema.ListNestedAttribute{
				Description: "List of Transport Capacities",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: TransportCapacityDataSchemaAttributes(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TransportCapacitiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *TransportCapacitiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := TransportCapacitiesDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "TransportCapacitiesDataSource: get TransportCapacities", map[string]interface{}{"TransportCapacity id": query.Id.ValueString()})

	var body []byte
	var err error
	if query.Id.IsNull() || strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/transport-capacities?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/transport-capacities/"+query.Id.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"TransportCapacitiesDataSource: read ##: Error Get TransportCapacities",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "TransportCapacitiesDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TransportCapacitiesDataSource: read ##: Error Get NetworkResource",
			"Get:Could not get TransportCapacitiesDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	query.TransportCapacities = []TransportCapacityResourceData{}
	tflog.Debug(ctx, "TransportCapacitiesDataSource: get ", map[string]interface{}{"TransportCapacities": data})
	switch data.(type) {
	case []interface{}:
		// it's an array
		TransportCapacitiesData := data.([]interface{})
		for _, tcdata := range TransportCapacitiesData {
			tc := tcdata.(map[string]interface{})
			TransportCapacity := TransportCapacityResourceData{}
			TransportCapacity.Populate(tc, ctx, &diags, true)
			query.TransportCapacities = append(query.TransportCapacities, TransportCapacity)
		}
	case map[string]interface{}:
		TransportCapacityData := data.(map[string]interface{})
		TransportCapacity := TransportCapacityResourceData{}
		TransportCapacity.Populate(TransportCapacityData, ctx, &diags, true)
		query.TransportCapacities = append(query.TransportCapacities, TransportCapacity)
	default:
		// it's something else
	}
	tflog.Debug(ctx, "TransportCapacitiesDataSource: TransportCapacities ", map[string]interface{}{"TransportCapacities": query.TransportCapacities})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "TransportCapacitiesDataSource: get ", map[string]interface{}{"TransportCapacities": query})
}

func TransportCapacityDataSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the Network.",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		//Config           NetworkConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Description: "config",
			Computed:    true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "Network Name",
					Computed:    true,
				},
				"capacity_mode": schema.StringAttribute{
					Description: "modulation",
					Computed:    true,
				},
				"labels": schema.MapAttribute{
					Description: "labels",
					Computed:    true,
					ElementType: types.StringType,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: TCStateAttributeType(),
		},
		//LeafModules      types.List `tfsdk:"leaf_modules"`
		"end_points":schema.ListNestedAttribute{
			Computed:     true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"tc_id": schema.StringAttribute{
						Description: "Numeric identifier of the TC.",
						Computed:    true,
					},
					"id": schema.StringAttribute{
						Description: "Numeric identifier of the Endpoint.",
						Computed:    true,
					},
					"href": schema.StringAttribute{
						Description: "href",
						Computed:    true,
					},
							//Config           NetworkConfig `tfsdk:"config"`
					"config": schema.SingleNestedAttribute{
						Description: "config",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"capacity": schema.Int64Attribute{
								Description: "capacity Name",
								Computed:    true,
							},
							"selector": common.IfSelectorSchema(),
						},
					},
					//State     types.Object   `tfsdk:"state"`
					"state": schema.ObjectAttribute{
						Computed: true,
						AttributeTypes: TCEndpointStateAttributeType(),
					},
				},
			},
		},
		"capacity_links":schema.ListAttribute{
			Computed: true,
			ElementType: TCCapacityLinkObjectType(),
		},
	}
}

