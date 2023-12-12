package transportcapacity

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
	_ datasource.DataSource              = &TCEndpointsDataSource{}
	_ datasource.DataSourceWithConfigure = &TCEndpointsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewTCEndpointsDataSource() datasource.DataSource {
	return &TCEndpointsDataSource{}
}

// coffeesDataSource is the data source implementation.
type TCEndpointsDataSource struct {
	client *ipm_pf.Client
}

type TCEndpointsDataSourceData struct {
	TCId       types.String   `tfsdk:"tc_id"`
	Endpoints  types.List     `tfsdk:"endpoints"`
}

// Metadata returns the data source type name.
func (r *TCEndpointsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tc_endpoints"
}

func (d *TCEndpointsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of transport capacities",
		Attributes: map[string]schema.Attribute{
			"tc_id": schema.StringAttribute{
				Description: "TC ID",
				Optional: true,
			},
			"endpoints":schema.ListAttribute{
				Computed: true,
				ElementType: TCEndpointObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TCEndpointsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *TCEndpointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := TCEndpointsDataSourceData{}
	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.TCId.IsNull() {
		diags.AddError(
			"TCEndpointsDataSource: Error get TCEndpoints",
			"Get: Could not Get TCEndpoints. TC Id is not specified",
		)
		return
	}

	tflog.Debug(ctx, "TCEndpointsDataSource: get Endpoints", map[string]interface{}{"TCEndPoints id": query.TCId.ValueString()})

	body, err := d.client.ExecuteIPMHttpCommand("GET", "/transport-capacities/"+query.TCId.ValueString()+"/endpoints?content=expanded", nil)

	if err != nil {
		diags.AddError(
			"TCEndpointsDataSource: read ##: Error Get TCEndpointsDataSource",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "TCEndpointsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TCEndpointsDataSource: read ##: Error Get NetworkResource",
			"Get:Could not get TCEndpointsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.Endpoints = types.ListValueMust(TCEndpointObjectType(), TCEndpointObjectsValue(data.([]interface{})))
	}
	case map[string]interface{}:
		capacityLink := []interface{}{data}
		query.Endpoints = types.ListValueMust(TCEndpointObjectType(), TCEndpointObjectsValue(capacityLink))
	default:
		// it's something else
}
	tflog.Debug(ctx, "TCEndpointsDataSource: Endpoints ", map[string]interface{}{"Endpoints": query.Endpoints})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "TCEndpointsDataSource: get ", map[string]interface{}{"Endpoints": query})
}


