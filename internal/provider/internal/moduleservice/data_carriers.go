package moduleservice

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
	_ datasource.DataSource              = &CarriersDataSource{}
	_ datasource.DataSourceWithConfigure = &CarriersDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewCarriersDataSource() datasource.DataSource {
	return &CarriersDataSource{}
}

// coffeesDataSource is the data source implementation.
type CarriersDataSource struct {
	client *ipm_pf.Client
}

type CarriersDataSourceData struct {
	ModuleId     types.String `tfsdk:"module_id"`
	LinePTPColId types.String `tfsdk:"line_ptp_col_id"`
	ColId        types.String `tfsdk:"col_id"`
	Carriers     types.List   `tfsdk:"carriers"`
}

// Metadata returns the data source type name.
func (r *CarriersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carriers"
}

func (d *CarriersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"module_id": schema.StringAttribute{
				Description: "module ID",
				Required:    true,
			},
			"line_ptp_col_id": schema.StringAttribute{
				Description: "line_ptp_col_id",
				Required:    true,
			},
			"col_id": schema.StringAttribute{
				Description: "carrier Col ID",
				Optional:    true,
			},
			"carriers": schema.ListAttribute{
				Computed:    true,
				ElementType: CarrierObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CarriersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *CarriersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := CarriersDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.ModuleId.IsNull() || query.LinePTPColId.IsNull() {
		diags.AddError(
			"Error Get Carriers",
			"CarriersDataSource: Could not get Carrier for a module. Module ID or Line PTP Col ID is not specified",
		)
		return
	}
	tflog.Debug(ctx, "CarriersDataSource: get Carriers", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.ColId.IsNull() || strings.Compare(strings.ToUpper(query.ColId.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/modules/"+query.ModuleId.ValueString()+"/linePtps/"+query.LinePTPColId.ValueString()+"/carriers?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/modules/"+query.ModuleId.ValueString()+"/linePtps/"+query.LinePTPColId.ValueString()+"/carriers/"+query.ColId.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"CarriersDataSource: read ##: Error read CarrierResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "CarriersDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"CarriersDataSource: read ##: Error Get CarrierResource",
			"Get:Could not get CarriersDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "CarriersDataSource: get ", map[string]interface{}{"Carriers": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.Carriers = types.ListValueMust(CarrierObjectType(), CarrierObjectsValue(data.([]interface{})))
		}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.Carriers = types.ListValueMust(CarrierObjectType(), CarrierObjectsValue(networks))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "CarriersDataSource: get ", map[string]interface{}{"Carriers": query})
}
