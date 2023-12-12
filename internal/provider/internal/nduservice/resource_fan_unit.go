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
	_ resource.Resource                = &FanUnitResource{}
	_ resource.ResourceWithConfigure   = &FanUnitResource{}
	_ resource.ResourceWithImportState = &FanUnitResource{}
)

// NewFanUnitResource is a helper function to simplify the provider implementation.
func NewFanUnitResource() resource.Resource {
	return &FanUnitResource{}
}

type FanUnitResource struct {
	client *ipm_pf.Client
}

type FanUnitResourceData struct {
	Id         types.String  `tfsdk:"id"`
	ParentId   types.String  `tfsdk:"parent_id"`
	Href       types.String  `tfsdk:"href"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	State types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *FanUnitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_fan_unit"
}

// Schema defines the schema for the data source.
func (r *FanUnitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type FanUnitResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an NDU fan",
		Attributes:  FanUnitResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *FanUnitResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r FanUnitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FanUnitResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "FanUnitResource: Create - ", map[string]interface{}{"FanUnitResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r FanUnitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FanUnitResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "FanUnitResource: Create - ", map[string]interface{}{"FanUnitResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r FanUnitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FanUnitResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "FanUnitResource: Update", map[string]interface{}{"FanUnitResourceData": data})

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

func (r FanUnitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FanUnitResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "FanUnitResource: Update", map[string]interface{}{"FanUnitResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *FanUnitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *FanUnitResource) read(state *FanUnitResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Href.IsNull() && state.Id.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() {
		diags.AddError(
			"Error Read FanUnitResource",
			"FanUnitResource: Could not read. Fan. Href or (NDUId and Fan ColId) is not specified.",
		)
		return
	}

	tflog.Debug(ctx, "FanUnitResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/fanUnit" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/fanUnit" + queryStr + "&q={\"state.fanAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/fanUnit" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else {
		diags.AddError(
			"FanUnitResource: read ##: Error Read FanUnitResource",
			"Read:Could not get FanUnitResource. No identifier specified.",
		)
		return
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"FanUnitResource: read ##: Error Read FanUnitResource",
			"Read:Could not get FanUnitResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "FanUnitResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})

	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"FanUnitResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal FanUnitResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"FanUnitResource: read ##: Can not get Module",
				"Read:Could not get EClient for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "FanUnitResource: read ## ", map[string]interface{}{"plan": state})
}

func (fanUnit *FanUnitResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "FanUnitResourceData: populate ## ", map[string]interface{}{"plan": data})

	fanUnit.Id = types.StringValue(data["id"].(string))
	fanUnit.ParentId = types.StringValue(data["parentId"].(string))
	fanUnit.Href = types.StringValue(data["href"].(string))

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		fanUnit.State = types.ObjectValueMust(FanUnitStateAttributeType(), FanUnitStateAttributeValue(state))
	}

	tflog.Debug(ctx, "FanUnitResourceData: read ## ", map[string]interface{}{"plan": state})
}

func FanUnitResourceSchemaAttributes() map[string]schema.Attribute {
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
			AttributeTypes: FanUnitStateAttributeType(),
		},
	}
}

func FanUnitObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: FanUnitAttributeType(),
	}
}

func FanUnitObjectsValue(data []interface{}) []attr.Value {
	fanUnits := []attr.Value{}
	for _, v := range data {
		fanUnit := v.(map[string]interface{})
		if fanUnit != nil {
			fanUnits = append(fanUnits, types.ObjectValueMust(
				FanUnitAttributeType(),
				FanUnitAttributeValue(fanUnit)))
		}
	}
	return fanUnits
}

func FanUnitAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":        types.StringType,
		"href":      types.StringType,
		"state":  types.ObjectType{AttrTypes: FanUnitStateAttributeType()},
	}
}

func FanUnitAttributeValue(fan map[string]interface{}) map[string]attr.Value {
	href := types.StringNull()
	if fan["href"] != nil {
		href = types.StringValue(fan["href"].(string))
	}
	id := types.StringNull()
	if fan["id"] != nil {
		id = types.StringValue(fan["id"].(string))
	}
	parentId := types.StringNull()
	if fan["parentId"] != nil {
		parentId = types.StringValue(fan["parentId"].(string))
	}
	state := types.ObjectNull(FanUnitStateAttributeType())
	if (fan["state"]) != nil {
		state = types.ObjectValueMust(FanUnitStateAttributeType(), FanUnitStateAttributeValue(fan["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"parent_id": parentId,
		"id":     id,
		"href":   href,
		"state":  state,
	}
}

func FanUnitStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"fan_aid":   types.StringType,
		"parent_aid":       types.StringType,
		"state":     types.StringType,
		"fans":      types.ListType{ElemType: FanObjectType()},
		"inventory": types.ObjectType{AttrTypes: common.InventoryAttributeType()},
	}
}

func FanUnitStateAttributeValue(fanState map[string]interface{}) map[string]attr.Value {
	parentAid := types.StringNull()
	if fanState["parentAid"] != nil {
		parentAids := fanState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	fanAid := types.StringNull()
	if fanState["fanAid"] != nil {
		fanAid = types.StringValue(fanState["fanAid"].(string))
	}
	state := types.StringNull()
	if fanState["state"] != nil {
		state = types.StringValue(fanState["state"].(string))
	}
	fans := types.ListNull(FanObjectType())
	if fanState["fans"] != nil {
		fans = types.ListValueMust(FanObjectType(), FanObjectsValue(fanState["fans"].([]interface{})))
	}
	inventory := types.ObjectNull(common.InventoryAttributeType())
	if fanState["inventory"] != nil {
		inventory = types.ObjectValueMust(common.InventoryAttributeType(), common.InventoryAttributeValue(fanState["inventory"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"parent_aid":       parentAid,
		"fan_aid":   fanAid,
		"state":     state,
		"fans":      fans,
		"inventory": inventory,
	}
}

func FanObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: FanAttributeType(),
	}
}

func FanObjectsValue(data []interface{}) []attr.Value {
	fans := []attr.Value{}
	for _, v := range data {
		fan := v.(map[string]interface{})
		if fan != nil {
			fans = append(fans, types.ObjectValueMust(
				FanAttributeType(),
				FanAttributeValue(fan)))
		}
	}
	return fans
}

func FanAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"state": types.StringType,
		"speed": types.Int64Type,
	}
}

func FanAttributeValue(fan map[string]interface{}) map[string]attr.Value {
	state := types.StringNull()
	if fan["state"] != nil {
		state = types.StringValue(fan["state"].(string))
	}
	speed := types.Int64Null()
	if fan["speed"] != nil {
		speed = types.Int64Value(int64(fan["speed"].(float64)))
	}

	return map[string]attr.Value{
		"state": state,
		"speed": speed,
	}
}
