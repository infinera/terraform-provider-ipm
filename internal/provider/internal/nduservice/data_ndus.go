package nduservice

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
	_ datasource.DataSource              = &NDUsDataSource{}
	_ datasource.DataSourceWithConfigure = &NDUsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewNDUsDataSource() datasource.DataSource {
	return &NDUsDataSource{}
}

// coffeesDataSource is the data source implementation.
type NDUsDataSource struct {
	client *ipm_pf.Client
}

type NDUsDataSourceData struct {
	Id   types.String `tfsdk:"id"`
	NDUs types.List   `tfsdk:"ndus"`
}

// Metadata returns the data source type name.
func (r *NDUsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndus"
}

func (d *NDUsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of ndus' information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "NDU ID",
				Optional:    true,
			},
			"ndus": schema.ListAttribute{
				Computed:    true,
				ElementType: NDUObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NDUsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *NDUsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := NDUsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "NDUsDataSource: get NDUs", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.Id.IsNull() || strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.Id.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"NDUsDataSource: read ##: Error read NDUResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "NDUsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NDUsDataSource: read ##: Error Get NDUResource",
			"Get:Could not get NDUsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "NDUsDataSource: get ", map[string]interface{}{"NDUs": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.NDUs = types.ListValueMust(NDUObjectType(), NDUObjectsValue(data.([]interface{})))
		}
	case map[string]interface{}:
		ndus := []interface{}{data}
		query.NDUs = types.ListValueMust(NDUObjectType(), NDUObjectsValue(ndus))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "NDUsDataSource: get ", map[string]interface{}{"NDUs": query})
}
