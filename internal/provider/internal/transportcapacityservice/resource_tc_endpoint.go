package transportcapacity

import (
	"context"
	"encoding/json"

	"terraform-provider-ipm/internal/ipm_pf"
	"terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &TCEndpointResource{}
	_ resource.ResourceWithConfigure   = &TCEndpointResource{}
	_ resource.ResourceWithImportState = &TCEndpointResource{}
)

// NewTCEndpointResource is a helper function to simplify the provider implementation.
func NewTCEndpointResource() resource.Resource {
	return &TCEndpointResource{}
}

type TCEndpointResource struct {
	client *ipm_pf.Client
}


type TCEndpointConfig struct {
	Capacity    types.Int64 `tfsdk:"capacity"`
	Selector    common.IfSelector `tfsdk:"selector"`
}

type TCEndpointResourceData struct {
	TCId      types.String `tfsdk:"tc_id"`
	Id        types.String `tfsdk:"id"`
	Href      types.String `tfsdk:"href"`
	Config    TCEndpointConfig   `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *TCEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_TCEndpoint"
}

// Schema defines the schema for the data source.
func (r *TCEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type TCEndpointResourceData struct 
	resp.Schema = schema.Schema{
		Description: "Manages an TCEndpoint",
		Attributes: TCEndpointSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *TCEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r TCEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TCEndpointResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "TCEndpointResource: Create - ", map[string]interface{}{"TCEndpointResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r TCEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TCEndpointResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "TCEndpointResource: Read - ", map[string]interface{}{"TCEndpointResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r TCEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TCEndpointResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"TCEndpointResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r TCEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TCEndpointResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"TCEndpointResourceData": data})

	resp.Diagnostics.Append(diags...)

	resp.State.RemoveResource(ctx)
}

func (r *TCEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (plan *TCEndpointResourceData) Update(client *ipm_pf.Client, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TCEndpointResourceData: update - plan", map[string]interface{}{"plan": plan})
	if plan.TCId.IsNull() || plan.Id.IsNull(){
		diags.AddError(
			"TCEndpointResourceData: Error Update TCEndpoint",
			"Update: Could not Update TCEndpoint. Host Id or Port Id is not specified",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get Network config settings
	if !plan.Config.Capacity.IsNull() {
		updateRequest["capacity"] = plan.Config.Capacity.ValueInt64()
	}

	// send Update request to server
	rb, err := json.Marshal(updateRequest)
	if err != nil {
		diags.AddError(
			"TCEndpointResourceData: Update ##: Error Update TCEndpoint",
			"Update: Could not Marshal TCEndpointResourceData, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "TCEndpointResourceData: update - rb", map[string]interface{}{"rb": rb})

	body, err := client.ExecuteIPMHttpCommand("PUT", "/transport-capacities/"+ plan.TCId.ValueString()+"/endpoints/"+plan.Id.ValueString(), rb)

	if err != nil {
		diags.AddError(
			"TCEndpointResourceData: Update ##: Error Update TCEndpointResourceData",
			"Update:Could not Update TCEndpointResourceData, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "TCEndpointResourceData: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TCEndpointResourceData: Update ##: Error Unmarshal TCEndpointResourceData",
			"Update:Could not Update TCEndpointResourceData, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "TCEndpointResourceData: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": data, "plan": plan})

}

func (r *TCEndpointResource) update(plan *TCEndpointResourceData, ctx context.Context, diags *diag.Diagnostics) {

	plan.Update(r.client, ctx, diags)

	if !diags.HasError() {
		r.read(plan, ctx, diags)
	}

	tflog.Debug(ctx, "TCEndpointResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *TCEndpointResource) read(state *TCEndpointResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.TCId.IsNull() || state.Id.IsNull() {
		diags.AddError(
			"TCEndpointResource: Error missing identifier",
			"Could not Read, Host Id or TCEndpoint ID is not specified.",
		)
		return
	}
	tflog.Debug(ctx, "TCEndpointResource: read ", map[string]interface{}{"state": state})

	body, err := r.client.ExecuteIPMHttpCommand("GET", "/transport-capacities/"+ state.TCId.ValueString()+"/endpoints/"+state.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"TCEndpointResource: read ##: Error Update TCEndpointResource",
			"Update:Could not read TCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "TCEndpointResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TCEndpointResource: read ##: Error Unmarshal response",
			"Update:Could not read TCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}

	// populate state
	TCEndpointData := data[0].(map[string]interface{})
	state.Populate(TCEndpointData, ctx, diags)

	tflog.Debug(ctx, "TCEndpointResource: read ## ", map[string]interface{}{"plan": state})
}

func (tcData *TCEndpointResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	tflog.Debug(ctx, "TCEndpointResourceData:## ")
	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}

	tflog.Debug(ctx, "TCEndpointResourceData: populate ## ", map[string]interface{}{"computeFlag": computeFlag, "data": data})
	if computeFlag {
		tcData.TCId = types.StringValue(data["parentId"].(string))
		tcData.Id = types.StringValue(data["id"].(string))
	}
	tcData.Href = types.StringValue(data["href"].(string))

	tflog.Debug(ctx, "TCEndpointResourceData: populate Config## ")
	// populate Config
	if data["config"] != nil {
		TCEndpointConfig := data["config"].(map[string]interface{})
		for k, v := range TCEndpointConfig {
			switch k {
			case "capacity": 
				if !tcData.Config.Capacity.IsNull() || computeFlag {
					tcData.Config.Capacity = types.Int64Value(int64(v.(float64)))
				}
			case "selector":
				common.IfSelectorPopulate(v.(map[string]interface{}), &tcData.Config.Selector, computeFlag)
			}
		}
	}
	tflog.Debug(ctx, "TCEndpointResourceData: populate State## ")
	// populate state
	if data["state"] != nil {
		tcData.State =types.ObjectValueMust(
			TCEndpointStateAttributeType(),TCEndpointStateAttributeValue(data["state"].(map[string]interface{})))
	}
	
	tflog.Debug(ctx, "TCEndpointResourceData: populate SUCCESS ")
}

func TCEndpointSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"tc_id": schema.StringAttribute{
			Description: "Numeric identifier of the Transport Capacity.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the port",
			Optional:    true,
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
		//Config    NodeConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"capacity": schema.StringAttribute{
					Description: "Capacity",
					Optional:    true,
				},
				"selector": common.IfSelectorSchema(),
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: TCEndpointStateAttributeType(),
		},
	}
}

func TCEndpointObjectType() (types.ObjectType) {
	return types.ObjectType{	
						AttrTypes: TCEndpointAttributeType(),
				}
}

func TCEndpointObjectsValue(data []interface{}) []attr.Value {
	tcEnpoints := []attr.Value{}
	for _, v := range data {
		tcEnpoint := v.(map[string]interface{})
		if tcEnpoint != nil {
			tcEnpoints = append(tcEnpoints, types.ObjectValueMust(
				TCEndpointAttributeType(),
				TCEndpointAttributeValue(tcEnpoint)))
		}
	}
	return tcEnpoints
}

func TCEndpointAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"tc_id":  types.StringType,
		"id":     types.StringType,
		"href":   types.StringType,
		"config": types.ObjectType {AttrTypes:TCEndpointConfigAttributeType()},
		"state":  types.ObjectType {AttrTypes:TCEndpointStateAttributeType()},
	}
}

func TCEndpointAttributeValue(tcEndpoint map[string]interface{}) map[string]attr.Value {
	tc_id := types.StringNull()
	if tcEndpoint["parentId"] != nil {
		tc_id = types.StringValue(tcEndpoint["parentId"].(string))
	}
	id := types.StringNull()
	if tcEndpoint["id"] != nil {
		id = types.StringValue(tcEndpoint["id"].(string))
	}
	href := types.StringNull()
	if tcEndpoint["href"] != nil {
		href = types.StringValue(tcEndpoint["href"].(string))
	}
	config := types.ObjectNull(TCEndpointConfigAttributeType())
	if (tcEndpoint["config"]) != nil {
		config = types.ObjectValueMust(TCEndpointConfigAttributeType(), TCEndpointConfigAttributeValue(tcEndpoint["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(TCEndpointStateAttributeType())
	if (tcEndpoint["state"]) != nil {
		state = types.ObjectValueMust(TCEndpointStateAttributeType(), TCEndpointStateAttributeValue(tcEndpoint["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"tc_id" : tc_id,
		"id":     id,
		"href":   href,
		"config": config,
		"state":  state,
	}
}


func TCEndpointConfigAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"capacity":    types.Int64Type,
		"selector"  :  types.ObjectType{AttrTypes: common.IfSelectorAttributeType()},
	}
}

func TCEndpointConfigAttributeValue(tcEndpointConfig map[string]interface{}) (map[string]attr.Value) {
	capacity := types.Int64Null()
	if tcEndpointConfig["capacity"] != nil {
		capacity = types.Int64Value(int64(tcEndpointConfig["capacity"].(float64)))
	}
	selector := types.ObjectNull(common.IfSelectorAttributeType())
	if tcEndpointConfig["tcEndpointConfig"] != nil {
		selector = types.ObjectValueMust(common.IfSelectorAttributeType(), common.IfSelectorAttributeValue(tcEndpointConfig["tcEndpointConfig"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"capacity": capacity,
		"selector" : selector,
	}
}

func TCEndpointStateAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"capacity":    types.Int64Type,
		"life_cycle_state":    types.StringType,
		"life_cycle_state_cause" :types.ObjectType{AttrTypes: common.LifecycleStateCauseAttributeType()},
		"host_port"  : types.ObjectType{AttrTypes: common.EndpointHostPortAttributeType()},
		"module_if" : types.ObjectType{AttrTypes: common.ModuleIfAttributeType()},
	}
}

func TCEndpointStateAttributeValue(capacityLinkState map[string]interface{}) (map[string]attr.Value) {
	capacity := types.Int64Null()
	if capacityLinkState["capacity"] != nil {
		capacity = types.Int64Value(int64(capacityLinkState["capacity"].(float64)))
	}
	lifecycleState := types.StringNull()
	if capacityLinkState["lifecycleState"] != nil {
		lifecycleState = types.StringValue(capacityLinkState["lifecycleState"].(string))
	}
	lifecycleStateCause := types.ObjectNull(common.LifecycleStateCauseAttributeType())
	if capacityLinkState["lifecycleStateCause"] != nil {
		lifecycleStateCause = types.ObjectValueMust(common.LifecycleStateCauseAttributeType(), common.LifecycleStateCauseAttributeValue(capacityLinkState["lifecycleStateCause"].(map[string]interface{})))
	}
	hostPort := types.ObjectNull(common.EndpointHostPortAttributeType())
	if capacityLinkState["hostPort"] != nil {
		hostPort = types.ObjectValueMust(common.EndpointHostPortAttributeType(), common.EndpointHostPortAttributeValue(capacityLinkState["hostPort"].(map[string]interface{})))
	}

	moduleIf := types.ObjectNull(common.ModuleIfAttributeType())
	if capacityLinkState["moduleIf"] != nil {
		moduleIf = types.ObjectValueMust(common.ModuleIfAttributeType(),common.ModuleIfAttributeValue(capacityLinkState["moduleIf"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"capacity": capacity,
		"life_cycle_state": lifecycleState,
		"life_cycle_state_cause": lifecycleStateCause,
		"module_if" : moduleIf,
		"host_port" : hostPort,
	}
}

func TCEndpointStateModuleAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"module_id":    types.StringType,
		"module_name":    types.StringType,
		"mac_address":    types.StringType,
		"dscg_id":    types.StringType,
		"dscg_aid":    types.StringType,
		"dscg_ctrl":    types.Int64Type,
		"dscg_shared":    types.BoolType,
		"life_cycle_state":    types.StringType,
		"life_cycle_state_cause" :types.ObjectType{AttrTypes: common.LifecycleStateCauseAttributeType()},
		"tx_cdscs":    types.ListType{ElemType: types.Int64Type},
		"rx_cdscs":    types.ListType{ElemType: types.Int64Type},
		"idle_cdscs":  types.ListType{ElemType: types.Int64Type},
	}
}

func TCEndpointStateModuleAttributeValue(capacityLinkStateModule map[string]interface{}) (map[string]attr.Value) {
	moduleId := types.StringNull()
	if capacityLinkStateModule["moduleId"] != nil {
		moduleId = types.StringValue(capacityLinkStateModule["moduleId"].(string))
	}
	moduleName := types.StringNull()
	if capacityLinkStateModule["moduleName"] != nil {
		moduleId = types.StringValue(capacityLinkStateModule["moduleId"].(string))
	}
	macAddress := types.StringNull()
	if capacityLinkStateModule["macAddress"] != nil {
		macAddress = types.StringValue(capacityLinkStateModule["macAddress"].(string))
	}
	dscgId := types.StringNull()
	if capacityLinkStateModule["dscgId"] != nil {
		dscgId = types.StringValue(capacityLinkStateModule["dscgId"].(string))
	}
	dscgCtrl := types.Int64Null()
	if capacityLinkStateModule["dscgCtrl"] != nil {
		dscgCtrl = types.Int64Value(int64(capacityLinkStateModule["dscgCtrl"].(float64)))
	}
	dscgAid := types.StringNull()
	if capacityLinkStateModule["dscgAid"] != nil {
		dscgAid = types.StringValue(capacityLinkStateModule["dscgAid"].(string))
	}
	dscgShared := types.BoolNull()
	if capacityLinkStateModule["dscgShared"] != nil {
		dscgShared = types.BoolValue(capacityLinkStateModule["dscgShared"].(bool))
	}
	lifecycleState := types.StringNull()
	if capacityLinkStateModule["lifecycleState"] != nil {
		lifecycleState = types.StringValue(capacityLinkStateModule["lifecycleState"].(string))
	}
	lifecycleStateCause := types.ObjectNull(common.LifecycleStateCauseAttributeType())
	if capacityLinkStateModule["lifecycleStateCause"] != nil {
		lifecycleStateCause =  types.ObjectValueMust( common.LifecycleStateCauseAttributeType(),
		common.LifecycleStateCauseAttributeValue(capacityLinkStateModule["lifecycleStateCause"].(map[string]interface{})))
	}
	txCDSCs := types.ListNull(types.Int64Type)
	if capacityLinkStateModule["txCDSCs"] != nil {
		txCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(capacityLinkStateModule["txCDSCs"].([]interface{})))
	}
	rxCDSCs := types.ListNull(types.Int64Type)
	if capacityLinkStateModule["rxCDSCs"] != nil {
		txCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(capacityLinkStateModule["rxCDSCs"].([]interface{})))
	}
	idleCDSCs := types.ListNull(types.Int64Type)
	if capacityLinkStateModule["idleCDSCs"] != nil {
		idleCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(capacityLinkStateModule["idleCDSCs"].([]interface{})))
	}
	return map[string]attr.Value {
		"module_id":    moduleId,
		"module_name":  moduleName,
		"mac_address":  macAddress,
		"dscg_id":    dscgId,
		"dscg_aid":   dscgAid,
		"dscg_ctrl":  dscgCtrl,
		"dscg_shared": dscgShared,
		"life_cycle_state": lifecycleState,
		"life_cycle_state_cause" : lifecycleStateCause,
		"tx_cdscs":    txCDSCs,
		"rx_cdscs":    rxCDSCs,
		"idle_cdscs":  idleCDSCs,
	}
}
