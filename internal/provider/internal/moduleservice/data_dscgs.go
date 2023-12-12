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
	_ datasource.DataSource              = &DSCGsDataSource{}
	_ datasource.DataSourceWithConfigure = &DSCGsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewDSCGsDataSource() datasource.DataSource {
	return &DSCGsDataSource{}
}

// coffeesDataSource is the data source implementation.
type DSCGsDataSource struct {
	client *ipm_pf.Client
}

type DSCGsDataSourceData struct {
	ModuleId     types.String `tfsdk:"module_id"`
	LinePTPColId types.String `tfsdk:"line_ptp_col_id"`
	CarrierColId    types.String `tfsdk:"carrier_col_id"`
	ColId        types.String `tfsdk:"col_id"`
	DSCGs        types.List   `tfsdk:"dscgs"`
}

// Metadata returns the data source type name.
func (r *DSCGsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dscgs"
}

func (d *DSCGsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"module_id": schema.StringAttribute{
				Description: "module ID",
				Required:    true,
			},
			"line_ptp_col_id": schema.StringAttribute{
				Description: "Line PTP col ID",
				Required:    true,
			},
			"carrier_col_id": schema.StringAttribute{
				Description: "Carrier ID",
				Required:    true,
			},
			"col_id": schema.StringAttribute{
				Description: "DSCG Col ID",
				Optional:    true,
			},
			"dscgs": schema.ListAttribute{
				Computed:    true,
				ElementType: DSCGObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *DSCGsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *DSCGsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := DSCGsDataSourceData{}
	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.ModuleId.IsNull() || query.LinePTPColId.IsNull() || query.CarrierColId.IsNull() {
		diags.AddError(
			"Error Get DSCGs",
			"DSCGsDataSource: Could not get DSCG for a module. Module ID, Line PTP  Col ID or Carrier ID is not specified",
		)
		return
	}

	tflog.Debug(ctx, "DSCGsDataSource: get DSCGs", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.ColId.IsNull() || strings.Compare(strings.ToUpper(query.ColId.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/modules/"+query.ModuleId.ValueString()+"/linePtps/"+query.LinePTPColId.ValueString()+"/carriers/"+query.CarrierColId.ValueString()+"/dscgs?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/modules/"+query.ModuleId.ValueString()+"/linePtps/"+query.LinePTPColId.ValueString()+"/carriers/"+query.CarrierColId.ValueString()+"/dscgs/"+query.ColId.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"DSCGsDataSource: read ##: Error read DSCGResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "DSCGsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"DSCGsDataSource: read ##: Error Get DSCGResource",
			"Get:Could not get DSCGsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "DSCGsDataSource: get ", map[string]interface{}{"DSCGs": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.DSCGs = types.ListValueMust(DSCGObjectType(), DSCGObjectsValue(data.([]interface{})))
		}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.DSCGs = types.ListValueMust(DSCGObjectType(), DSCGObjectsValue(networks))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "DSCGsDataSource: get ", map[string]interface{}{"DSCGs": query})
}
