package network

import (
	"context"
	"encoding/json"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	//"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ReachableModuleResource{}
	_ resource.ResourceWithConfigure   = &ReachableModuleResource{}
	_ resource.ResourceWithImportState = &ReachableModuleResource{}
)

// NewReachableModuleResource is a helper function to simplify the provider implementation.
func NewReachableModuleResource() resource.Resource {
	return &ReachableModuleResource{}
}

type ReachableModuleResource struct {
	client *ipm_pf.Client
}

type ReachableModuleResourceData struct {
	NetworkId types.String `tfsdk:"network_id"`
	Id        types.String `tfsdk:"id"`
	Href      types.String `tfsdk:"href"`
	State     types.Object   `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *ReachableModuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reachable_module"
}

// Schema defines the schema for the data source.
func (r *ReachableModuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type ReachableModuleResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an Reachable Module",
		Attributes:  ReachableModuleSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *ReachableModuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r ReachableModuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ReachableModuleResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "ReachableModuleResource: Create - ", map[string]interface{}{"ReachableModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)

}

func (r ReachableModuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ReachableModuleResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "ReachableModuleResource: Create - ", map[string]interface{}{"ReachableModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r ReachableModuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ReachableModuleResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"ReachableModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r ReachableModuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ReachableModuleResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "CfgResource: Delete", map[string]interface{}{"ReachableModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ReachableModuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ReachableModuleResource) create(plan *ReachableModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "ReachableModuleResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *ReachableModuleResource) update(plan *ReachableModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "ReachableModuleResource: update ## ", map[string]interface{}{"plan": plan})
}

func (r *ReachableModuleResource) read(state *ReachableModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.NetworkId.IsNull() || state.Id.IsNull() {
		diags.AddError(
			"Error Update Module",
			"Update: Could not Update  NetworkId , Module ID or Role is not specified.",
		)
		return
	}
	var body []byte
	var err error

	body, err = r.client.ExecuteIPMHttpCommand("GET", "/xr-Modules/"+state.NetworkId.ValueString()+"/reachableModules/"+state.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"ReachableModuleResource: read ##: Error Update ReachableModuleResource",
			"Update:Could not read ReachableModuleResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "ReachableModuleResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"ReachableModuleResource: read ##: Error Unmarshal response",
			"Update:Could not read ReachableModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	// populate state
	state.Populate(data, ctx, diags)

	tflog.Debug(ctx, "ReachableModuleResource: read ## ", map[string]interface{}{"plan": state})
}

func (r *ReachableModuleResource) delete(plan *ReachableModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.NetworkId.IsNull() || plan.Id.IsNull() {
		diags.AddError(
			"Error Delete ReachableModuleResource",
			"Read: Could not delete. Network ID or Module Id is not specified",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", "/xr-modules/"+plan.NetworkId.ValueString()+"/reachableModules/"+plan.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"ReachableModuleResource: delete ##: Error Delete ReachableModuleResource",
			"Update:Could not delete ReachableModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "ReachableModuleResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (mData *ReachableModuleResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeEntity_optional ...bool) {

	computeFlag := false
	if len(computeEntity_optional) > 0 {
		computeFlag = computeEntity_optional[0]
	}

	if !mData.Id.IsNull() || (computeFlag && data["id"] != nil) {
		mData.Id = types.StringValue(data["id"].(string))
	}
	if !mData.NetworkId.IsNull() || (computeFlag && data["parentId"] != nil) {
		mData.NetworkId = types.StringValue(data["parentId"].(string))
	}

	tflog.Debug(ctx, "ReachableModuleResourceData: populate Success ")
	// populate state
	if data["state"] != nil {
		state := data["state"].(map[string]interface{})
		lifecycleState := types.StringNull()
		if state["lifecycleState"] != nil {
			lifecycleState = types.StringValue(state["lifecycleState"].(string))
		}
		managedBy := types.StringNull()
		if state["managedBy"] != nil {
			managedBy = types.StringValue(state["managedBy"].(string))
		}
		endpoints := types.ListNull(NWModuleEndpointObjectType())
		if state["endpoints"] != nil {
			endpoints =  types.ListValueMust(NWModuleEndpointObjectType(), NWModuleEndpointsValue(state["endpoints"].([]interface{})))
		}
		lifecycleStateCause := types.ObjectNull(common.LifecycleStateCauseAttributeType())
		if state["lifecycleStateCause"] != nil {
			lifecycleStateCause =  types.ObjectValueMust( common.LifecycleStateCauseAttributeType(),
			common.LifecycleStateCauseAttributeValue(state["lifecycleStateCause"].(map[string]interface{})))
		}
		mData.State =types.ObjectValueMust(
			map[string]attr.Type{
				"module"  : types.ObjectType {AttrTypes:NWModuleStateAttributeType()},
				"endpoints" : types.ListType { ElemType:NWModuleEndpointObjectType()},
				"lifecycle_state":    types.StringType,
				"managed_by": types.StringType,
				"lifecycle_state_cause" :types.ObjectType {AttrTypes:common.LifecycleStateCauseAttributeType()},
			},
			map[string]attr.Value{
				"module": types.ObjectValueMust(NWModuleStateAttributeType(), NWModuleStateAttributeValue(state["module"].(map[string]interface{}))),
				"endpoints" : endpoints,
				"lifecycle_state": lifecycleState,
				"managed_by": managedBy,
				"lifecycle_state_cause" : lifecycleStateCause,
			},
		)
	}
}

func ComputedOnlyReachableModuleSchemaAttributes() map[string]schema.Attribute {
	return ReachableModuleSchemaAttributes(true)
}

func ReachableModuleSchemaAttributes(computeEntity_optional ...bool) map[string]schema.Attribute {
	computeFlag := true
	optionalFlag := false
	if len(computeEntity_optional) > 0 {
		computeFlag = computeEntity_optional[0]
		optionalFlag = !computeFlag
	}
	return map[string]schema.Attribute{
		"network_id": schema.StringAttribute{
			Description: "Numeric identifier of the Constellation Network.",
			Computed:    computeFlag,
			Optional:    optionalFlag,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the network module",
			Computed:    computeFlag,
			Optional:    optionalFlag,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"href": schema.StringAttribute{
			Description: "href of the network module",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
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

func NWReachableModuleObjectType() (types.ObjectType) {
	return types.ObjectType{	
						AttrTypes: NWReachableModuleAttributeType(),
				}
}

func NWReachableModuleAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"id":    types.StringType,
		"href": types.StringType,
		"state": types.ObjectType {AttrTypes:NWModuleStateAttributeType()},
	}
}

func NWReachableModulesValue(data []interface{}) ([]attr.Value) {
	modules := []attr.Value{}
	for _, v := range data {
		module := v.(map[string]interface{})
		modules = append(modules, types.ObjectValueMust(
											NWReachableModuleAttributeType(),
													NWReachableModuleValue(module)))
	}
	return modules
}

func NWReachableModuleValue(module map[string]interface{}) (map[string]attr.Value) {
	id := types.StringNull()
	if module["id"] != nil {
		id = types.StringValue(module["id"].(string))
	}
	href := types.StringNull()
	if module["href"] != nil {
		href = types.StringValue(module["href"].(string))
	}
	state := types.ObjectNull(NWModuleStateAttributeType())
	if( module["state"]) != nil {
		state = types.ObjectValueMust(NWModuleStateAttributeType(), NWModuleStateAttributeValue(module["state"].(map[string]interface{})))
	}
	
	return map[string]attr.Value{
													"id": id,
													"href": href,
													"state": state,
												}
}
