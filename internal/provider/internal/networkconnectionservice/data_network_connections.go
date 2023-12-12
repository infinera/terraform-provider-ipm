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
	_ datasource.DataSource              = &NetworkConnectionsDataSource{}
	_ datasource.DataSourceWithConfigure = &NetworkConnectionsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewNetworkConnectionsDataSource() datasource.DataSource {
	return &NetworkConnectionsDataSource{}
}

// coffeesDataSource is the data source implementation.
type NetworkConnectionsDataSource struct {
	client *ipm_pf.Client
}

type NetworkConnectionsDataSourceData struct {
	Id types.String      `tfsdk:"id"`
	NCs  types.List        `tfsdk:"ncs"`
}

// Metadata returns the data source type name.
func (r *NetworkConnectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_connections"
}

func (d *NetworkConnectionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "NC ID",
				Optional:    true,
			},
			"ncs":schema.ListAttribute{
				Computed: true,
				ElementType: NetworkConnectionObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NetworkConnectionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *NetworkConnectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := NetworkConnectionsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "NetworkConnectionsDataSource: get NetworkConnections", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.Id.IsNull() || strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/network-connections?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/network-connections/"+query.Id.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"NetworkConnectionsDataSource: read ##: Error Update NetworkConnectionResource",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "NetworkConnectionsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NetworkConnectionsDataSource: read ##: Error Get NetworkConnectionResource",
			"Get:Could not get NetworkConnectionsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	switch data.(type) {
		case []interface{}:
			if len(data.([]interface{})) > 0 {
				query.NCs = types.ListValueMust(NetworkConnectionObjectType(), NetworkConnectionObjectsValue(data.([]interface{})))
		}
		case map[string]interface{}:
			ncs := []interface{}{data}
			query.NCs = types.ListValueMust(NetworkConnectionObjectType(), NetworkConnectionObjectsValue(ncs))
		default:
			// it's something else
	}
	
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "NetworkConnectionsDataSource: get ", map[string]interface{}{"NetworkConnections": query})
}

