package moduleservice

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
	_ resource.Resource                = &ODUResource{}
	_ resource.ResourceWithConfigure   = &ODUResource{}
	_ resource.ResourceWithImportState = &ODUResource{}
)

// NewODUResource is a helper function to simplify the provider implementation.
func NewODUResource() resource.Resource {
	return &ODUResource{}
}

type ODUResource struct {
	client *ipm_pf.Client
}

type ODUDiagnostics struct {
	FacPRBSGen types.Bool `tfsdk:"fac_prbs_gen"`
	FacPRBSMon types.Bool `tfsdk:"fac_prbs_mon"`
}

type ODUConfig struct {
	Diagnostics ODUDiagnostics `tfsdk:"diagnostics"`
}

type ODUResourceData struct {
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href       types.String              `tfsdk:"href"`
	ColId      types.Int64               `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config     *ODUConfig                 `tfsdk:"config"`
	State      types.Object              `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *ODUResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_odu"
}

// Schema defines the schema for the data source.
func (r *ODUResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type ODUResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages ODU",
		Attributes:  ODUResourceSchemaAttributes(true),
	}
}

// Configure adds the provider configured client to the data source.
func (r *ODUResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r ODUResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ODUResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "ODUResource: Create - ", map[string]interface{}{"ODUResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r ODUResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ODUResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "ODUResource: Create - ", map[string]interface{}{"ODUResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r ODUResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ODUResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "ODUResource: Update", map[string]interface{}{"ODUResourceData": data})

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

func (r ODUResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ODUResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "ODUResource: Update", map[string]interface{}{"ODUResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ODUResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ODUResource) update(plan *ODUResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "ODUResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"ODUResource: Error update ODU",
			"ODUResource: Could not update ODU. Href, ModuleId, LinePtpColId, CarrierColId or ODU ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get TC config settings
	if !plan.Config.Diagnostics.FacPRBSGen.IsNull() || !plan.Config.Diagnostics.FacPRBSMon.IsNull() {
		diagnostics := make(map[string]interface{})
		if !plan.Config.Diagnostics.FacPRBSGen.IsNull() {
			diagnostics["facPRBSGen"] = plan.Config.Diagnostics.FacPRBSGen.ValueBool()
		}
		if !plan.Config.Diagnostics.FacPRBSMon.IsNull() {
			diagnostics["facPRBSMon"] = plan.Config.Diagnostics.FacPRBSMon.ValueBool()
		}
		updateRequest["diagnostics"] = diagnostics
	}

	tflog.Debug(ctx, "ODUResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"ODUResource: update ##: Error Create AC",
				"Create: Could not Marshal ODUResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ParentColId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/modules/" + plan.Identifier.DeviceId.ValueString() + "/otus/" +  plan.Identifier.ParentColId.ValueString()  + "/odus/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"ODUResource: update ##: Error update ODU",
				"Update: Could not update OTUResource, Identfier (DeviceID, parentColID or ColId) is not specified: ",
			)
			return
		}
		if err != nil {
			if !strings.Contains(err.Error(), "status: 202") {
				diags.AddError(
					"ODUResource: update ##: Error update ODUResource",
					"Create:Could not update ODUResource, unexpected error: "+err.Error(),
				)
				return
			}
		}

		tflog.Debug(ctx, "ODUResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"ODUResource: Create ##: Error Unmarshal response",
				"Update:Could not Create ODUResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "ODUResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "ODUResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *ODUResource) read(state *ODUResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && ((state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read ODUResource",
			"ODUResource: Could not read. Id, and Href, and identifiers are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "ODUResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/otus/" + state.Identifier.ParentColId.ValueString() + "/odus" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/otus/" + state.Identifier.ParentColId.ValueString() + "/odus" + queryStr + "&q={\"state.oduAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/otus/" + state.Identifier.ParentColId.ValueString() + "/odus" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/otus/" + state.Identifier.ParentColId.ValueString() + "/odus/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/modules" + state.Identifier.DeviceId.ValueString() + "/otus/" + state.Identifier.ParentColId.ValueString() + "/odus" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"ODUResource: read ##: Error Read ODUResource",
			"Read:Could not get ODUResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "ODUResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"ODUResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal ODUResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"ODUResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "ODUResource: read ## ", map[string]interface{}{"plan": state})
}

func (odu *ODUResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "ODUResourceData: populate ## ", map[string]interface{}{"plan": data})

	odu.Id = types.StringValue(data["id"].(string))
	odu.Href = types.StringValue(data["href"].(string))
	odu.ParentId = types.StringValue(data["parentId"].(string))
	odu.ColId = types.Int64Value(int64(data["colid"].(float64)))

	// populate config
	var config = data["config"].(map[string]interface{})
	if odu.Config == nil {
		odu.Config = &ODUConfig{}
	}
	for k, v := range config {
		switch k {
		case "diagnostics":
			diagnostics := v.(map[string]interface{})
			if diagnostics["facPRBSGen"] != nil && !odu.Config.Diagnostics.FacPRBSGen.IsNull() {
				odu.Config.Diagnostics.FacPRBSGen = types.BoolValue(v.(bool))
			}
			if diagnostics["facPRBSMon"] != nil && !odu.Config.Diagnostics.FacPRBSMon.IsNull() {
				odu.Config.Diagnostics.FacPRBSMon = types.BoolValue(v.(bool))
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		odu.State = types.ObjectValueMust(ODUStateAttributeType(), ODUStateAttributeValue(state))
	}

	tflog.Debug(ctx, "ODUResourceData: read ## ", map[string]interface{}{"plan": state})
}

func ODUResourceSchemaAttributes(read_only bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the ODU.",
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
		//Config           NetworkConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Description: "config",
			Computed:    read_only,
			Optional:    !read_only,
			Attributes: map[string]schema.Attribute{
				"diagnostics": schema.SingleNestedAttribute{
					Description: "diagnostics",
					Computed:    read_only,
					Optional:    !read_only,
					Attributes: map[string]schema.Attribute{
						"fac_prbs_gen": schema.BoolAttribute{
							Description: "fac_prbs_gen",
							Computed:    read_only,
							Optional:    !read_only,
						},
						"fac_prbs_mon": schema.BoolAttribute{
							Description: "fac_prbs_mon",
							Computed:    read_only,
							Optional:    !read_only,
						},
					},
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: ODUStateAttributeType(),
		},
	}
}

func ODUObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: ODUAttributeType(),
	}
}

func ODUObjectsValue(data []interface{}) []attr.Value {
	odus := []attr.Value{}
	for _, v := range data {
		odu := v.(map[string]interface{})
		if odu != nil {
			odus = append(odus, types.ObjectValueMust(
				ODUAttributeType(),
				ODUAttributeValue(odu)))
		}
	}
	return odus
}

func ODUAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":        types.StringType,
		"href":      types.StringType,
		"col_id":    types.Int64Type,
		"config":    types.ObjectType{AttrTypes: ODUConfigAttributeType()},
		"state":     types.ObjectType{AttrTypes: ODUStateAttributeType()},
	}
}

func ODUAttributeValue(odu map[string]interface{}) map[string]attr.Value {
	col_id := types.Int64Null()
	if odu["colId"] != nil {
		col_id = types.Int64Value(int64(odu["colId"].(float64)))
	}
	href := types.StringNull()
	if odu["href"] != nil {
		href = types.StringValue(odu["href"].(string))
	}
	parentId := types.StringNull()
	if odu["parentId"] != nil {
		parentId = types.StringValue(odu["parentId"].(string))
	}
	id := types.StringNull()
	if odu["id"] != nil {
		id = types.StringValue(odu["id"].(string))
	}
	config := types.ObjectNull(ODUConfigAttributeType())
	if (odu["config"]) != nil {
		config = types.ObjectValueMust(ODUConfigAttributeType(), ODUConfigAttributeValue(odu["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(ODUStateAttributeType())
	if (odu["state"]) != nil {
		state = types.ObjectValueMust(ODUStateAttributeType(), ODUStateAttributeValue(odu["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"col_id":    col_id,
		"parent_id": parentId,
		"id":        id,
		"href":      href,
		"config":    config,
		"state":     state,
	}
}

func ODUDiagnosticsAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"fac_prbs_gen": types.BoolType,
		"fac_prbs_mon": types.BoolType,
	}
}

func ODUDiagnosticsAttributeValue(diagnostics map[string]interface{}) map[string]attr.Value {
	facPRBSGen := types.BoolNull()
	if diagnostics["facPRBSGen"] != nil {
		facPRBSGen = types.BoolValue(diagnostics["facPRBSGen"].(bool))
	}
	facPRBSMon := types.BoolNull()
	if diagnostics["facPRBSMon"] != nil {
		facPRBSGen = types.BoolValue(diagnostics["facPRBSMon"].(bool))
	}

	return map[string]attr.Value{
		"fac_prbs_gen": facPRBSGen,
		"fac_prbs_mon": facPRBSMon,
	}
}

func ODUConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"diagnostics": types.ObjectType{AttrTypes: ODUDiagnosticsAttributeType()},
	}
}

func ODUConfigAttributeValue(oduConfig map[string]interface{}) map[string]attr.Value {
	diagnostics := types.ObjectNull(ODUDiagnosticsAttributeType())
	if (oduConfig["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(ODUDiagnosticsAttributeType(), ODUDiagnosticsAttributeValue(oduConfig["diagnostics"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"diagnostics": diagnostics,
	}
}

func ODUStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":       types.StringType,
		"odu_aid":          types.StringType,
		"odu_type":         types.StringType,
		"diagnostics":      types.ObjectType{AttrTypes: ODUDiagnosticsAttributeType()},
		"life_cycle_state": types.StringType,
	}
}

func ODUStateAttributeValue(oduState map[string]interface{}) map[string]attr.Value {
	oduAid := types.StringNull()
	if oduState["oduAid"] != nil {
		oduAid = types.StringValue(oduState["oduAid"].(string))
	}
	parentAid := types.StringNull()
	if oduState["parentAid"] != nil {
		parentAids := oduState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	oduType := types.StringNull()
	if oduState["oduType"] != nil {
		oduType = types.StringValue(oduState["oduType"].(string))
	}

	diagnostics := types.ObjectNull(ODUDiagnosticsAttributeType())
	if (oduState["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(ODUDiagnosticsAttributeType(), ODUDiagnosticsAttributeValue(oduState["diagnostics"].(map[string]interface{})))
	}
	lifecycleState := types.StringNull()
	if oduState["lifecycleState"] != nil {
		lifecycleState = types.StringValue(oduState["lifecycleState"].(string))
	}

	return map[string]attr.Value{
		"parent_aid":       parentAid,
		"odu_aid":          oduAid,
		"odu_type":         oduType,
		"diagnostics":      diagnostics,
		"life_cycle_state": lifecycleState,
	}
}
