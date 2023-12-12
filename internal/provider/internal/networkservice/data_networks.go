package network

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
	_ datasource.DataSource              = &NetworksDataSource{}
	_ datasource.DataSourceWithConfigure = &NetworksDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewNetworksDataSource() datasource.DataSource {
	return &NetworksDataSource{}
}

// coffeesDataSource is the data source implementation.
type NetworksDataSource struct {
	client *ipm_pf.Client
}

type NetworksDataSourceData struct {
	Id        types.String  `tfsdk:"id"`
	Networks  types.List    `tfsdk:"networks"`
}

// Metadata returns the data source type name.
func (r *NetworksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

func (d *NetworksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network ID",
				Optional:    true,
			},
			"networks":schema.ListAttribute{
				Computed: true,
				ElementType: NetworkObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NetworksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *NetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := NetworksDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "NetworksDataSource: get Networks", map[string]interface{}{"network id": query.Id.ValueString()})

	var body []byte
	var err error
	if query.Id.IsNull() || strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/xr-networks?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/xr-networks/"+query.Id.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"NetworksDataSource: read ##: Error Get Networks",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "NetworksDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NetworksDataSource: read ##: Error Get NetworkResource",
			"Get:Could not get NetworksDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	query.Networks = types.ListNull(NetworkObjectType())
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.Networks = types.ListValueMust(NetworkObjectType(), NetworksValue(data.([]interface{})))
	}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.Networks = types.ListValueMust(NetworkObjectType(), NetworksValue(networks))
	default:
		// it's something else
	}
	tflog.Debug(ctx, "NetworksDataSource: Networks ", map[string]interface{}{"Networks": query.Networks})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "NetworksDataSource: get ", map[string]interface{}{"Networks": query})
}

