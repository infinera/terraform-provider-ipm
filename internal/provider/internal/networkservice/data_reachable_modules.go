package network

import (
	"context"
	"encoding/json"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &ReachableModulesDataSource{}
	_ datasource.DataSourceWithConfigure = &ReachableModulesDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewReachableModulesDataSource() datasource.DataSource {
	return &ReachableModulesDataSource{}
}

// coffeesDataSource is the data source implementation.
type ReachableModulesDataSource struct {
	client *ipm_pf.Client
}

type ReachableModulesDataSourceData struct {
	NetworkId     types.String          `tfsdk:"network_id"`
	Modules       []ReachableModuleResourceData           `tfsdk:"modules"`
}

// Metadata returns the data source type name.
func (r *ReachableModulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reachable_modules"
}

func (d *ReachableModulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Reachable Modules of a network",
		Attributes: map[string]schema.Attribute{
			"network_id": schema.StringAttribute{
				Description: "Network ID",
				Optional:    true,
			},
			"modules": schema.ListNestedAttribute{
				Description: "List of Reachable modules",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: ReachableModuleDataSchemaAttributes(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ReachableModulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *ReachableModulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	query := ReachableModulesDataSourceData{}
	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	if query.NetworkId.IsNull() {
			diags.AddError(
				"Error Get leaf modules",
				"ReachableModulesDataSource: Could not get reachable modules for a network. Network ID is not specified",
			)
			return
	}

	tflog.Debug(ctx, "ReachableModulesDataSource: get ReachableModules", map[string]interface{}{"queryReachableModules": query})

	body, err := d.client.ExecuteIPMHttpCommand("GET", "/xr-networks/"+query.NetworkId.ValueString()+"/reachableModules?content=expanded", nil)
	if err != nil {
		diags.AddError(
			"Error: read ##: Error Get Leaf Modules",
			"ReachableModulesDataSource: Could not get, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "ReachableModulesDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"ReachableModulesDataSource: read ##: Error Get ReachableModulesResource",
			"Get:Could not get ReachableModulesDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "ReachableModulesDataSource: get ", map[string]interface{}{"ReachableModules": data})
	query.Modules = []ReachableModuleResourceData{}
	switch data.(type) {
		case []interface{}:
			// it's an array
			modulesData := data.([]interface{})
			for _, reachableData := range modulesData {
				reachableModule := reachableData.(map[string]interface{})
				module := ReachableModuleResourceData{}
				module.Populate(reachableModule, ctx, &diags, true)
				query.Modules = append(query.Modules, module)
			}
		case map[string]interface{}:
			reachableModule := data.(map[string]interface{})
			module := ReachableModuleResourceData{}
			module.Populate(reachableModule, ctx, &diags, true)
			query.Modules = append(query.Modules, module)
		default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "ReachableModulesDataSource: get ", map[string]interface{}{"ReachableModules": query})
}

func ReachableModuleDataSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"network_id": schema.StringAttribute{
			Description: "Numeric identifier of the Constellation Network.",
			Computed:    true,
		},
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the network module",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href of the network module",
			Computed:    true,
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: map[string]attr.Type{
				"managed_by": types.StringType,
				"lifecycle_state": types.StringType,
				"lifecycle_state_cause": types.ObjectType{AttrTypes: common.LifecycleStateCauseAttributeType() },
				"module": types.ObjectType{AttrTypes: NWModuleStateAttributeType() },
				"endpoints": types.ListType{ElemType:  NWModuleEndpointObjectType()},
			},
		},
	}
}

