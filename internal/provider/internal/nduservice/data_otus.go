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
	_ datasource.DataSource              = &OTUsDataSource{}
	_ datasource.DataSourceWithConfigure = &OTUsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewOTUsDataSource() datasource.DataSource {
	return &OTUsDataSource{}
}

// coffeesDataSource is the data source implementation.
type OTUsDataSource struct {
	client *ipm_pf.Client
}

type OTUsDataSourceData struct {
	NDUId types.String `tfsdk:"ndu_id"`
	ColId types.String `tfsdk:"col_id"`
	OTUs  types.List   `tfsdk:"otus"`
}

// Metadata returns the data source type name.
func (r *OTUsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_otus"
}

func (d *OTUsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of ndus' carries information",
		Attributes: map[string]schema.Attribute{
			"ndu_id": schema.StringAttribute{
				Description: "ndu ID",
				Required:    true,
			},
			"col_id": schema.StringAttribute{
				Description: "col_id",
				Required:    true,
			},
			"otus": schema.ListAttribute{
				Computed:    true,
				ElementType: OTUObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *OTUsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *OTUsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := OTUsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.NDUId.IsNull() {
		diags.AddError(
			"Error Get OTUs",
			"OTUsDataSource: Could not get OTU for a ndu. ndu ID is not specified",
		)
		return
	}
	tflog.Debug(ctx, "OTUsDataSource: get OTUs", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.ColId.IsNull() || strings.Compare(strings.ToUpper(query.ColId.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/otus/?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/otus/"+query.ColId.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"OTUsDataSource: read ##: Error read OTUResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "OTUsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"OTUsDataSource: read ##: Error Get OTUResource",
			"Get:Could not get OTUsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "OTUsDataSource: get ", map[string]interface{}{"OTUs": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.OTUs = types.ListValueMust(OTUObjectType(), OTUObjectsValue(data.([]interface{})))
		}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.OTUs = types.ListValueMust(OTUObjectType(), OTUObjectsValue(networks))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "OTUsDataSource: get ", map[string]interface{}{"OTUs": query})
}
