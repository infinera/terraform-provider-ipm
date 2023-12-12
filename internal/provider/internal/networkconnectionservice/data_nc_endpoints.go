package networkconnection

import (
	"context"
	"encoding/json"

	"terraform-provider-ipm/internal/ipm_pf"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &NCEndpointsDataSource{}
	_ datasource.DataSourceWithConfigure = &NCEndpointsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewNCEndpointsDataSource() datasource.DataSource {
	return &NCEndpointsDataSource{}
}

// coffeesDataSource is the data source implementation.
type NCEndpointsDataSource struct {
	client *ipm_pf.Client
}

type NCEndpointsDataSourceData struct {
	NCId       types.String          `tfsdk:"nc_id"`
	Endpoints  types.List            `tfsdk:"endpoints"`
}

// Metadata returns the data source type name.
func (r *NCEndpointsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nc_endpoints"
}

func (d *NCEndpointsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of NCEndpoint information",
		Attributes: map[string]schema.Attribute{
			"nc_id": schema.StringAttribute{
				Description: "NC ID",
				Optional: true,
			},
			"endpoints":schema.ListAttribute{
				Computed: true,
				ElementType: NCEndpointObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NCEndpointsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *NCEndpointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := NCEndpointsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.NCId.IsNull() {
		diags.AddError(
			"NCEndpointsDataSource: read ##: Error Read NCEndpoints Data",
			"Read:Could not read, Host IS is not specified",
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Debug(ctx, "NCEndpointsDataSource: get NCEndpoints", map[string]interface{}{"NCEndpoint id": query.NCId.ValueString()})

	body, err := d.client.ExecuteIPMHttpCommand("GET", "/network-connections/"+query.NCId.ValueString()+"/endpoints?content=expanded", nil)
	if err != nil {
		diags.AddError(
			"NCEndpointsDataSource: read ##: Error Read NCEndpointsDataSource",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "NCEndpointsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NCEndpointsDataSource: read ##: Error Get NetworkResource",
			"Get:Could not get NCEndpointsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	if len(data.([]interface{})) > 0 {
		query.Endpoints = types.ListValueMust(NCEndpointObjectType(), NCEndpointObjectsValue(data.([]interface{})))
	}

	tflog.Debug(ctx, "NCEndpointsDataSource: Hosts ", map[string]interface{}{"NCEndpoints": query.Endpoints})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "NCEndpointsDataSource: get ", map[string]interface{}{"NCEndpoints": query})
}


