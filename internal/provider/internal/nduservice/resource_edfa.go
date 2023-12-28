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
	_ resource.Resource                = &EDFAResource{}
	_ resource.ResourceWithConfigure   = &EDFAResource{}
	_ resource.ResourceWithImportState = &EDFAResource{}
)

// NewEDFAResource is a helper function to simplify the provider implementation.
func NewEDFAResource() resource.Resource {
	return &EDFAResource{}
}

type EDFAResource struct {
	client *ipm_pf.Client
}

type EDFAConfig struct {
	Name            types.String `tfsdk:"name"`
	RequiredType    types.String `tfsdk:"required_type"`
	RequiredSubType types.String `tfsdk:"required_sub_type"`
	AmplifierEnable types.Bool   `tfsdk:"amplifier_enable"`
	GainTarget      types.Int64  `tfsdk:"gain_target"`
}

type EDFAResourceData struct {
	Id         types.String  `tfsdk:"id"`
	ParentId   types.String   `tfsdk:"parent_id"`
	Href       types.String   `tfsdk:"href"`
	ColId      types.Int64    `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config    *EDFAConfig  `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *EDFAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_edfa"
}

// Schema defines the schema for the data source.
func (r *EDFAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type EDFAResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages NDU EDFA",
		Attributes:  EDFAResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *EDFAResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r EDFAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EDFAResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "EDFAResource: Create - ", map[string]interface{}{"EDFAResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r EDFAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EDFAResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "EDFAResource: Create - ", map[string]interface{}{"EDFAResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r EDFAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EDFAResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "EDFAResource: Update", map[string]interface{}{"EDFAResourceData": data})

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

func (r EDFAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EDFAResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "EDFAResource: Update", map[string]interface{}{"EDFAResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *EDFAResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *EDFAResource) create(plan *EDFAResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "EDFAResource: create ##", map[string]interface{}{"plan": plan})
	
		if plan.Identifier.DeviceId.IsNull() && plan.Identifier.ParentColId.IsNull() {
			diags.AddError(
				"Error Create EDFAResource",
				"Create: Could not create EDFAResource. Resource Identifier is not specified",
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
		if !plan.Config.AmplifierEnable.IsNull() {
			configRequest["amplifierEnable"] = plan.Config.AmplifierEnable.ValueBool()
		}
		if !plan.Config.RequiredType.IsNull() {
			configRequest["requiredType"] = plan.Config.RequiredType.ValueString()
		}
		if !plan.Config.GainTarget.IsNull() {
			configRequest["gainTarget"] = plan.Config.GainTarget.ValueInt64()
		}
	
		tflog.Debug(ctx, "EDFAResource: create ## ", map[string]interface{}{"Create Request": configRequest})
	
		// send create request to server
		rb, err := json.Marshal(configRequest)
		if err != nil {
			diags.AddError(
				"EDFAResource: create ##: Error Create AC",
				"Create: Could not Marshal EDFAResource, unexpected error: "+err.Error(),
			)
			return
		}
		body, err := r.client.ExecuteIPMHttpCommand("POST", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" + plan.Identifier.ParentColId.ValueString() + "/edfa", rb)
		if err != nil {
				diags.AddError(
					"EDFAResource: create ##: Error create EDFAResource",
					"Create:Could not create EDFAResource, unexpected error: "+err.Error(),
				)
				return
		}
	
		tflog.Debug(ctx, "EDFAResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"EDFAResource: Create ##: Error Unmarshal response",
				"Update:Could not Create EDFAResource, unexpected error: "+err.Error(),
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
			tflog.Debug(ctx, "EDFAResource: create failed. Can't find the created VOA")
			plan.Id = types.StringNull()
			plan.Href = types.StringNull()
			return
		}
	
		tflog.Debug(ctx, "EDFAResource: create ##", map[string]interface{}{"plan": plan})
	
}

func (r *EDFAResource) update(plan *EDFAResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "EDFAResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"EDFAResource: Error update EDFA",
			"EDFAResource: Could not update EDFA. Href, NDUId, PortColId, LinePTPColId, EDFA ColId is not specified.",
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
	if !plan.Config.AmplifierEnable.IsNull() {
		updateRequest["amplifierEnable"] = plan.Config.AmplifierEnable.ValueBool()
	}
	if !plan.Config.GainTarget.IsNull() {
		updateRequest["gainTarget"] = plan.Config.GainTarget.ValueInt64()
	}

	tflog.Debug(ctx, "EDFAResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"EDFAResource: update ##: Error Create AC",
				"Create: Could not Marshal EDFAResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ParentColId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" +  plan.Identifier.ParentColId.ValueString()  + "/edfa/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"EDFAResource: update ##: Error update EDFAResource",
				"Update: Could not update EDFAResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}
		tflog.Debug(ctx, "EDFAResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"EDFAResource: Create ##: Error Unmarshal response",
				"Update:Could not Create EDFAResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "EDFAResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "EDFAResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *EDFAResource) read(state *EDFAResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read EDFAResource",
			"EDFAResource: Could not read. Id, and Href, and identifiers are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "EDFAResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/edfa" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/edfa" + queryStr + "&q={\"state.edfaAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/edfa" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/edfa/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		diags.AddError(
			"EDFAResource: read ##: Error Read EDFAResource",
			"Read:Could not get EDFAResource, No Identifier specified: ",
		)
		return
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"EDFAResource: read ##: Error Read EDFAResource",
			"Read:Could not get EDFAResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "EDFAResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"EDFAResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal EDFAResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"EDFAResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "EDFAResource: read ## ", map[string]interface{}{"plan": state})
	
}

func (r *EDFAResource) delete(state *EDFAResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Href.IsNull() && state.Id.IsNull() {
		diags.AddError(
			"Error Delete EDFAResource",
			"Read: Could not delete. EDFAResource Href and ID is not specified",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", state.Href.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"EDFAResource: delete ##: Error Delete EDFAResource",
			"Update:Could not delete EDFAResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "EDFAResource: delete ## ", map[string]interface{}{"state": state})
}

func (edfaData *EDFAResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "EDFAResourceData: populate ## ", map[string]interface{}{"plan": data})

	edfaData.Id = types.StringValue(data["id"].(string))
	edfaData.ColId = types.Int64Value(int64(data["colid"].(float64)))
	edfaData.Href = types.StringValue(data["href"].(string))
	edfaData.ParentId = types.StringValue(data["parentId"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if edfaData.Config == nil {
		edfaData.Config = &EDFAConfig{}
	}
	for k, v := range config {
		switch k {
		case "name":
			if !edfaData.Config.Name.IsNull() {
				edfaData.Config.Name = types.StringValue(v.(string))
			}
		case "requiredType":
			if !edfaData.Config.RequiredType.IsNull() {
				edfaData.Config.RequiredType = types.StringValue(v.(string))
			}
		case "requiredSubType":
			if !edfaData.Config.RequiredSubType.IsNull() {
				edfaData.Config.RequiredSubType = types.StringValue(v.(string))
			}
		case "amplifierEnable":
			if !edfaData.Config.AmplifierEnable.IsNull() {
				edfaData.Config.AmplifierEnable = types.BoolValue(v.(bool))
			}
		case "gainTarget":
			if !edfaData.Config.GainTarget.IsNull() {
				edfaData.Config.GainTarget = types.Int64Value(int64(v.(float64)))
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		edfaData.State = types.ObjectValueMust(EDFAStateAttributeType(), EDFAStateAttributeValue(state))
	}

	tflog.Debug(ctx, "EDFAResourceData: read ## ", map[string]interface{}{"edfaData": edfaData})
}

func EDFAResourceSchemaAttributes() map[string]schema.Attribute {
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
				"amplifier_enable": schema.BoolAttribute{
					Description: "amplifier_enable",
					Optional:    true,
				},
				"gain_target": schema.Int64Attribute{
					Description: "gainTarget",
					Optional:    true,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: EDFAStateAttributeType(),
		},
	}
}

func EDFAObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: EDFAAttributeType(),
	}
}

func EDFAObjectsValue(data []interface{}) []attr.Value {
	edfas := []attr.Value{}
	for _, v := range data {
		edfa := v.(map[string]interface{})
		if edfa != nil {
			edfas = append(edfas, types.ObjectValueMust(
				EDFAAttributeType(),
				EDFAAttributeValue(edfa)))
		}
	}
	return edfas
}

func EDFAAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.StringType,
		"parent_id": types.StringType,
		"href":      types.StringType,
		"col_id":    types.Int64Type,
		"config":      types.ObjectType{AttrTypes: EDFAConfigAttributeType()},
		"state":       types.ObjectType{AttrTypes: EDFAStateAttributeType()},
	}
}

func EDFAAttributeValue(edfa map[string]interface{}) map[string]attr.Value {
	colId := types.Int64Null()
	if edfa["colId"] != nil {
		colId = types.Int64Value(int64(edfa["colId"].(float64)))
	}
	href := types.StringNull()
	if edfa["href"] != nil {
		href = types.StringValue(edfa["href"].(string))
	}
	id := types.StringNull()
	if edfa["id"] != nil {
		id = types.StringValue(edfa["id"].(string))
	}
	parentId := types.StringNull()
	if edfa["parentId"] != nil {
		parentId = types.StringValue(edfa["parentId"].(string))
	}
	config := types.ObjectNull(EDFAConfigAttributeType())
	if (edfa["config"]) != nil {
		config = types.ObjectValueMust(EDFAConfigAttributeType(), EDFAConfigAttributeValue(edfa["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(EDFAStateAttributeType())
	if (edfa["state"]) != nil {
		state = types.ObjectValueMust(EDFAStateAttributeType(), EDFAStateAttributeValue(edfa["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"col_id":          colId,
		"parent_id": parentId,
		"id":              id,
		"href":            href,
		"config":      config,
		"state":       state,
	}
}
func EDFAConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":              types.StringType,
		"required_type":     types.StringType,
		"required_sub_type": types.StringType,
		"amplifier_enable":  types.BoolType,
		"gain_target":       types.Int64Type,
	}
}

func EDFAConfigAttributeValue(edfaConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	requiredType := types.StringNull()
	requiredSubType := types.StringNull()
	amplifierEnable := types.BoolNull()
	gainTarget := types.Int64Null()

	for k, v := range edfaConfig {
		switch k {
		case "name":
			name = types.StringValue(v.(string))
		case "requiredType":
			requiredType = types.StringValue(v.(string))
		case "requiredSubType":
			requiredSubType = types.StringValue(v.(string))
		case "amplifierEnable":
			amplifierEnable = types.BoolValue(v.(bool))
		case "gainTarget":
			gainTarget = types.Int64Value(int64(v.(float64)))
		}
	}

	return map[string]attr.Value{
		"name":              name,
		"required_type":     requiredType,
		"required_sub_type": requiredSubType,
		"amplifier_enable":  amplifierEnable,
		"gain_target":       gainTarget,
	}
}

func EDFAStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"parent_aid":         types.StringType,
		"edfa_aid":           types.StringType,
		"required_type":      types.StringType,
		"required_sub_type":  types.StringType,
		"amplifier_enable":   types.BoolType,
		"gain_target":        types.Int64Type,
		"current_gain":       types.Int64Type,
		"supported_min_gain": types.Int64Type,
		"supported_max_gain": types.Int64Type,
		"inventory":          types.ObjectType{AttrTypes: common.InventoryAttributeType()},
		"lifecycle_state":    types.StringType,
	}
}

func EDFAStateAttributeValue(edfaState map[string]interface{}) map[string]attr.Value {
	edfaAid := types.StringNull()
	parentAid := types.StringNull()
	name := types.StringNull()
	requiredType := types.StringNull()
	requiredSubType := types.StringNull()
	amplifierEnable := types.BoolNull()
	gainTarget := types.Int64Null()
	currentGain := types.Int64Null()
	supportedMinGain := types.Int64Null()
	supportedMaxGain := types.Int64Null()
	inventory := types.ObjectNull(common.InventoryAttributeType())
	lifecycleState := types.StringNull()

	for k, v := range edfaState {
		switch k {
		case "edfaAid":
			edfaAid = types.StringValue(v.(string))
		case "parentAid":
			parentAids := v.([]interface{})
			parentAid = types.StringValue(parentAids[0].(string))
		case "name":
			name = types.StringValue(v.(string))
		case "requiredType":
			requiredType = types.StringValue(v.(string))
		case "requiredSubType":
			requiredSubType = types.StringValue(v.(string))
		case "amplifierEnable":
			amplifierEnable = types.BoolValue(v.(bool))
		case "gainTarget":
			gainTarget = types.Int64Value(int64(v.(float64)))
		case "currentGain":
			currentGain = types.Int64Value(int64(v.(float64)))
		case "supportedMinGain":
			supportedMinGain = types.Int64Value(int64(v.(float64)))
		case "supportedMaxGain":
			supportedMaxGain = types.Int64Value(int64(v.(float64)))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		case "inventory":
			inventory = types.ObjectValueMust(common.InventoryAttributeType(), common.InventoryAttributeValue(v.(map[string]interface{})))
		}
	}

	return map[string]attr.Value{
		"edfa_aid":           edfaAid,
		"parent_aid":          parentAid,
		"name":               name,
		"required_type":      requiredType,
		"required_sub_type":  requiredSubType,
		"amplifier_enable":   amplifierEnable,
		"gain_target":        gainTarget,
		"current_gain":       currentGain,
		"supported_min_gain": supportedMinGain,
		"supported_max_gain": supportedMaxGain,
		"lifecycle_state":   lifecycleState,
		"inventory":          inventory,
	}
}
