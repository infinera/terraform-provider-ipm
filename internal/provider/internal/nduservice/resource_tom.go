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
	_ resource.Resource                = &TOMResource{}
	_ resource.ResourceWithConfigure   = &TOMResource{}
	_ resource.ResourceWithImportState = &TOMResource{}
)

// NewTOMResource is a helper function to simplify the provider implementation.
func NewTOMResource() resource.Resource {
	return &TOMResource{}
}

type TOMResource struct {
	client *ipm_pf.Client
}

type TOMConfig struct {
	Name            types.String `tfsdk:"name"`
	RequiredType    types.String `tfsdk:"required_type"`
	RequiredSubType types.String `tfsdk:"required_sub_type"`
	PhyMode         types.String `tfsdk:"phy_mode"`
}

type TOMResourceData struct {
	Id         types.String  `tfsdk:"id"`
	ParentId   types.String   `tfsdk:"parent_id"`
	Href       types.String   `tfsdk:"href"`
	ColId      types.Int64    `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config    *TOMConfig   `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *TOMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_tom"
}

// Schema defines the schema for the data source.
func (r *TOMResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type TOMResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages NDU TOM",
		Attributes:  TOMResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *TOMResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r TOMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TOMResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "TOMResource: Create - ", map[string]interface{}{"TOMResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r TOMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TOMResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "TOMResource: Create - ", map[string]interface{}{"TOMResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r TOMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TOMResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "TOMResource: Update", map[string]interface{}{"TOMResourceData": data})

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

func (r TOMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TOMResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "TOMResource: Update", map[string]interface{}{"TOMResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *TOMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TOMResource) create(plan *TOMResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TOMResource: create ##", map[string]interface{}{"plan": plan})

	if plan.Identifier.DeviceId.IsNull() && plan.Identifier.ParentColId.IsNull() {
		diags.AddError(
			"Error Create TOMResource",
			"Create: Could not create TOMResource. Resource Identifier is not specified",
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
	if !plan.Config.PhyMode.IsNull() {
		configRequest["phyMode"] = plan.Config.PhyMode.ValueString()
	}
	if !plan.Config.RequiredType.IsNull() {
		configRequest["requiredType"] = plan.Config.RequiredType.ValueString()
	}

	tflog.Debug(ctx, "TOMResource: create ## ", map[string]interface{}{"Create Request": configRequest})

	// send create request to server
	rb, err := json.Marshal(configRequest)
	if err != nil {
		diags.AddError(
			"TOMResource: create ##: Error Create AC",
			"Create: Could not Marshal TOMResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" + plan.Identifier.ParentColId.ValueString() + "/tom", rb)
	if err != nil {
		if !strings.Contains(err.Error(), "status: 202") {
			diags.AddError(
				"TOMResource: create ##: Error create TOMResource",
				"Create:Could not create TOMResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "TOMResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TOMResource: Create ##: Error Unmarshal response",
			"Update:Could not Create TOMResource, unexpected error: "+err.Error(),
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
		tflog.Debug(ctx, "TOMResource: create failed. Can't find the created VOA")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "TOMResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *TOMResource) update(plan *TOMResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TOMResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"TOMResource: Error update TOM",
			"TOMResource: Could not update TOM. Href, NDUId, PortColId, LinePTPColId, TOM ColId is not specified.",
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
	if !plan.Config.PhyMode.IsNull() {
		updateRequest["phyMode"] = plan.Config.PhyMode.ValueString()
	}

	tflog.Debug(ctx, "TOMResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"TOMResource: update ##: Error Create AC",
				"Create: Could not Marshal TOMResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ParentColId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" +  plan.Identifier.ParentColId.ValueString()  + "/tom/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"TOMResource: update ##: Error update TOMResource",
				"Update: Could not update TOMResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "TOMResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"TOMResource: Create ##: Error Unmarshal response",
				"Update:Could not Create TOMResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "TOMResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "TOMResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *TOMResource) delete(plan *TOMResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.Href.IsNull() && plan.Id.IsNull() {
		diags.AddError(
			"Error Delete TOMResource",
			"Read: Could not delete. NC ID is not specified",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", plan.Href.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"TOMResource: delete ##: Error Delete TOMResource",
			"Update:Could not delete TOMResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "TOMResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (r *TOMResource) read(state *TOMResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read TOMResource",
			"TOMResource: Could not read. ParentId and Id and Href are not specified.",
		)
		return
	}

	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/tom" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/tom" + queryStr + "&q={\"state.tomAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/tom" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/tom/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		diags.AddError(
			"TOMResource: read ##: Error Read TOMResource",
			"Read:Could not get TOMResource. No identifier specified",
		)
		return
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"TOMResource: read ##: Error Read TOMResource",
			"Read:Could not get TOMResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "TOMResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"TOMResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal TOMResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"TOMResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "TOMResource: read ## ", map[string]interface{}{"plan": state})
}

func (tomData *TOMResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "TOMResourceData: populate ## ", map[string]interface{}{"plan": data})

	tomData.Id = types.StringValue(data["id"].(string))
	tomData.ColId = types.Int64Value(int64(data["colid"].(float64)))
	tomData.Href = types.StringValue(data["href"].(string))
	tomData.ParentId = types.StringValue(data["parentId"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if tomData.Config == nil {
		tomData.Config = &TOMConfig{}
	}
	if config != nil {
		for k, v := range config {
			switch k {
			case "name":
				if !tomData.Config.Name.IsNull() {
					tomData.Config.Name = types.StringValue(v.(string))
				}
			case "requiredType":
				if !tomData.Config.RequiredType.IsNull() {
					tomData.Config.RequiredType = types.StringValue(v.(string))
				}
			case "requiredSubType":
				if !tomData.Config.RequiredSubType.IsNull() {
					tomData.Config.RequiredSubType = types.StringValue(v.(string))
				}
			case "phyMode":
				if !tomData.Config.PhyMode.IsNull() {
					tomData.Config.PhyMode = types.StringValue(v.(string))
				}
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		tomData.State = types.ObjectValueMust(TOMStateAttributeType(), TOMStateAttributeValue(state))
	}

	tflog.Debug(ctx, "TOMResourceData: read ## ", map[string]interface{}{"tomData": tomData})
}

func TOMResourceSchemaAttributes() map[string]schema.Attribute {
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
				"phy_mode": schema.StringAttribute{
					Description: "phy_mode",
					Optional:    true,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: TOMStateAttributeType(),
		},
	}
}

func TOMObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: TOMAttributeType(),
	}
}

func TOMObjectsValue(data []interface{}) []attr.Value {
	toms := []attr.Value{}
	for _, v := range data {
		tom := v.(map[string]interface{})
		if tom != nil {
			toms = append(toms, types.ObjectValueMust(
				TOMAttributeType(),
				TOMAttributeValue(tom)))
		}
	}
	return toms
}

func TOMAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":          types.StringType,
		"href":        types.StringType,
		"col_id":      types.Int64Type,
		"config":      types.ObjectType{AttrTypes: TOMConfigAttributeType()},
		"state":       types.ObjectType{AttrTypes: TOMStateAttributeType()},
	}
}

func TOMAttributeValue(tom map[string]interface{}) map[string]attr.Value {
	colId := types.Int64Null()
	if tom["colId"] != nil {
		colId = types.Int64Value(int64(tom["colId"].(float64)))
	}
	href := types.StringNull()
	if tom["href"] != nil {
		href = types.StringValue(tom["href"].(string))
	}
	id := types.StringNull()
	if tom["id"] != nil {
		id = types.StringValue(tom["id"].(string))
	}
	parentId := types.StringNull()
	if tom["parentId"] != nil {
		parentId = types.StringValue(tom["parentId"].(string))
	}
	config := types.ObjectNull(TOMConfigAttributeType())
	if (tom["config"]) != nil {
		config = types.ObjectValueMust(TOMConfigAttributeType(), TOMConfigAttributeValue(tom["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(TOMStateAttributeType())
	if (tom["state"]) != nil {
		state = types.ObjectValueMust(TOMStateAttributeType(), TOMStateAttributeValue(tom["state"].(map[string]interface{})))
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
func TOMConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":              types.StringType,
		"required_type":     types.StringType,
		"required_sub_type": types.StringType,
		"phy_mode":          types.StringType,
	}
}

func TOMConfigAttributeValue(tomConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	requiredType := types.StringNull()
	requiredSubType := types.StringNull()
	phyMode := types.StringNull()

	if (tomConfig["name"]) != nil {
		name = types.StringValue(tomConfig["name"].(string))
	}
	if (tomConfig["requiredType"]) != nil {
		requiredType = types.StringValue(tomConfig["requiredType"].(string))
	}
	if (tomConfig["requiredSubType"]) != nil {
		requiredSubType = types.StringValue(tomConfig["requiredSubType"].(string))
	}
	if (tomConfig["phyMode"]) != nil {
		phyMode = types.StringValue(tomConfig["phyMode"].(string))
	}

	return map[string]attr.Value{
		"name":              name,
		"required_type":     requiredType,
		"required_sub_type": requiredSubType,
		"phy_mode":          phyMode,
	}
}

func TOMStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":             types.StringType,
		"tom_aid":                types.StringType,
		"required_type":          types.StringType,
		"required_sub_type":      types.StringType,
		"name":                   types.StringType,
		"phy_mode":               types.StringType,
		"supported_phy_modes":    types.ListType{ElemType: types.StringType},
		"vendor_compliance_code": types.Int64Type,
		"inventory":              types.ObjectType{AttrTypes: common.InventoryAttributeType()},
		"lifecycle_state":        types.StringType,
	}
}

func TOMStateAttributeValue(tomState map[string]interface{}) map[string]attr.Value {
	tomAid := types.StringNull()
	parentAid := types.StringNull()
	name := types.StringNull()
	requiredType := types.StringNull()
	requiredSubType := types.StringNull()
	phyMode := types.StringNull()
	supportedPhyModes := types.ListNull(types.StringType)
	vendorComplianceCode := types.Int64Null()
	inventory := types.ObjectNull(common.InventoryAttributeType())
	lifecycleState := types.StringNull()

	for k, v := range tomState {
		switch k {
		case "tomAid":
			tomAid = types.StringValue(v.(string))
		case "parentAid":
			parentAids := v.([]interface{})
			parentAid = types.StringValue(parentAids[0].(string))
		case "name":
			name = types.StringValue(v.(string))
		case "requiredType":
			requiredType = types.StringValue(v.(string))
		case "requiredSubType":
			requiredSubType = types.StringValue(v.(string))
		case "phyMode":
			phyMode = types.StringValue(v.(string))
		case "vendorComplianceCode":
			vendorComplianceCode = types.Int64Value(int64(v.(float64)))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		case "inventory":
			inventory = types.ObjectValueMust(common.InventoryAttributeType(), common.InventoryAttributeValue(v.(map[string]interface{})))
		case "supportedPhyModes":
			supportedPhyModes = types.ListValueMust(types.StringType, common.ListAttributeStringValue(v.([]interface{})))
		}
	}

	return map[string]attr.Value{
		"tom_aid":                tomAid,
		"parent_aid":          parentAid,
		"name":                   name,
		"required_type":          requiredType,
		"required_sub_type":      requiredSubType,
		"phy_mode":               phyMode,
		"supported_phy_modes":    supportedPhyModes,
		"vendor_compliance_code": vendorComplianceCode,
		"lifecycle_state":        lifecycleState,
		"inventory":              inventory,
	}
}
