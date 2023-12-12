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
	_ datasource.DataSource              = &ACsDataSource{}
	_ datasource.DataSourceWithConfigure = &ACsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewACsDataSource() datasource.DataSource {
	return &ACsDataSource{}
}

// coffeesDataSource is the data source implementation.
type ACsDataSource struct {
	client *ipm_pf.Client
}

type ACsDataSourceData struct {
	Id   types.String     `tfsdk:"id"`
	ACs  types.List       `tfsdk:"acs"`
}

// Metadata returns the data source type name.
func (r *ACsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nc_acs"
}

func (d *ACsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network ID",
				Optional:    true,
			},
			"acs":schema.ListAttribute{
				Computed: true,
				ElementType: ACObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ACsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *ACsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := ACsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "ACsDataSource: get ACs", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.Id.IsNull() || strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/acs?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/acs/"+query.Id.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"ACsDataSource: read ##: Error read ACResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "ACsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"ACsDataSource: read ##: Error Get ACResource",
			"Get:Could not get ACsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "ACsDataSource: get ", map[string]interface{}{"ACs": data})
	switch data.(type) {
		case []interface{}:
			if len(data.([]interface{})) > 0 {
				query.ACs = types.ListValueMust(ACObjectType(), ACObjectsValue(data.([]interface{})))
		}
		case map[string]interface{}:
			networks := []interface{}{data}
			query.ACs = types.ListValueMust(ACObjectType(), ACObjectsValue(networks))
		default:
			// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "ACsDataSource: get ", map[string]interface{}{"ACs": query})
}
