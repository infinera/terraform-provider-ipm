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
	_ datasource.DataSource              = &FoundTransportCapacitiesDataSource{}
	_ datasource.DataSourceWithConfigure = &FoundTransportCapacitiesDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewFoundTransportCapacitiesDataSource() datasource.DataSource {
	return &FoundTransportCapacitiesDataSource{}
}

// coffeesDataSource is the data source implementation.
type FoundTransportCapacitiesDataSource struct {
	client *ipm_pf.Client
}

type FoundTransportCapacitiesDataSourceData struct {
	Id                   types.String                       `tfsdk:"id"`
	Href                 types.String                       `tfsdk:"href"`
	Name                 types.String                       `tfsdk:"name"`
	CapacityMode         types.String                       `tfsdk:"capacity_mode"`
	ASelector            *common.IfSelector                  `tfsdk:"a_selector"`
	ZSelector            *common.IfSelector                  `tfsdk:"z_selector"`
	FoundTransportCapacities  []TransportCapacityResourceData    `tfsdk:"transport_capacities"`
}

// Metadata returns the data source type name.
func (r *FoundTransportCapacitiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_found_transport_capacities"
}

func (d *FoundTransportCapacitiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of transport capacities",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "TransportCapacity ID",
				Optional: true,
			},
			"href": schema.StringAttribute{
				Description: "Network href",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "TC Name",
				Optional:    true,
			},
			"capacity_mode": schema.StringAttribute{
				Description: "capacity_mode",
				Optional:    true,
			},
			"a_selector": common.IfSelectorSchema(),
			"z_selector": common.IfSelectorSchema(),
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
func (d *FoundTransportCapacitiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *FoundTransportCapacitiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := FoundTransportCapacitiesDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "FoundTransportCapacitiesDataSource: get TransportCapacities", map[string]interface{}{"TransportCapacity id": query.Id.ValueString()})

	queryString := "?content=expanded"
	if !query.Id.IsNull() {
		if strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") != 0 {
			queryString = "/" + query.Id.ValueString() + queryString
		}
	} else if !query.Href.IsNull() {
		queryString = queryString + "&q={\"href\":\"" + query.Href.ValueString() + "\"}"
	} else {
		queryString = queryString + "&q={\"$and\":[{\"config.capacityMode\":\""+ query.CapacityMode.ValueString() + "\"},{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleName.moduleName\":\""+ query.ASelector.ModuleIfSelectorByModuleName.ModuleName.ValueString()+ "\",\"config.selector.moduleIfSelectorByModuleName.moduleClientIfAid\":\"" + query.ASelector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString()+ "\"}}},{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleName.moduleName\":\"" + query.ZSelector.ModuleIfSelectorByModuleName.ModuleName.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleName.moduleClientIfAid\":\""+ query.ZSelector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString() + "\"}}}]}"
	}

	tflog.Debug(ctx, "FoundNetworksDataSource: get Network", map[string]interface{}{"queryString": "/transport-capacities" + queryString})
	body, err := d.client.ExecuteIPMHttpCommand("GET", "/transport-capacities" + queryString, nil)
	if err != nil {
		diags.AddError(
			"FoundTransportCapacitiesDataSource: read ##: Error Get TransportCapacities",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "FoundTransportCapacitiesDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"FoundTransportCapacitiesDataSource: read ##: Error Get NetworkResource",
			"Get:Could not get FoundTransportCapacitiesDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	query.FoundTransportCapacities = []TransportCapacityResourceData{}
	tflog.Debug(ctx, "FoundTransportCapacitiesDataSource: get ", map[string]interface{}{"FoundTransportCapacities": data})
	switch data.(type) {
	case []interface{}:
		// it's an array
		FoundTransportCapacitiesData := data.([]interface{})
		for _, tcdata := range FoundTransportCapacitiesData {
			tc := tcdata.(map[string]interface{})
			TransportCapacity := TransportCapacityResourceData{}
			TransportCapacity.Populate(tc, ctx, &diags, true)
			query.FoundTransportCapacities = append(query.FoundTransportCapacities, TransportCapacity)
		}
	case map[string]interface{}:
		TransportCapacityData := data.(map[string]interface{})
		TransportCapacity := TransportCapacityResourceData{}
		TransportCapacity.Populate(TransportCapacityData, ctx, &diags, true)
		query.FoundTransportCapacities = append(query.FoundTransportCapacities, TransportCapacity)
	default:
		// it's something else
	}
	tflog.Debug(ctx, "FoundTransportCapacitiesDataSource: FoundTransportCapacities ", map[string]interface{}{"FoundTransportCapacities": query.FoundTransportCapacities})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "FoundTransportCapacitiesDataSource: get ", map[string]interface{}{"FoundTransportCapacities": query})
}


