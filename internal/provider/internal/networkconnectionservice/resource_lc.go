package networkconnection

import (
	"context"
	"encoding/json"

	"terraform-provider-ipm/internal/ipm_pf"

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
	_ resource.Resource                = &LCResource{}
	_ resource.ResourceWithConfigure   = &LCResource{}
	_ resource.ResourceWithImportState = &LCResource{}
)

// NewLCResource is a helper function to simplify the provider implementation.
func NewLCResource() resource.Resource {
	return &LCResource{}
}

type LCResource struct {
	client *ipm_pf.Client
}

type LCResourceData struct {
	Id       types.String `tfsdk:"id"`
	Href     types.String `tfsdk:"href"`
	Config   types.Object `tfsdk:"config"`
	State    types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *LCResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nc_lc"
}

// Schema defines the schema for the data source.
func (r *LCResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type LCResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages a LC",
		Attributes:  map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "LC ID",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"href": schema.StringAttribute{
				Description: "href of the LC",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			//Config     types.Object   `tfsdk:"config"`
			"config": schema.ObjectAttribute{
				Computed: true,
				AttributeTypes: LCConfigAttributeType(),
			},
			//State     types.Object   `tfsdk:"state"`
			"state": schema.ObjectAttribute{
				Computed: true,
				AttributeTypes: LCStateAttributeType(),
			},
		},
	}
}
// Configure adds the provider configured client to the data source.
func (r *LCResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r LCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LCResourceData

	diags := req.Config.Get(ctx, &data)
	tflog.Debug(ctx, "NetworkModuleResource: Create - ", map[string]interface{}{"NetworkModuleResourceData": data})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

}

func (r LCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LCResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LCResource: Create - ", map[string]interface{}{"LCResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LCResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"LCResourceData": data})

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r LCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LCResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LCResource: Delete", map[string]interface{}{"LCResource": data})

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *LCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *LCResource) read(state *LCResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() {
		diags.AddError(
			"Error Update Module",
			"Update: Could not Update Module. Id and Href are not specified.",
		)
		return
	}
	var body []byte
	var err error
	if !state.Id.IsNull() {
		body, err = r.client.ExecuteIPMHttpCommand("GET", "/lcs/"+state.Id.ValueString()+"?content=expanded", nil)
	} else {
		body, err = r.client.ExecuteIPMHttpCommand("GET", state.Href.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"LCResource: read ##: Error Read LCResource",
			"Update:Could not read LCResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "LCResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"LCResource: read ##: Error Unmarshal response",
			"Update:Could not read LCResource, unexpected error: "+err.Error(),
		)
		return
	}
	state.Populate(data, ctx, diags)

	tflog.Debug(ctx, "LCResource: read ## ", map[string]interface{}{"plan": state})

}

func (lcData *LCResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}

	tflog.Debug(ctx, "LCResourceData: populate ## ", map[string]interface{}{"data": data})
	if computeFlag {
		lcData.Id = types.StringValue(data["id"].(string))
	}
	lcData.Href = types.StringValue(data["href"].(string))

	//populate state
	if data["state"] != nil {
		lcData.State = types.ObjectValueMust(
			LCStateAttributeType(), LCStateAttributeValue(data["state"].(map[string]interface{})))
	}
		//populate config
	if data["config"] != nil {
		lcData.Config = types.ObjectValueMust(
			LCConfigAttributeType(), LCConfigAttributeValue(data["config"].(map[string]interface{})))
	}
	tflog.Debug(ctx, "LCResourceData: populate ## ", map[string]interface{}{"plan": lcData})

}


func LCObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: LCAttributeType(),
	}
}

func LCObjectsValue(data []interface{}) []attr.Value {
	LCs := []attr.Value{}
	for _, v := range data {
		LC := v.(map[string]interface{})
		if LC != nil {
			LCs = append(LCs, types.ObjectValueMust(
				LCAttributeType(),
				LCAttributeValue(LC)))
		}
	}
	return LCs
}

func LCAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":     types.StringType,
		"href":   types.StringType,
		"config": types.ObjectType{AttrTypes: LCConfigAttributeType()},
		"state":  types.ObjectType{AttrTypes: LCStateAttributeType()},
	}
}

func LCAttributeValue(LC map[string]interface{}) map[string]attr.Value {
	id := types.StringNull()
	if LC["id"] != nil {
		id = types.StringValue(LC["id"].(string))
	}
	href := types.StringNull()
	if LC["href"] != nil {
		href = types.StringValue(LC["href"].(string))
	}
	config := types.ObjectNull(LCConfigAttributeType())
	if (LC["config"]) != nil {
		config = types.ObjectValueMust(LCConfigAttributeType(), LCConfigAttributeValue(LC["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(LCStateAttributeType())
	if (LC["state"]) != nil {
		state = types.ObjectValueMust(LCStateAttributeType(), LCStateAttributeValue(LC["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"id":     id,
		"href":   href,
		"config": config,
		"state":  state,
	}
}

func LCStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"col_id":        types.Int64Type,
		"lc_aid":      types.StringType,
		"direction":           types.StringType,
		"lc_ctrl":       types.Int64Type,
		"module_id": types.StringType,
		"client_aid":           types.StringType,
		"dscg_aid": types.StringType,
		"mac_address":       types.StringType,
		"line_aid" : types.StringType,
		"remote_module_id": types.StringType,
		"remote_client_id": types.StringType,
	}
}


func LCStateAttributeValue(state map[string]interface{}) map[string]attr.Value {
	lcAid := types.StringNull()
	if state["lcAid"] != nil {
		lcAid = types.StringValue(state["lcAid"].(string))
	}
	colId := types.Int64Null()
	if state["colId"] != nil {
		colId = types.Int64Value(int64(state["colId"].(float64)))
	}
	direction := types.StringNull()
	if state["direction"] != nil {
		direction = types.StringValue(state["direction"].(string))
	}
	moduleId := types.StringNull()
	if state["moduleId"] != nil {
		moduleId = types.StringValue(state["moduleId"].(string))
	}
	clientAid := types.StringNull()
	if state["clientAid"] != nil {
		clientAid = types.StringValue(state["clientAid"].(string))
	}
	dscgAid := types.StringNull()
	if state["dscgAid"] != nil {
		dscgAid = types.StringValue(state["dscgAid"].(string))
	}
	macAddress := types.StringNull()
	if state["macAddress"] != nil {
		macAddress = types.StringValue(state["macAddress"].(string))
	}
	lineAid := types.StringNull()
	if state["lineAid"] != nil {
		lineAid = types.StringValue(state["lineAid"].(string))
	}
	remoteModuleId := types.StringNull()
	if state["remoteModuleId"] != nil {
		remoteModuleId = types.StringValue(state["remoteModuleId"].(string))
	}
	remoteClientId := types.StringNull()
	if state["remoteClientId"] != nil {
		remoteClientId = types.StringValue(state["remoteClientId"].(string))
	}
	lcCtrl := types.Int64Null()
	if state["lcCtrl"] != nil {
		lcCtrl = types.Int64Value(int64(state["lcCtrl"].(float64)))
	}
	return map[string]attr.Value {
		"col_id":      colId,
		"lc_aid":      lcAid,
		"direction":   direction,
		"lc_ctrl":     lcCtrl,
		"module_id":   moduleId,
		"client_aid":  clientAid,
		"dscg_aid":    dscgAid,
		"mac_address": macAddress,
		"line_aid" :   lineAid,
		"remote_module_id": remoteModuleId,
		"remote_client_id": remoteClientId,
	}
}

func LCConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"direction":     types.StringType,
		"lc_ctrl":       types.Int64Type,
		"module_id":     types.StringType,
		"client_aid":    types.StringType,
		"dscg_aid":      types.StringType,
	}
}

func LCConfigAttributeValue(config map[string]interface{}) map[string]attr.Value {
	
	lcCtrl := types.Int64Null()
	if config["lcCtrl"] != nil {
		lcCtrl = types.Int64Value(int64(config["lcCtrl"].(float64)))
	}
	clientAid := types.StringNull()
	if config["clientAid"] != nil {
		clientAid = types.StringValue(config["clientAid"].(string))
	}
	dscgAid := types.StringNull()
	if config["dscgAid"] != nil {
		dscgAid = types.StringValue(config["dscgAid"].(string))
	}
	direction := types.StringNull()
	if config["direction"] != nil {
		direction = types.StringValue(config["direction"].(string))
	}
	moduleId := types.StringNull()
	if config["moduleId"] != nil {
		moduleId = types.StringValue(config["moduleId"].(string))
	}
	
	return map[string]attr.Value{
		"direction":     direction,
		"lc_ctrl":       lcCtrl,
		"module_id":     moduleId,
		"client_aid":    clientAid,
		"dscg_aid":      dscgAid,
	}
}
