package networkconnection

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
	_ datasource.DataSource              = &LCsDataSource{}
	_ datasource.DataSourceWithConfigure = &LCsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewLCsDataSource() datasource.DataSource {
	return &LCsDataSource{}
}

// coffeesDataSource is the data source implementation.
type LCsDataSource struct {
	client *ipm_pf.Client
}

type LCsDataSourceData struct {
	Id   types.String     `tfsdk:"id"`
	LCs  types.List       `tfsdk:"lcs"`
}

// Metadata returns the data source type name.
func (r *LCsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nc_lcs"
}

func (d *LCsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network ID",
				Optional:    true,
			},
			"lcs":schema.ListAttribute{
				Computed: true,
				ElementType: LCObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *LCsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *LCsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := LCsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "LCsDataSource: get LCs", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.Id.IsNull() || strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/lcs?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/lcs/"+query.Id.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"LCsDataSource: read ##: Error read LCResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "LCsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"LCsDataSource: read ##: Error Get LCResource",
			"Get:Could not get LCsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "LCsDataSource: get ", map[string]interface{}{"LCs": data})
	switch data.(type) {
		case []interface{}:
			if len(data.([]interface{})) > 0 {
				query.LCs = types.ListValueMust(LCObjectType(), LCObjectsValue(data.([]interface{})))
		}
		case map[string]interface{}:
			networks := []interface{}{data}
			query.LCs = types.ListValueMust(LCObjectType(), LCObjectsValue(networks))
		default:
			// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "LCsDataSource: get ", map[string]interface{}{"LCs": query})
}
