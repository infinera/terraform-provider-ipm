package nduservice

import (
	"context"
	"encoding/json"
	"strconv"
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
	_ resource.Resource                = &VOAResource{}
	_ resource.ResourceWithConfigure   = &VOAResource{}
	_ resource.ResourceWithImportState = &VOAResource{}
)

// NewVOAResource is a helper function to simplify the provider implementation.
func NewVOAResource() resource.Resource {
	return &VOAResource{}
}

type VOAResource struct {
	client *ipm_pf.Client
}

type VOAConfig struct {
	Name             types.String `tfsdk:"name"`
	RequiredType     types.String `tfsdk:"required_type"`
	RequiredSubType  types.String `tfsdk:"required_sub_type"`
	VOAShutterTx     types.Bool   `tfsdk:"voa_shutter_tx"`
	VOAAttenuationTx types.Int64  `tfsdk:"voa_attenuation_tx"`
}

type VOAResourceData struct {
	Id         types.String  `tfsdk:"id"`
	ParentId   types.String   `tfsdk:"parent_id"`
	Href       types.String   `tfsdk:"href"`
	ColId      types.Int64    `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config    *VOAConfig   `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *VOAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_voa"
}

// Schema defines the schema for the data source.
func (r *VOAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type VOAResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages NDU VOA",
		Attributes:  VOAResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *VOAResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r VOAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VOAResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "VOAResource: Create - ", map[string]interface{}{"VOAResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r VOAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VOAResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "VOAResource: Read - ", map[string]interface{}{"VOAResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r VOAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VOAResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "VOAResource: Update", map[string]interface{}{"VOAResourceData": data})

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

func (r VOAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VOAResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "VOAResource: Update", map[string]interface{}{"VOAResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *VOAResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VOAResource) create(plan *VOAResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "VOAResource: create ##", map[string]interface{}{"plan": plan})

	if plan.Identifier.DeviceId.IsNull() && plan.Identifier.ColId.IsNull() {
		diags.AddError(
			"Error Create VOAResource",
			"Create: Could not create VOAResource. Device ID and resource Col Id is not specified",
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
	if !plan.Config.VOAAttenuationTx.IsNull() {
		configRequest["voaAttenuationTx"] = plan.Config.VOAAttenuationTx.ValueInt64()
	}
	if !plan.Config.VOAShutterTx.IsNull() {
		configRequest["voaShutterTx"] = plan.Config.VOAShutterTx.ValueBool()
	}

	tflog.Debug(ctx, "VOAResource: create ## ", map[string]interface{}{"Create Request": configRequest})

	// send create request to server
	rb, err := json.Marshal(configRequest)
	if err != nil {
		diags.AddError(
			"VOAResource: create ##: Error Create AC",
			"Create: Could not Marshal VOAResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/voa/" + plan.Identifier.ColId.ValueString(), rb)
	if err != nil {
		if !strings.Contains(err.Error(), "status: 202") {
			diags.AddError(
				"VOAResource: create ##: Error create VOAResource",
				"Create:Could not create VOAResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "VOAResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"VOAResource: Create ##: Error Unmarshal response",
			"Update:Could not Create VOAResource, unexpected error: "+err.Error(),
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
		tflog.Debug(ctx, "VOAResource: create failed. Can't find the created VOA")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "VOAResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *VOAResource) update(plan *VOAResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "VOAResource: update ## ", map[string]interface{}{"plan": plan})
	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"VOAResource: Error update VOA",
			"VOAResource: Could not update VOA. Href, NDUId, PortColId, LinePTPColId, VOA ColId is not specified.",
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
	if !plan.Config.VOAShutterTx.IsNull() {
		updateRequest["voaShutterTx"] = plan.Config.VOAShutterTx.ValueBool()
	}
	if !plan.Config.VOAAttenuationTx.IsNull() {
		updateRequest["voaAttenuationTx"] = plan.Config.VOAAttenuationTx.ValueInt64()
	}

	tflog.Debug(ctx, "VOAResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"VOAResource: update ##: Error Create AC",
				"Create: Could not Marshal VOAResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.ColId.IsNull() && !plan.ParentId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.ParentId.ValueString() + "/voa/" +  strconv.FormatInt(plan.ColId.ValueInt64(),10), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/voa/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"VOAResource: update ##: Error update VOAResource",
				"Update: Could not update VOAResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "VOAResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"VOAResource: Create ##: Error Unmarshal response",
				"Update:Could not Create VOAResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "VOAResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "VOAResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *VOAResource) read(state *VOAResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read VOAResource",
			"VOAResource: Could not read. ParentId and Id and Href are not specified.",
		)
		return
	}

	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/voa" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/voa" + queryStr + "&q={\"state.tomAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/voa" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/voa/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/voa" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"VOAResource: read ##: Error Read VOAResource",
			"Read:Could not get VOAResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "VOAResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"VOAResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal VOAResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"VOAResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "VOAResource: read ## ", map[string]interface{}{"plan": state})
}

func (r *VOAResource) delete(state *VOAResourceData, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "VOAResource: delete ## ", map[string]interface{}{"state": state})

		if state.Href.IsNull() && state.Id.IsNull() {
			diags.AddError(
				"Error Delete VOAResource",
				"Read: Could not delete. NC ID is not specified",
			)
			return
		}
	
		_, err := r.client.ExecuteIPMHttpCommand("DELETE", state.Href.ValueString(), nil)
		if err != nil {
			diags.AddError(
				"VOAResource: delete ##: Error Delete VOAResource",
				"Update:Could not delete VOAResource, unexpected error: "+err.Error(),
			)
			return
		}
	
		tflog.Debug(ctx, "VOAResource: delete ## ", map[string]interface{}{"state": state})
}

func (voaData *VOAResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "VOAResourceData: populate ## ", map[string]interface{}{"plan": data})

	voaData.Id = types.StringValue(data["id"].(string))
	voaData.ColId = types.Int64Value(int64(data["colid"].(float64)))
	voaData.Href = types.StringValue(data["href"].(string))
	voaData.ParentId = types.StringValue(data["parentId"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if config != nil {
		if voaData.Config == nil {
			voaData.Config = &VOAConfig{}
		}
		for k, v := range config {
			switch k {
			case "name":
				if !voaData.Config.Name.IsNull() {
					voaData.Config.Name = types.StringValue(v.(string))
				}
			case "requiredType":
				if !voaData.Config.RequiredType.IsNull() {
					voaData.Config.RequiredType = types.StringValue(v.(string))
				}
			case "requiredSubType":
				if !voaData.Config.RequiredSubType.IsNull() {
					voaData.Config.RequiredSubType = types.StringValue(v.(string))
				}
			case "voaShutterTx":
				if !voaData.Config.VOAShutterTx.IsNull() {
					voaData.Config.VOAShutterTx = types.BoolValue(v.(bool))
				}
			case "voaAttenuationTx":
				if !voaData.Config.VOAAttenuationTx.IsNull() {
					voaData.Config.VOAAttenuationTx = types.Int64Value(int64(v.(float64)))
				}
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		voaData.State = types.ObjectValueMust(VOAStateAttributeType(), VOAStateAttributeValue(state))
	}

	tflog.Debug(ctx, "VOAResourceData: read ## ", map[string]interface{}{"voaData": voaData})
}

func VOAResourceSchemaAttributes() map[string]schema.Attribute {
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
				"voa_shutter_tx": schema.BoolAttribute{
					Description: "voa_shutter_tx",
					Optional:    true,
				},
				"voa_attenuation_tx": schema.Int64Attribute{
					Description: "voa_attenuation_tx",
					Optional:    true,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: VOAStateAttributeType(),
		},
	}
}

func VOAObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: VOAAttributeType(),
	}
}

func VOAObjectsValue(data []interface{}) []attr.Value {
	voas := []attr.Value{}
	for _, v := range data {
		voa := v.(map[string]interface{})
		if voa != nil {
			voas = append(voas, types.ObjectValueMust(
				VOAAttributeType(),
				VOAAttributeValue(voa)))
		}
	}
	return voas
}

func VOAAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":          types.StringType,
		"href":        types.StringType,
		"col_id":      types.Int64Type,
		"config":      types.ObjectType{AttrTypes: VOAConfigAttributeType()},
		"state":       types.ObjectType{AttrTypes: VOAStateAttributeType()},
	}
}

func VOAAttributeValue(voa map[string]interface{}) map[string]attr.Value {
	colId := types.Int64Null()
	if voa["colId"] != nil {
		colId = types.Int64Value(int64(voa["colId"].(float64)))
	}
	href := types.StringNull()
	if voa["href"] != nil {
		href = types.StringValue(voa["href"].(string))
	}
	id := types.StringNull()
	if voa["id"] != nil {
		id = types.StringValue(voa["id"].(string))
	}
	parentId := types.StringNull()
	if voa["parentId"] != nil {
		parentId = types.StringValue(voa["parentId"].(string))
	}
	config := types.ObjectNull(VOAConfigAttributeType())
	if (voa["config"]) != nil {
		config = types.ObjectValueMust(VOAConfigAttributeType(), VOAConfigAttributeValue(voa["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(VOAStateAttributeType())
	if (voa["state"]) != nil {
		state = types.ObjectValueMust(VOAStateAttributeType(), VOAStateAttributeValue(voa["state"].(map[string]interface{})))
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
func VOAConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":      types.StringType,
		"required_type":      types.StringType,
		"required_sub_type":  types.StringType,
		"voa_shutter_tx":     types.BoolType,
		"voa_attenuation_tx": types.Int64Type,
	}
}

func VOAConfigAttributeValue(voaConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	requiredType := types.StringNull()
	requiredSubType := types.StringNull()
	voaShutterTx := types.BoolNull()
	voaAttenuationTx := types.Int64Null()

	for k, v := range voaConfig {
		switch k {
		case "name":
			requiredType = types.StringValue(v.(string))
		case "requiredType":
			requiredType = types.StringValue(v.(string))
		case "requiredSubType":
			requiredSubType = types.StringValue(v.(string))
		case "amplifierEnable":
			voaShutterTx = types.BoolValue(v.(bool))
		case "gainTarget":
			voaAttenuationTx = types.Int64Value(int64(v.(float64)))
		}
	}

	return map[string]attr.Value{
		"name":               name,
		"required_type":      requiredType,
		"required_sub_type":  requiredSubType,
		"voa_shutter_tx":     voaShutterTx,
		"voa_attenuation_tx": voaAttenuationTx,
	}
}

func VOAStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":         			types.StringType,
		"name":                      types.StringType,
		"voa_aid":                   types.StringType,
		"required_type":             types.StringType,
		"required_sub_type":         types.StringType,
		"voa_shutter_tx":            types.BoolType,
		"voa_attenuation_tx":        types.Int64Type,
		"supported_min_attenuation": types.Int64Type,
		"supported_max_attenuation": types.Int64Type,
		"inventory":                 types.ObjectType{AttrTypes: common.InventoryAttributeType()},
		"lifecycle_state":           types.StringType,
	}
}

func VOAStateAttributeValue(voaState map[string]interface{}) map[string]attr.Value {
	voaAid := types.StringNull()
	parentAid := types.StringNull()
	name := types.StringNull()
	requiredType := types.StringNull()
	requiredSubType := types.StringNull()
	voaShutterTx := types.BoolNull()
	voaAttenuationTx := types.Int64Null()
	supportedMinAttenuation := types.Int64Null()
	supportedMaxAttenuation := types.Int64Null()
	inventory := types.ObjectNull(common.InventoryAttributeType())
	lifecycleState := types.StringNull()

	for k, v := range voaState {
		switch k {
		case "voaAid":
			voaAid = types.StringValue(v.(string))
		case "parentAid":
			parentAids := v.([]interface{})
			parentAid = types.StringValue(parentAids[0].(string))
		case "name":
			name = types.StringValue(v.(string))
		case "requiredType":
			requiredType = types.StringValue(v.(string))
		case "requiredSubType":
			requiredSubType = types.StringValue(v.(string))
		case "voaShutterTx":
			voaShutterTx = types.BoolValue(v.(bool))
		case "voaAttenuationTx":
			voaAttenuationTx = types.Int64Value(int64(v.(float64)))
		case "supportedMinAttenuation":
			supportedMinAttenuation = types.Int64Value(int64(v.(float64)))
		case "supportedMaxAttenuation":
			supportedMaxAttenuation = types.Int64Value(int64(v.(float64)))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		case "inventory":
			inventory = types.ObjectValueMust(common.InventoryAttributeType(), common.InventoryAttributeValue(v.(map[string]interface{})))
		}
	}

	return map[string]attr.Value{
		"voa_aid":                   voaAid,
		"parent_aid":                parentAid,
		"name":                      name,
		"required_type":             requiredType,
		"required_sub_type":         requiredSubType,
		"voa_shutter_tx":            voaShutterTx,
		"voa_attenuation_tx":        voaAttenuationTx,
		"supported_min_attenuation": supportedMinAttenuation,
		"supported_max_attenuation": supportedMaxAttenuation,
		"lifecycle_state":           lifecycleState,
		"inventory":                 inventory,
	}
}
