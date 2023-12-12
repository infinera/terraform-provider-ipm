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
	_ resource.Resource                = &PolPTPResource{}
	_ resource.ResourceWithConfigure   = &PolPTPResource{}
	_ resource.ResourceWithImportState = &PolPTPResource{}
)

// NewPolPTPResource is a helper function to simplify the provider implementation.
func NewPolPTPResource() resource.Resource {
	return &PolPTPResource{}
}

type PolPTPResource struct {
	client *ipm_pf.Client
}

type PolPTPConfig struct {
	Name types.String `tfsdk:"name"`
}

type PolPTPResourceData struct {
	Id         types.String  `tfsdk:"id"`
	ParentId   types.String  `tfsdk:"parent_id"`
	Href       types.String  `tfsdk:"href"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	ColId      types.Int64   `tfsdk:"colid"`
	Config    *PolPTPConfig `tfsdk:"config"`
	State     types.Object  `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *PolPTPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_polPtp"
}

// Schema defines the schema for the data source.
func (r *PolPTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type PolPTPResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages NDU PolPTP",
		Attributes:  PolPTPResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *PolPTPResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r PolPTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PolPTPResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "PolPTPResource: Create - ", map[string]interface{}{"PolPTPResourceData": data})

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

func (r PolPTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PolPTPResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "PolPTPResource: Create - ", map[string]interface{}{"PolPTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r PolPTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PolPTPResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "PolPTPResource: Update", map[string]interface{}{"PolPTPResourceData": data})

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

func (r PolPTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PolPTPResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "PolPTPResource: Update", map[string]interface{}{"PolPTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *PolPTPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PolPTPResource) update(plan *PolPTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "PolPTPResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"PolPTPResource: Error update PolPTP",
			"PolPTPResource: Could not update PolPTP. Href, NDUId, PortColId, LinePTPColId, PolPTP ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})
	// get TC config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}

	tflog.Debug(ctx, "PolPTPResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"PolPTPResource: update ##: Error Create AC",
				"Create: Could not Marshal PolPTPResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ParentColId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" +  plan.Identifier.ParentColId.ValueString()  + "/polPtps/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"PolPTPResource: update ##: Error update PolPTPResource",
				"Update: Could not update PolPTPResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "PolPTPResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"PolPTPResource: Create ##: Error Unmarshal response",
				"Update:Could not Create PolPTPResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "PolPTPResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "PolPTPResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *PolPTPResource) read(state *PolPTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read PolPTPResource",
			"PolPTPResource: Could not read. ParentId and Id and Href are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "PolPTPResource: read ## ", map[string]interface{}{"plan": state})

	tflog.Debug(ctx, "LinePTPResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/polPtps" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/polPtps" + queryStr + "&q={\"state.edfaAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/polPtps" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/polPtps/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/polPtps" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"LinePTPResource: read ##: Error Read LinePTPResource",
			"Read:Could not get LinePTPResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "LinePTPResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"LinePTPResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal LinePTPResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"LinePTPResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "PolPTPResource: read ## ", map[string]interface{}{"plan": state})
}

func (r *PolPTPResource) delete(plan *PolPTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "PolPTPResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (polPtpData *PolPTPResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "PolPTPResourceData: populate ## ", map[string]interface{}{"plan": data})

	polPtpData.Id = types.StringValue(data["id"].(string))
	polPtpData.ColId = types.Int64Value(int64(data["colid"].(float64)))
	polPtpData.Href = types.StringValue(data["href"].(string))
	polPtpData.ParentId = types.StringValue(data["parentId"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if config != nil {
		if polPtpData.Config == nil {
			polPtpData.Config = &PolPTPConfig{}
		}
		for k, v := range config {
			switch k {
			case "name":
				if !polPtpData.Config.Name.IsNull() {
					polPtpData.Config.Name = types.StringValue(v.(string))
				}
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		polPtpData.State = types.ObjectValueMust(PolPTPStateAttributeType(), PolPTPStateAttributeValue(state))
	}

	tflog.Debug(ctx, "PolPTPResourceData: read ## ", map[string]interface{}{"polPtpData": polPtpData})
}

func PolPTPResourceSchemaAttributes() map[string]schema.Attribute {
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
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: PolPTPStateAttributeType(),
		},
	}
}

func PolPTPObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: PolPTPAttributeType(),
	}
}

func PolPTPObjectsValue(data []interface{}) []attr.Value {
	polPtps := []attr.Value{}
	for _, v := range data {
		polPtp := v.(map[string]interface{})
		if polPtp != nil {
			polPtps = append(polPtps, types.ObjectValueMust(
				PolPTPAttributeType(),
				PolPTPAttributeValue(polPtp)))
		}
	}
	return polPtps
}

func PolPTPAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":          types.StringType,
		"href":        types.StringType,
		"col_id":      types.Int64Type,
		"config":      types.ObjectType{AttrTypes: PolPTPConfigAttributeType()},
		"state":       types.ObjectType{AttrTypes: PolPTPStateAttributeType()},
	}
}

func PolPTPAttributeValue(polPtp map[string]interface{}) map[string]attr.Value {
	colId := types.Int64Null()
	if polPtp["colId"] != nil {
		colId = types.Int64Value(int64(polPtp["colId"].(float64)))
	}
	href := types.StringNull()
	if polPtp["href"] != nil {
		href = types.StringValue(polPtp["href"].(string))
	}
	id := types.StringNull()
	if polPtp["id"] != nil {
		id = types.StringValue(polPtp["id"].(string))
	}
	parentId := types.StringNull()
	if polPtp["parentId"] != nil {
		parentId = types.StringValue(polPtp["parentId"].(string))
	}
	config := types.ObjectNull(PolPTPConfigAttributeType())
	if (polPtp["config"]) != nil {
		config = types.ObjectValueMust(PolPTPConfigAttributeType(), PolPTPConfigAttributeValue(polPtp["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(PolPTPStateAttributeType())
	if (polPtp["state"]) != nil {
		state = types.ObjectValueMust(PolPTPStateAttributeType(), PolPTPStateAttributeValue(polPtp["state"].(map[string]interface{})))
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

func PolPTPConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":              types.StringType,
		"required_type":     types.StringType,
		"required_sub_type": types.StringType,
		"amplifier_enable":  types.BoolType,
		"gain_target":       types.Int64Type,
	}
}

func PolPTPConfigAttributeValue(polPtpConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()

	for k, v := range polPtpConfig {
		switch k {
		case "name":
			name = types.StringValue(v.(string))
		}
	}

	return map[string]attr.Value{
		"name": name,
	}
}

func PolPTPStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":            types.StringType,
		"parent_aid":      types.StringType,
		"pol_ptp_aid":     types.StringType,
		"lifecycle_state": types.StringType,
	}
}

func PolPTPStateAttributeValue(polPtpState map[string]interface{}) map[string]attr.Value {
	polPtpAid := types.StringNull()
	name := types.StringNull()
	lifecycleState := types.StringNull()
	parentAid := types.StringNull()

	for k, v := range polPtpState {
		switch k {
		case "parentAid":
			parentAids := v.([]interface{})
			parentAid = types.StringValue(parentAids[0].(string))
		case "polPtpAid":
			polPtpAid = types.StringValue(v.(string))
		case "name":
			name = types.StringValue(v.(string))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		}
	}

	return map[string]attr.Value{
		"parent_aid":          parentAid,
		"pol_ptp_aid":     polPtpAid,
		"name":            name,
		"lifecycle_state": lifecycleState,
	}
}
