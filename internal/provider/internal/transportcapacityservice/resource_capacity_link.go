package transportcapacity

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

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &TCCapacityLinkResource{}
	_ resource.ResourceWithConfigure   = &TCCapacityLinkResource{}
	_ resource.ResourceWithImportState = &TCCapacityLinkResource{}
)

// NewModuleResource is a helper function to simplify the provider implementation.
func NewTCCapacityLinkResource() resource.Resource {
	return &TCCapacityLinkResource{}
}

type TCCapacityLinkResource struct {
	client *ipm_pf.Client
}

type TCCapacityLinkResourceData struct {
	Id        types.String `tfsdk:"id"`
	Href      types.String `tfsdk:"href"`
	Config    types.Object   `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
}


// Metadata returns the data source type name.
func (r *TCCapacityLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_capacity_link"
}

// Schema defines the schema for the data source.
func (r *TCCapacityLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type TCCapacityLinkResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an Module",
		Attributes:  TCCapacityLinkSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *TCCapacityLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r TCCapacityLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TCCapacityLinkResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "NetworkResource: Create - ", map[string]interface{}{"NetworkResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	//r.create(&data, ctx, &resp.Diagnostics)
	resp.State.Set(ctx, &data)

}

func (r TCCapacityLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TCCapacityLinkResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "TCCapacityLinkResource: Read - ", map[string]interface{}{"TCCapacityLinkResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r TCCapacityLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TCCapacityLinkResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "TCCapacityLinkResource: Update", map[string]interface{}{"TCCapacityLinkResourceData": data})

	//.r.update(&data, ctx, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r TCCapacityLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TCCapacityLinkResourceData

	diags := req.State.Get(ctx, &data)
	diags.AddError(
		"Error Delete TCCapacityLink",
		"Delete: Could not delete TCCapacityLink. It is deleted together with its Transport Capacity",
	)
	resp.Diagnostics.Append(diags...)
}

func (r *TCCapacityLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TCCapacityLinkResource) read(state *TCCapacityLinkResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() || state.Id.IsNull() {
		diags.AddError(
			"TCCapacityLinkResource: Error Read Capacity Link",
			"Could not Read, Host Id or Capacity Link ID is not specified.",
		)
		return
	}
	tflog.Debug(ctx, "TCCapacityLinkResource: read ", map[string]interface{}{"state": state})

	body, err := r.client.ExecuteIPMHttpCommand("GET", "/capacity-links/"+ state.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"TCCapacityLinkResource: read ##: Error Update TCCapacityLinkResource",
			"Update:Could not read TCCapacityLinkResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "TCCapacityLinkResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TCCapacityLinkResource: read ##: Error Unmarshal response",
			"Update:Could not read TCCapacityLinkResource, unexpected error: "+err.Error(),
		)
		return
	}

	// populate state
	TCCapacityLinkData := data[0].(map[string]interface{})
	state.Populate(TCCapacityLinkData, ctx, diags)

	tflog.Debug(ctx, "TCCapacityLinkResource: read ## ", map[string]interface{}{"plan": state})
}

func (clData *TCCapacityLinkResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TCCapacityLinkResourceData: populate ## ")
	clData.Id = types.StringValue(data["id"].(string))
	clData.Href = types.StringValue(data["href"].(string))

	// populate Config
	if data["config"] != nil {
		clData.Config =types.ObjectValueMust( TCCapacityLinkConfigAttributeType(),TCCapacityLinkConfigAttributeValue(data["config"].(map[string]interface{})))
	}

	// populate state
	if data["state"] != nil {
		clData.State =types.ObjectValueMust(TCCapacityLinkStateAttributeType(),TCCapacityLinkStateAttributeValue(data["state"].(map[string]interface{})))
	}
	
	tflog.Debug(ctx, "TCCapacityLinkResourceData: populate SUCCESS ")
}

func TCCapacityLinkSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the port",
			Computed:    true,
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
		"config": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: TCCapacityLinkConfigAttributeType(),
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: TCCapacityLinkStateAttributeType(),
		},
	}
}

func TCCapacityLinkObjectType() (types.ObjectType) {
	return types.ObjectType{	
						AttrTypes: TCCapacityLinkAttributeType(),
				}
}

func TCCapacityLinkObjectsValue(data []interface{}) []attr.Value {
	tcCapacityLinks := []attr.Value{}
	for _, v := range data {
		tcCapacityLink := v.(map[string]interface{})
		if tcCapacityLink != nil {
			tcCapacityLinks = append(tcCapacityLinks, types.ObjectValueMust(
				TCCapacityLinkAttributeType(),
				TCCapacityLinkAttributeValue(tcCapacityLink)))
		}
	}
	return tcCapacityLinks
}

func TCCapacityLinkAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"id":    types.StringType,
		"href": types.StringType,
		"config": types.ObjectType {AttrTypes:TCCapacityLinkConfigAttributeType()},
		"state": types.ObjectType {AttrTypes:TCCapacityLinkStateAttributeType()},
	}
}


func TCCapacityLinkConfigAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"directionality":    types.StringType,
		"hub_module"  : types.ObjectType{AttrTypes: TCCapacityLinkConfigModuleAttributeType()},
		"leaf_module" : types.ObjectType{AttrTypes: TCCapacityLinkConfigModuleAttributeType()},
	}
}

func TCCapacityLinkConfigAttributeValue(capacityLinkState map[string]interface{}) (map[string]attr.Value) {
	directionality := types.StringNull()
	if capacityLinkState["directionality"] != nil {
		directionality = types.StringValue(capacityLinkState["directionality"].(string))
	}
	hubModule := types.ObjectNull(TCCapacityLinkConfigModuleAttributeType())
	if capacityLinkState["hubModule"] != nil {
		hubModule = types.ObjectValueMust(TCCapacityLinkConfigModuleAttributeType(), TCCapacityLinkConfigModuleAttributeValue(capacityLinkState["hubModule"].(map[string]interface{})))
	}
	leafModule := types.ObjectNull(TCCapacityLinkConfigModuleAttributeType())
	if capacityLinkState["leafModule"] != nil {
		leafModule = types.ObjectValueMust(TCCapacityLinkConfigModuleAttributeType(),TCCapacityLinkConfigModuleAttributeValue(capacityLinkState["leafModule"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"directionality": directionality,
		"hub_module" : hubModule,
		"leaf_module" : leafModule,
	}
}

func TCCapacityLinkConfigModuleAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"module_id":    types.StringType,
		"dscg_ctrl":    types.Int64Type,
		"dscg_shared":    types.BoolType,
		"tx_cdscs":    types.ListType{ElemType: types.Int64Type},
		"rx_cdscs":    types.ListType{ElemType: types.Int64Type},
		"idle_cdscs":  types.ListType{ElemType: types.Int64Type},
	}
}

func TCCapacityLinkConfigModuleAttributeValue(capacityLinkStateModule map[string]interface{}) (map[string]attr.Value) {
	moduleId := types.StringNull()
	if capacityLinkStateModule["moduleId"] != nil {
		moduleId = types.StringValue(capacityLinkStateModule["moduleId"].(string))
	}
	dscgCtrl := types.Int64Null()
	if capacityLinkStateModule["dscgCtrl"] != nil {
		dscgCtrl = types.Int64Value(int64(capacityLinkStateModule["dscgCtrl"].(float64)))
	}
	dscgShared := types.BoolNull()
	if capacityLinkStateModule["dscgShared"] != nil {
		dscgShared = types.BoolValue(capacityLinkStateModule["dscgShared"].(bool))
	}
	txCDSCs := types.ListNull(types.Int64Type)
	if capacityLinkStateModule["txCDSCs"] != nil {
		txCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(capacityLinkStateModule["txCDSCs"].([]interface{})))
	}
	rxCDSCs := types.ListNull(types.Int64Type)
	if capacityLinkStateModule["rxCDSCs"] != nil {
		rxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(capacityLinkStateModule["rxCDSCs"].([]interface{})))
	}
	idleCDSCs := types.ListNull(types.Int64Type)
	if capacityLinkStateModule["idleCDSCs"] != nil {
		idleCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(capacityLinkStateModule["idleCDSCs"].([]interface{})))
	}
	return map[string]attr.Value {
		"module_id":    moduleId,
		"dscg_ctrl":  dscgCtrl,
		"dscg_shared": dscgShared,
		"tx_cdscs":    txCDSCs,
		"rx_cdscs":    rxCDSCs,
		"idle_cdscs":  idleCDSCs,
	}
}

func TCCapacityLinkStateAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"directionality":    types.StringType,
		"life_cycle_state":    types.StringType,
		"life_cycle_state_cause" :types.ObjectType{AttrTypes: common.LifecycleStateCauseAttributeType()},
		"hub_module"  : types.ObjectType{AttrTypes: TCCapacityLinkStateModuleAttributeType()},
		"leaf_module" : types.ObjectType{AttrTypes: TCCapacityLinkStateModuleAttributeType()},
	}
}

func TCCapacityLinkStateAttributeValue(capacityLinkState map[string]interface{}) (map[string]attr.Value) {
	directionality := types.StringNull()
	if capacityLinkState["directionality"] != nil {
		directionality = types.StringValue(capacityLinkState["directionality"].(string))
	}
	lifecycleState := types.StringNull()
	if capacityLinkState["lifecycleState"] != nil {
		lifecycleState = types.StringValue(capacityLinkState["lifecycleState"].(string))
	}
	lifecycleStateCause := types.ObjectNull(common.LifecycleStateCauseAttributeType())
	if capacityLinkState["lifecycleStateCause"] != nil {
		lifecycleStateCause = types.ObjectValueMust(common.LifecycleStateCauseAttributeType(), common.LifecycleStateCauseAttributeValue(capacityLinkState["lifecycleStateCause"].(map[string]interface{})))
	}
	hubModule := types.ObjectNull(TCCapacityLinkStateModuleAttributeType())
	if capacityLinkState["hubModule"] != nil {
		hubModule = types.ObjectValueMust(TCCapacityLinkStateModuleAttributeType(), TCCapacityLinkStateModuleAttributeValue(capacityLinkState["hubModule"].(map[string]interface{})))
	}

	leafModule := types.ObjectNull(TCCapacityLinkStateModuleAttributeType())
	if capacityLinkState["leafModule"] != nil {
		leafModule = types.ObjectValueMust(TCCapacityLinkStateModuleAttributeType(),TCCapacityLinkStateModuleAttributeValue(capacityLinkState["leafModule"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"directionality": directionality,
		"life_cycle_state": lifecycleState,
		"life_cycle_state_cause": lifecycleStateCause,
		"hub_module" : hubModule,
		"leaf_module" : leafModule,
	}
}

func TCCapacityLinkStateModuleAttributeType() (map[string]attr.Type) {
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

func TCCapacityLinkStateModuleAttributeValue(capacityLinkStateModule map[string]interface{}) (map[string]attr.Value) {
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

func TCCapacityLinksAttributeValue(data []interface{}) []attr.Value {
	tcCapacityLinks := []attr.Value{}
	for _, v := range data {
		tcCapacityLink := v.(map[string]interface{})
		if tcCapacityLink != nil {
			tcCapacityLinks = append(tcCapacityLinks, types.ObjectValueMust(
				TCCapacityLinkAttributeType(),
				TCCapacityLinkAttributeValue(tcCapacityLink)))
		}
	}
	return tcCapacityLinks
}

func TCCapacityLinkAttributeValue(tcCapacityLink map[string]interface{}) map[string]attr.Value {
	id := types.StringNull()
	if tcCapacityLink["id"] != nil {
		id = types.StringValue(tcCapacityLink["id"].(string))
	}
	href := types.StringNull()
	if tcCapacityLink["href"] != nil {
		href = types.StringValue(tcCapacityLink["href"].(string))
	}
	config := types.ObjectNull(TCCapacityLinkConfigAttributeType())
	if (tcCapacityLink["config"]) != nil {
		config = types.ObjectValueMust(TCCapacityLinkConfigAttributeType(), TCCapacityLinkConfigAttributeValue(tcCapacityLink["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(TCCapacityLinkStateAttributeType())
	if (tcCapacityLink["state"]) != nil {
		state = types.ObjectValueMust(TCCapacityLinkStateAttributeType(), TCCapacityLinkStateAttributeValue(tcCapacityLink["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"id":     id,
		"href":   href,
		"config": config,
		"state":  state,
	}
}