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
	_ resource.Resource                = &TribPTPResource{}
	_ resource.ResourceWithConfigure   = &TribPTPResource{}
	_ resource.ResourceWithImportState = &TribPTPResource{}
)

// NewTribPTPResource is a helper function to simplify the provider implementation.
func NewTribPTPResource() resource.Resource {
	return &TribPTPResource{}
}

type TribPTPResource struct {
	client *ipm_pf.Client
}

type TribPTPConfig struct {
	Name          types.String `tfsdk:"name"`
	TxLaserEnable types.Bool   `tfsdk:"tx_laser_enable"`
}

type TribPTPResourceData struct {
	Id         types.String  `tfsdk:"id"`
	ParentId   types.String   `tfsdk:"parent_id"`
	Href       types.String   `tfsdk:"href"`
	ColId      types.Int64    `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config    *TribPTPConfig `tfsdk:"config"`
	State     types.Object   `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *TribPTPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_tribPtp"
}

// Schema defines the schema for the data source.
func (r *TribPTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type TribPTPResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages NDU TribPTP",
		Attributes:  TribPTPResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *TribPTPResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r TribPTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TribPTPResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "TribPTPResource: Create - ", map[string]interface{}{"TribPTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(diags...)
}

func (r TribPTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TribPTPResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "TribPTPResource: Create - ", map[string]interface{}{"TribPTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r TribPTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TribPTPResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "TribPTPResource: Update", map[string]interface{}{"TribPTPResourceData": data})

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

func (r TribPTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TribPTPResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "TribPTPResource: Update", map[string]interface{}{"TribPTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *TribPTPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TribPTPResource) create(plan *TribPTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TribPTPResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *TribPTPResource) update(plan *TribPTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TribPTPResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"TribPTPResource: Error update TribPTP",
			"TribPTPResource: Could not update TribPTP. Href, NDUId, PortColId, LinePTPColId, TribPTP ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})
	// get TC config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.TxLaserEnable.IsNull() {
		updateRequest["txLaserEnable"] = plan.Config.TxLaserEnable.ValueBool()
	}

	tflog.Debug(ctx, "TribPTPResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"TribPTPResource: update ##: Error Create AC",
				"Create: Could not Marshal TribPTPResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ParentColId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" +  plan.Identifier.ParentColId.ValueString()  + "/tribPtps/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"TribPTPResource: update ##: Error update TribPTPResource",
				"Update: Could not update TribPTPResource, Identfier (DeviceID and ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "TribPTPResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"TribPTPResource: Create ##: Error Unmarshal response",
				"Update:Could not Create TribPTPResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "TribPTPResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "TribPTPResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *TribPTPResource) read(state *TribPTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read TribPTPResource",
			"TribPTPResource: Could not read. ParentId and Id and Href are not specified.",
		)
		return
	}

	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/tribPtp" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/tribPtp" + queryStr + "&q={\"state.tomAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/tribPtp" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/tribPtp/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/tribPtp" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"TribPTPResource/tribPtp: read ##: Error Read TribPTPResource/tribPtp",
			"Read:Could not get TribPTPResource/tribPtp, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "TribPTPResource/tribPtp: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"TribPTPResource/tribPtp: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal TribPTPResource/tribPtp, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"TribPTPResource/tribPtp: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "TribPTPResource: read ## ", map[string]interface{}{"plan": state})
}

func (r *TribPTPResource) delete(plan *TribPTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TribPTPResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (tribPtpData *TribPTPResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "TribPTPResourceData: populate ## ", map[string]interface{}{"plan": data})

	tribPtpData.Id = types.StringValue(data["id"].(string))
	tribPtpData.ColId = types.Int64Value(int64(data["colid"].(float64)))
	tribPtpData.Href = types.StringValue(data["href"].(string))
	tribPtpData.ParentId = types.StringValue(data["parentId"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if config != nil {
		if tribPtpData.Config == nil {
			tribPtpData.Config = &TribPTPConfig{}
		}
		for k, v := range config {
			switch k {
			case "name":
				if !tribPtpData.Config.Name.IsNull() {
					tribPtpData.Config.Name = types.StringValue(v.(string))
				}
			case "txLaserEnable":
				if !tribPtpData.Config.TxLaserEnable.IsNull() {
					tribPtpData.Config.TxLaserEnable = types.BoolValue(v.(bool))
				}
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		tribPtpData.State = types.ObjectValueMust(TribPTPStateAttributeType(), TribPTPStateAttributeValue(state))
	}

	tflog.Debug(ctx, "TribPTPResourceData: read ## ", map[string]interface{}{"tribPtpData": tribPtpData})
}

func TribPTPResourceSchemaAttributes() map[string]schema.Attribute {
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
				"tx_laser_enable": schema.BoolAttribute{
					Description: "amplifier_enable",
					Optional:    true,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: TribPTPStateAttributeType(),
		},
	}
}

func TribPTPObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: TribPTPAttributeType(),
	}
}

func TribPTPObjectsValue(data []interface{}) []attr.Value {
	tribPtps := []attr.Value{}
	for _, v := range data {
		tribPtp := v.(map[string]interface{})
		if tribPtp != nil {
			tribPtps = append(tribPtps, types.ObjectValueMust(
				TribPTPAttributeType(),
				TribPTPAttributeValue(tribPtp)))
		}
	}
	return tribPtps
}

func TribPTPAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":          types.StringType,
		"href":        types.StringType,
		"col_id":      types.Int64Type,
		"config":      types.ObjectType{AttrTypes: TribPTPConfigAttributeType()},
		"state":       types.ObjectType{AttrTypes: TribPTPStateAttributeType()},
	}
}

func TribPTPAttributeValue(tribPtp map[string]interface{}) map[string]attr.Value {
	colId := types.Int64Null()
	if tribPtp["colId"] != nil {
		colId = types.Int64Value(int64(tribPtp["colId"].(float64)))
	}
	href := types.StringNull()
	if tribPtp["href"] != nil {
		href = types.StringValue(tribPtp["href"].(string))
	}
	id := types.StringNull()
	if tribPtp["id"] != nil {
		id = types.StringValue(tribPtp["id"].(string))
	}
	parentId := types.StringNull()
	if tribPtp["parentId"] != nil {
		parentId = types.StringValue(tribPtp["parentId"].(string))
	}
	config := types.ObjectNull(TribPTPConfigAttributeType())
	if (tribPtp["config"]) != nil {
		config = types.ObjectValueMust(TribPTPConfigAttributeType(), TribPTPConfigAttributeValue(tribPtp["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(TribPTPStateAttributeType())
	if (tribPtp["state"]) != nil {
		state = types.ObjectValueMust(TribPTPStateAttributeType(), TribPTPStateAttributeValue(tribPtp["state"].(map[string]interface{})))
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
func TribPTPConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":            types.StringType,
		"tx_laser_enable": types.BoolType,
	}
}

func TribPTPConfigAttributeValue(tribPtpConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	txLaserEnable := types.BoolNull()

	for k, v := range tribPtpConfig {
		switch k {
		case "name":
			name = types.StringValue(v.(string))
		case "txLaserEnable":
			txLaserEnable = types.BoolValue(v.(bool))
		}
	}

	return map[string]attr.Value{
		"name":            name,
		"tx_laser_enable": txLaserEnable,
	}
}

func TribPTPStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":         types.StringType,
		"trib_ptp_aid":    types.StringType,
		"name":            types.StringType,
		"tx_laser_enable": types.BoolType,
		"lifecycle_state": types.StringType,
	}
}

func TribPTPStateAttributeValue(tribPtpState map[string]interface{}) map[string]attr.Value {
	parentAid := types.StringNull()
	tribPtpAid := types.StringNull()
	name := types.StringNull()
	txLaserEnable := types.BoolNull()
	lifecycleState := types.StringNull()

	for k, v := range tribPtpState {
		switch k {
		case "tribPtpAid":
			tribPtpAid = types.StringValue(v.(string))
		case "parentAid":
			parentAids := v.([]interface{})
			parentAid = types.StringValue(parentAids[0].(string))
		case "name":
			name = types.StringValue(v.(string))
		case "txLaserEnable":
			txLaserEnable = types.BoolValue(v.(bool))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		}
	}

	return map[string]attr.Value{
		"parent_aid":          parentAid,
		"trib_ptp_aid":     tribPtpAid,
		"name":             name,
		"tx_laser_enable":  txLaserEnable,
		"lifecycle_state": lifecycleState,
	}
}
