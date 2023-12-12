package network

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
	_ datasource.DataSource              = &LeafModulesDataSource{}
	_ datasource.DataSourceWithConfigure = &LeafModulesDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewLeafModulesDataSource() datasource.DataSource {
	return &LeafModulesDataSource{}
}

// coffeesDataSource is the data source implementation.
type LeafModulesDataSource struct {
	client *ipm_pf.Client
}

type LeafModulesDataSourceData struct {
	NetworkId     types.String          `tfsdk:"network_id"`
	Modules       types.List            `tfsdk:"modules"`
}

// Metadata returns the data source type name.
func (r *LeafModulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_leaf_modules"
}

func (d *LeafModulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"network_id": schema.StringAttribute{
				Description: "Network ID",
				Optional:    true,
			},
			"modules":schema.ListAttribute{
				Computed: true,
				ElementType: NWModuleObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *LeafModulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *LeafModulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	query := LeafModulesDataSourceData{}
	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	if query.NetworkId.IsNull() {
			diags.AddError(
				"Error Get leaf modules",
				"LeafModulesDataSource: Could not get leaf modules for a network. Network ID is not specified",
			)
			return
	}

	tflog.Debug(ctx, "LeafModulesDataSource: get LeafModules", map[string]interface{}{"queryLeafModules": query})

	body, err := d.client.ExecuteIPMHttpCommand("GET", "/xr-networks/"+query.NetworkId.ValueString()+"/leafModules?content=expanded", nil)
	if err != nil {
		diags.AddError(
			"Error: read ##: Error Get Leaf Modules",
			"LeafModulesDataSource: Could not get, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "LeafModulesDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"LeafModulesDataSource: read ##: Error Get LeafModulesResource",
			"Get:Could not get LeafModulesDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	query.Modules = types.ListNull(NWModuleObjectType())
	if len(data.([]interface{})) > 0 {
		query.Modules = types.ListValueMust(NWModuleObjectType(), NWModulesValue(data.([]interface{})))
	}
	tflog.Debug(ctx, "LeafModulesDataSource: Hosts ", map[string]interface{}{"modules": query.Modules})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "LeafModulesDataSource: get ", map[string]interface{}{"LeafModules": query})
}


