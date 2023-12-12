package nduservice

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

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &LEDsResource{}
	_ resource.ResourceWithConfigure   = &LEDsResource{}
	_ resource.ResourceWithImportState = &LEDsResource{}
)

// NewLEDsResource is a helper function to simplify the provider implementation.
func NewLEDsResource() resource.Resource {
	return &LEDsResource{}
}

type LEDsResource struct {
	client *ipm_pf.Client
}

type LEDsResourceData struct {
	Id    types.String `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href  types.String `tfsdk:"href"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	State types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *LEDsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_leds"
}

// Schema defines the schema for the data source.
func (r *LEDsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type LEDsResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an NDU leds",
		Attributes:  LEDsResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *LEDsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r LEDsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LEDsResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "LEDsResource: Create - ", map[string]interface{}{"LEDsResourceData": data})

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r LEDsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LEDsResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LEDsResource: Create - ", map[string]interface{}{"LEDsResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LEDsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LEDsResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "LEDsResource: Update", map[string]interface{}{"LEDsResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LEDsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LEDsResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LEDsResource: Update", map[string]interface{}{"LEDsResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *LEDsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *LEDsResource) read(state *LEDsResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Href.IsNull() && state.Id.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() {
		diags.AddError(
			"Error Read LEDsResource",
			"LEDsResource: Could not read. Leds. Href or (NDUId and Leds ColId) is not specified.",
		)
		return
	}
	tflog.Debug(ctx, "LEDsResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/leds" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Id.ValueString() + "/leds" + queryStr 
	}else if !state.Identifier.DeviceId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/leds" + queryStr 
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/leds" + queryStr 
	} else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() + "/leds" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"LEDsResource: read ##: Error Read LEDsResource",
			"Read:Could not get LEDsResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "LEDsResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})

	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"LEDsResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal LEDsResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"LEDsResource: read ##: Can not get Module",
				"Read:Could not get EClient for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "LEDsResource: read ## ", map[string]interface{}{"plan": state})
}

func (leds *LEDsResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "LEDsResourceData: populate ## ", map[string]interface{}{"plan": data})

leds.Id = types.StringValue(data["id"].(string))
leds.ParentId = types.StringValue(data["parentId"].(string))
leds.Href = types.StringValue(data["href"].(string))

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		leds.State = types.ObjectValueMust(LEDsStateAttributeType(), LEDsStateAttributeValue(state))
	}

	tflog.Debug(ctx, "LEDsResourceData: read ## ", map[string]interface{}{"plan": state})
}

func LEDsResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Module.",
			Computed:    true,
		},
		"parent_id": schema.StringAttribute{
			Description: "module id",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		"identifier": common.ResourceIdentifierAttribute(),
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: LEDsStateAttributeType(),
		},
	}
}

func LEDsObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: LEDsAttributeType(),
	}
}

func LEDsObjectsValue(data []interface{}) []attr.Value {
	leds := []attr.Value{}
	for _, v := range data {
		led := v.(map[string]interface{})
		if led != nil {
			leds = append(leds, types.ObjectValueMust(
				LEDsAttributeType(),
				LEDsAttributeValue(led)))
		}
	}
	return leds
}

func LEDsAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":     types.StringType,
		"href":   types.StringType,
		"state":  types.ObjectType{AttrTypes: LEDsStateAttributeType()},
	}
}

func LEDsAttributeValue(leds map[string]interface{}) map[string]attr.Value {
	href := types.StringNull()
	if leds["href"] != nil {
		href = types.StringValue(leds["href"].(string))
	}
	id := types.StringNull()
	if leds["id"] != nil {
		id = types.StringValue(leds["id"].(string))
	}
	parentId := types.StringNull()
	if leds["parentId"] != nil {
		parentId = types.StringValue(leds["parentId"].(string))
	}
	state := types.ObjectNull(LEDsStateAttributeType())
	if (leds["state"]) != nil {
		state = types.ObjectValueMust(LEDsStateAttributeType(), LEDsStateAttributeValue(leds["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"parent_id": parentId,
		"id":     id,
		"href":   href,
		"state":  state,
	}
}

func LEDsStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":       types.StringType,
		"leds_aid":  types.StringType,
		"led_states": types.ListType{ElemType: LEDsStateStateObjectType()},
	}
}

func LEDsStateAttributeValue(ledsState map[string]interface{}) map[string]attr.Value {
	ledsAid := types.StringNull()
	parentAid := types.StringNull()
	ledStates := types.ListNull(LEDsStateStateObjectType())

	for k, v := range ledsState {
		switch k {
		case "ledsAid":
			ledsAid = types.StringValue(v.(string))
		case "parentAid":
			parentAids := v.([]interface{})
			parentAid = types.StringValue(parentAids[0].(string))
		case "ledStates":
			ledStates = types.ListValueMust(LEDsStateStateObjectType(), LEDsStateStateObjectsValue(v.([]interface{})))
		}
	}

	return map[string]attr.Value{
		"parent_aid":       parentAid,
		"leds_aid":   ledsAid,
		"led_states": ledStates,
	}
}

func LEDsStateStateObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: LEDsStateStateAttributeType(),
	}
}

func LEDsStateStateObjectsValue(data []interface{}) []attr.Value {
	ledsStateStates := []attr.Value{}
	for _, v := range data {
		ledsStateState := v.(map[string]interface{})
		if ledsStateState != nil {
			ledsStateStates = append(ledsStateStates, types.ObjectValueMust(
				LEDsStateStateAttributeType(),
				LEDsStateStateAttributValue(ledsStateState)))
		}
	}
	return ledsStateStates
}

func LEDsStateStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"state":  types.StringType,
		"led_id": types.StringType,
	}
}

func LEDsStateStateAttributValue(data map[string]interface{}) map[string]attr.Value {
	state := types.StringNull()
	ledId := types.StringNull()

	for k, v := range data {
		switch k {
		case "ledId":
			ledId = types.StringValue(v.(string))
		case "state":
			state = types.StringValue(v.(string))
		}
	}

	return map[string]attr.Value{
		"state":  state,
		"led_id": ledId,
	}
}
