package nduservice

import (
	"context"
	"encoding/json"
	"strings"

	"terraform-provider-ipm/internal/ipm_pf"
	"terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &XRResource{}
	_ resource.ResourceWithConfigure   = &XRResource{}
	_ resource.ResourceWithImportState = &XRResource{}
)

// NewXRResource is a helper function to simplify the provider implementation.
func NewXRResource() resource.Resource {
	return &XRResource{}
}

type XRResource struct {
	client *ipm_pf.Client
}

type XRConfig struct {
	Name            types.String `tfsdk:"name"`
	RequiredType    types.String `tfsdk:"required_type"`
	RequiredSubType types.String `tfsdk:"required_sub_type"`
}

type XRResourceData struct {
	Id         types.String  `tfsdk:"id"`
	ParentId   types.String   `tfsdk:"parent_id"`
	Href       types.String   `tfsdk:"href"`
	ColId      types.Int64    `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config    *XRConfig    `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *XRResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_xr"
}

// Schema defines the schema for the data source.
func (r *XRResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type XRResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages NDU XR",
		Attributes:  XRResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *XRResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r XRResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data XRResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "XRResource: Create - ", map[string]interface{}{"XRResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.create(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(diags...)
}

func (r XRResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data XRResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "XRResource: Read - ", map[string]interface{}{"XRResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r XRResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data XRResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "XRResource: Update", map[string]interface{}{"XRResourceData": data})

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

func (r XRResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data XRResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "XRResource: Update", map[string]interface{}{"XRResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *XRResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *XRResource) create(plan *XRResourceData, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "XRResource: create ##", map[string]interface{}{"plan": plan})

	if plan.Identifier.DeviceId.IsNull() && plan.Identifier.ParentColId.IsNull() &&  plan.Identifier.ColId.IsNull() {
		diags.AddError(
			"Error Create PortResource",
			"Create: Could not create PortResource. Device ID and resource parent col ID and its Col Id is not specified",
		)
		return
	}

	// get Network config settings
	var configRequest = make(map[string]interface{})
	if !plan.Config.Name.IsNull() {
		configRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.RequiredSubType.IsNull() {
		configRequest["requiredSubType"] = plan.Config.RequiredSubType.ValueString()
	}
	if !plan.Config.RequiredType.IsNull() {
		configRequest["requiredType"] = plan.Config.RequiredType.ValueString()
	}

	tflog.Debug(ctx, "PortResource: create ## ", map[string]interface{}{"Create Request": configRequest})

	// send create request to server
	rb, err := json.Marshal(configRequest)
	if err != nil {
		diags.AddError(
			"PortResource: create ##: Error Create AC",
			"Create: Could not Marshal PortResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" +  plan.Identifier.ParentColId.ValueString() + "/xr/" + plan.Identifier.ColId.ValueString(), rb)
	if err != nil {
		if !strings.Contains(err.Error(), "status: 202") {
			diags.AddError(
				"PortResource: create ##: Error create PortResource",
				"Create:Could not create PortResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "PortResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"PortResource: Create ##: Error Unmarshal response",
			"Update:Could not Create PortResource, unexpected error: "+err.Error(),
		)
		return
	}

	result := data[0].(map[string]interface{})

	href := result["href"].(string)
	splits := strings.Split(href, "/")
	id := splits[len(splits)-1]
	plan.Href = types.StringValue(href)
	plan.Id = types.StringValue(id)

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "PortResource: create failed. Can't find the created VOA")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "PortResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *XRResource) update(plan *XRResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "XRResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"XRResource: Error update XR",
			"XRResource: Could not update XR. Href, NDUId, PortColId, LinePTPColId, XR ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})
	// get TC config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.RequiredType.IsNull() {
		updateRequest["requiredType"] = plan.Config.RequiredType.ValueString()
	}
	if !plan.Config.RequiredSubType.IsNull() {
		updateRequest["requiredSubType"] = plan.Config.RequiredSubType.ValueString()
	}

	tflog.Debug(ctx, "XRResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"XRResource: update ##: Error Create AC",
				"Create: Could not Marshal XRResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ParentColId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" +  plan.Identifier.ParentColId.ValueString()  + "/xr/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"XRResource: update ##: Error update XRResource",
				"Update: Could not update XRResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "XRResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"XRResource: Create ##: Error Unmarshal response",
				"Update:Could not Create XRResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "XRResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "XRResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *XRResource) delete(state *XRResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Href.IsNull() && state.Id.IsNull() {
		diags.AddError(
			"Error Delete XRResource",
			"Read: Could not delete. NC ID is not specified",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", state.Href.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"XRResource: delete ##: Error Delete XRResource",
			"Update:Could not delete XRResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "XRResource: delete ## ", map[string]interface{}{"state": state})
}

func (r *XRResource) read(state *XRResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read XRResource",
			"XRResource: Could not read. ParentId and Id and Href are not specified.",
		)
		return
	}

	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/xr" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/xr" + queryStr + "&q={\"state.tomAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/xr" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/xr/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/xr" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"XRResource: read ##: Error Read XRResource",
			"Read:Could not get XRResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "XRResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"XRResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal XRResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
		case []interface{}:
			if len(resp.([]interface{})) > 0 {
				state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
			} else {
				diags.AddError(
					"XRResource: read ##: Can not get Module",
					"Read:Could not get ODU for query: "+queryStr,
				)
				return
			}
		case map[string]interface{}:
			state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "XRResource: read ## ", map[string]interface{}{"plan": state})
}

func (xrData *XRResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "XRResourceData: populate ## ", map[string]interface{}{"plan": data})

	xrData.Id = types.StringValue(data["id"].(string))
	xrData.ColId = types.Int64Value(int64(data["colid"].(float64)))
	xrData.Href = types.StringValue(data["href"].(string))
	xrData.ParentId = types.StringValue(data["parentId"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if xrData.Config == nil {
		xrData.Config = &XRConfig{}
	}
	if config != nil {
		if xrData.Config == nil {
			xrData.Config = &XRConfig{}
		}
		for k, v := range config {
			switch k {
			case "name":
				if !xrData.Config.Name.IsNull() {
					xrData.Config.Name = types.StringValue(v.(string))
				}
			case "requiredType":
				if !xrData.Config.RequiredType.IsNull() {
					xrData.Config.RequiredType = types.StringValue(v.(string))
				}
			case "requiredSubType":
				if !xrData.Config.RequiredSubType.IsNull() {
					xrData.Config.RequiredSubType = types.StringValue(v.(string))
				}
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		xrData.State = types.ObjectValueMust(XRStateAttributeType(), XRStateAttributeValue(state))
	}

	tflog.Debug(ctx, "XRResourceData: read ## ", map[string]interface{}{"xrData": xrData})
}

func XRResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Carrier.",
			Computed:    true,
		},
		"parent_id": schema.StringAttribute{
			Description: "parent id",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		"col_id": schema.Int64Attribute{
			Description: "col id",
			Computed:    true,
		},
		"identifier": common.ResourceIdentifierAttribute(),
		"config": schema.SingleNestedAttribute{
			Description: "Network Connection LC Config",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "term_lb",
					Optional:    true,
				},
				"required_type": schema.StringAttribute{
					Description: "required_type",
					Optional:    true,
				},
				"required_sub_type": schema.StringAttribute{
					Description: "required_sub_type",
					Optional:    true,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: XRStateAttributeType(),
		},
	}
}

func XRObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: XRAttributeType(),
	}
}

func XRObjectsValue(data []interface{}) []attr.Value {
	xrs := []attr.Value{}
	for _, v := range data {
		xr := v.(map[string]interface{})
		if xr != nil {
			xrs = append(xrs, types.ObjectValueMust(
				XRAttributeType(),
				XRAttributeValue(xr)))
		}
	}
	return xrs
}

func XRAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":          types.StringType,
		"href":        types.StringType,
		"col_id":      types.Int64Type,
		"config":      types.ObjectType{AttrTypes: XRConfigAttributeType()},
		"state":       types.ObjectType{AttrTypes: XRStateAttributeType()},
	}
}

func XRAttributeValue(xr map[string]interface{}) map[string]attr.Value {
	colId := types.Int64Null()
	if xr["colId"] != nil {
		colId = types.Int64Value(int64(xr["colId"].(float64)))
	}
	href := types.StringNull()
	if xr["href"] != nil {
		href = types.StringValue(xr["href"].(string))
	}
	id := types.StringNull()
	if xr["id"] != nil {
		id = types.StringValue(xr["id"].(string))
	}
	parentId := types.StringNull()
	if xr["parentId"] != nil {
		parentId = types.StringValue(xr["parentId"].(string))
	}
	config := types.ObjectNull(XRConfigAttributeType())
	if (xr["config"]) != nil {
		config = types.ObjectValueMust(XRConfigAttributeType(), XRConfigAttributeValue(xr["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(XRStateAttributeType())
	if (xr["state"]) != nil {
		state = types.ObjectValueMust(XRStateAttributeType(), XRStateAttributeValue(xr["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"col_id":      colId,
		"parent_id": parentId,
		"id":          id,
		"href":        href,
		"config":      config,
		"state":       state,
	}
}
func XRConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"Name":              types.StringType,
		"required_type":     types.StringType,
		"required_sub_type": types.StringType,
	}
}

func XRConfigAttributeValue(xrConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	requiredType := types.StringNull()
	requiredSubType := types.StringNull()
	phyMode := types.StringNull()

	if (xrConfig["name"]) != nil {
		name = types.StringValue(xrConfig["name"].(string))
	}
	if (xrConfig["requiredType"]) != nil {
		requiredType = types.StringValue(xrConfig["requiredType"].(string))
	}
	if (xrConfig["requiredSubType"]) != nil {
		requiredSubType = types.StringValue(xrConfig["requiredSubType"].(string))
	}

	return map[string]attr.Value{
		"name":              name,
		"required_type":     requiredType,
		"required_sub_type": requiredSubType,
		"phy_mode":          phyMode,
	}
}

func XRStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":           types.StringType,
		"xr_aid":               types.StringType,
		"required_type":        types.StringType,
		"required_sub_type":    types.StringType,
		"name":                 types.StringType,
		"restart_action":       types.StringType,
		"factory_reset_action": types.BoolType,
		"module":               types.ObjectType{AttrTypes: ModuleAttributeType()},
		"inventory":            types.ObjectType{AttrTypes: common.InventoryAttributeType()},
		"lifecycle_state":      types.StringType,
	}
}

func XRStateAttributeValue(xrState map[string]interface{}) map[string]attr.Value {
	xrAid := types.StringNull()
	parentAid := types.StringNull()
	name := types.StringNull()
	requiredType := types.StringNull()
	requiredSubType := types.StringNull()
	restartAction := types.StringNull()
	factoryResetAction := types.BoolNull()
	module := types.ObjectNull(ModuleAttributeType())
	inventory := types.ObjectNull(common.InventoryAttributeType())
	lifecycleState := types.StringNull()

	for k, v := range xrState {
		switch k {
		case "xrAid":
			xrAid = types.StringValue(v.(string))
		case "parentAid":
			parentAids := v.([]interface{})
			parentAid = types.StringValue(parentAids[0].(string))
		case "name":
			name = types.StringValue(v.(string))
		case "requiredType":
			requiredType = types.StringValue(v.(string))
		case "requiredSubType":
			requiredSubType = types.StringValue(v.(string))
		case "restartAction":
			restartAction = types.StringValue(v.(string))
		case "factoryResetAction":
			factoryResetAction = types.BoolValue(v.(bool))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		case "inventory":
			inventory = types.ObjectValueMust(common.InventoryAttributeType(), common.InventoryAttributeValue(v.(map[string]interface{})))
		case "module":
			module = types.ObjectValueMust(ModuleAttributeType(), ModuleAttributeValue(v.(map[string]interface{})))
		}
	}

	return map[string]attr.Value{
		"xr_aid":               xrAid,
		"parent_aid":          parentAid,
		"name":                 name,
		"required_type":        requiredType,
		"required_sub_type":    requiredSubType,
		"restart_action":       restartAction,
		"factory_reset_action": factoryResetAction,
		"module":               module,
		"lifecycle_state":      lifecycleState,
		"inventory":            inventory,
	}
}
func ModuleObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: ModuleAttributeType(),
	}
}

func ModuleObjectsValue(data []interface{}) []attr.Value {
	xrs := []attr.Value{}
	for _, v := range data {
		xr := v.(map[string]interface{})
		if xr != nil {
			xrs = append(xrs, types.ObjectValueMust(
				ModuleAttributeType(),
				ModuleAttributeValue(xr)))
		}
	}
	return xrs
}

func ModuleAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_id":          types.StringType,
		"module_name":        types.StringType,
		"connectivity_state": types.StringType,
	}
}

func ModuleAttributeValue(module map[string]interface{}) map[string]attr.Value {
	moduleId := types.StringNull()
	moduleName := types.StringNull()
	connectivityState := types.StringNull()

	for k, v := range module {
		switch k {
		case "moduleId":
			moduleId = types.StringValue(v.(string))
		case "moduleName":
			moduleName = types.StringValue(v.(string))
		case "connectivityState":
			connectivityState = types.StringValue(v.(string))
		}
	}

	return map[string]attr.Value{
		"module_id":          moduleId,
		"module_name":        moduleName,
		"connectivity_state": connectivityState,
	}
}
