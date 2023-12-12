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
	_ datasource.DataSource              = &ODUsDataSource{}
	_ datasource.DataSourceWithConfigure = &ODUsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewODUsDataSource() datasource.DataSource {
	return &ODUsDataSource{}
}

// coffeesDataSource is the data source implementation.
type ODUsDataSource struct {
	client *ipm_pf.Client
}

type ODUsDataSourceData struct {
	ModuleId types.String `tfsdk:"module_id"`
	OTUColId types.String `tfsdk:"otu_col_id"`
	ColId    types.String `tfsdk:"col_id"`
	ODUs     types.List   `tfsdk:"odus"`
}

// Metadata returns the data source type name.
func (r *ODUsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_odus"
}

func (d *ODUsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"module_id": schema.StringAttribute{
				Description: "module ID",
				Required:    true,
			},
			"otu_col_id": schema.StringAttribute{
				Description: "OTU Col ID",
				Optional:    true,
			},
			"col_id": schema.StringAttribute{
				Description: "ODU Col ID",
				Optional:    true,
			},
			"odus": schema.ListAttribute{
				Computed:    true,
				ElementType: ODUObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ODUsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *ODUsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := ODUsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.ModuleId.IsNull() || query.OTUColId.IsNull() {
		diags.AddError(
			"Error Get ODUs",
			"ODUsDataSource: Could not get ODU for module. Module ID or OTU ID is not specified",
		)
		return
	}
	tflog.Debug(ctx, "ODUsDataSource: get ODUs", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.ColId.IsNull() || strings.Compare(strings.ToUpper(query.ColId.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/modules/"+query.ModuleId.ValueString()+"/otus/"+query.OTUColId.ValueString()+"/odus?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/modules/"+query.ModuleId.ValueString()+"/otus/"+query.OTUColId.ValueString()+"/odus/"+query.ColId.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"ODUsDataSource: read ##: Error read ODUResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "ODUsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"ODUsDataSource: read ##: Error Get ODUResource",
			"Get:Could not get ODUsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "ODUsDataSource: get ", map[string]interface{}{"ODUs": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.ODUs = types.ListValueMust(ODUObjectType(), ODUObjectsValue(data.([]interface{})))
		}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.ODUs = types.ListValueMust(ODUObjectType(), ODUObjectsValue(networks))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "ODUsDataSource: get ", map[string]interface{}{"ODUs": query})
}
