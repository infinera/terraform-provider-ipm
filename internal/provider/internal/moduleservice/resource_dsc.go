package moduleservice

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
	_ resource.Resource                = &DSCResource{}
	_ resource.ResourceWithConfigure   = &DSCResource{}
	_ resource.ResourceWithImportState = &DSCResource{}
)

// NewDSCResource is a helper function to simplify the provider implementation.
func NewDSCResource() resource.Resource {
	return &DSCResource{}
}

type DSCResource struct {
	client *ipm_pf.Client
}

type DSCDiagnostics struct {
	FacPRBSGen types.Bool `tfsdk:"fac_prbs_gen"`
	FacPRBSMon types.Bool `tfsdk:"fac_prbs_mon"`
}

type DSCConfig struct {
	RelativeDPO types.Int64    `tfsdk:"relative_dpo"`
	Diagnostics DSCDiagnostics `tfsdk:"diagnostics"`
}

type DSCResourceData struct {
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href       types.String              `tfsdk:"href"`
	ColId      types.Int64               `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config     *DSCConfig                 `tfsdk:"config"`
	State      types.Object              `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *DSCResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dsc"
}

// Schema defines the schema for the data source.
func (r *DSCResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type DSCResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages DSC",
		Attributes:  DSCResourceSchemaAttributes(true),
	}
}

// Configure adds the provider configured client to the data source.
func (r *DSCResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r DSCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DSCResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "DSCResource: Create - ", map[string]interface{}{"DSCResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r DSCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DSCResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "DSCResource: Create - ", map[string]interface{}{"DSCResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r DSCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DSCResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "DSCResource: Update", map[string]interface{}{"DSCResourceData": data})

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

func (r DSCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DSCResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "DSCResource: Update", map[string]interface{}{"DSCResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *DSCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DSCResource) update(plan *DSCResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "DSCResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.GrandParentColId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"DSCResource: Error update DSC",
			"DSCResource: Could not update DSC. Href and ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get TC config settings
	if !plan.Config.RelativeDPO.IsNull() {
		updateRequest["relativeDPO"] = plan.Config.RelativeDPO.ValueInt64()
	}
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

	tflog.Debug(ctx, "DSCResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"DSCResource: update ##: Error Create AC",
				"Create: Could not Marshal DSCResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ParentColId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/modules/" + plan.Identifier.DeviceId.ValueString() + "/linePtps/" +  plan.Identifier.GrandParentColId.ValueString()  + "/carriers/" +  plan.Identifier.ParentColId.ValueString() + "/dscs/" + plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"DSCResource: update ##: Error update DSC",
				"Update: Could not update DSCResource, Identfier (DeviceID, grand parent COLID, parentColID or ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "DSCResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"DSCResource: Create ##: Error Unmarshal response",
				"Update:Could not Create DSCResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "DSCResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "DSCResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *DSCResource) read(state *DSCResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull() || state.Identifier.GrandParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read DSCResource",
			"DSCResource: Could not read. Id, and Href, and identifiers are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "DSCResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.GrandParentColId.ValueString() + "/carriers/" + state.Identifier.ParentColId.ValueString() + "/dscs" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.GrandParentColId.ValueString() + "/carriers/" + state.Identifier.ParentColId.ValueString() + "/dscs" + queryStr + "&q={\"state.dscAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.GrandParentColId.ValueString() + "/carriers/" + state.Identifier.ParentColId.ValueString() + "/dscs" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.GrandParentColId.ValueString() + "/carriers/" + state.Identifier.ParentColId.ValueString() + "/dscgs/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/modules" + state.Identifier.DeviceId.ValueString() + "/linePtps" + state.Identifier.ParentColId.ValueString() + "/carriers" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"DSCResource: read ##: Error Read DSCResource",
			"Read:Could not get DSCResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "DSCResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"DSCResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal DSCResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"DSCResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "DSCResource: read ## ", map[string]interface{}{"plan": state})
}

func (dsc *DSCResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "DSCResourceData: populate ## ", map[string]interface{}{"plan": data})

	dsc.Id = types.StringValue(data["id"].(string))
	dsc.Href = types.StringValue(data["href"].(string))
	dsc.ParentId = types.StringValue(data["parentId"].(string))
	dsc.ColId = types.Int64Value(int64(data["colId"].(float64)))

	// populate config
	var config = data["config"].(map[string]interface{})
	if dsc.Config == nil {
		dsc.Config = &DSCConfig{}
	}
	for k, v := range config {
		switch k {
		case "relativeDPO":
			if !dsc.Config.RelativeDPO.IsNull() {
				dsc.Config.RelativeDPO = types.Int64Value(int64(v.(float64)))
			}
		case "diagnostics":
			diagnostics := v.(map[string]interface{})
			if diagnostics["facPRBSGen"] != nil && !dsc.Config.Diagnostics.FacPRBSGen.IsNull() {
				dsc.Config.Diagnostics.FacPRBSGen = types.BoolValue(v.(bool))
			}
			if diagnostics["facPRBSMon"] != nil && !dsc.Config.Diagnostics.FacPRBSMon.IsNull() {
				dsc.Config.Diagnostics.FacPRBSMon = types.BoolValue(v.(bool))
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		dsc.State = types.ObjectValueMust(DSCStateAttributeType(), DSCStateAttributeValue(state))
	}

	tflog.Debug(ctx, "DSCResourceData: read ## ", map[string]interface{}{"plan": state})
}

func DSCResourceSchemaAttributes(read_only bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the DSC.",
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
				"relative_dpo": schema.Int64Attribute{
					Description: "relative DPO",
					Computed:    read_only,
					Optional:    !read_only,
				},
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
			AttributeTypes: DSCStateAttributeType(),
		},
	}
}

func DSCObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: DSCAttributeType(),
	}
}

func DSCObjectsValue(data []interface{}) []attr.Value {
	dscs := []attr.Value{}
	for _, v := range data {
		dsc := v.(map[string]interface{})
		if dsc != nil {
			dscs = append(dscs, types.ObjectValueMust(
				DSCAttributeType(),
				DSCAttributeValue(dsc)))
		}
	}
	return dscs
}

func DSCAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":        types.StringType,
		"href":      types.StringType,
		"col_id":    types.Int64Type,
		"config":    types.ObjectType{AttrTypes: DSCConfigAttributeType()},
		"state":     types.ObjectType{AttrTypes: DSCStateAttributeType()},
	}
}

func DSCAttributeValue(dsc map[string]interface{}) map[string]attr.Value {
	col_id := types.Int64Null()
	if dsc["colId"] != nil {
		col_id = types.Int64Value(int64(dsc["colId"].(float64)))
	}
	href := types.StringNull()
	if dsc["href"] != nil {
		href = types.StringValue(dsc["href"].(string))
	}
	parentId := types.StringNull()
	if dsc["parentId"] != nil {
		parentId = types.StringValue(dsc["parentId"].(string))
	}
	id := types.StringNull()
	if dsc["id"] != nil {
		id = types.StringValue(dsc["id"].(string))
	}
	config := types.ObjectNull(DSCConfigAttributeType())
	if (dsc["config"]) != nil {
		config = types.ObjectValueMust(DSCConfigAttributeType(), DSCConfigAttributeValue(dsc["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(DSCStateAttributeType())
	if (dsc["state"]) != nil {
		state = types.ObjectValueMust(DSCStateAttributeType(), DSCStateAttributeValue(dsc["state"].(map[string]interface{})))
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

func DSCDiagnosticsAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"fac_prbs_gen": types.BoolType,
		"fac_prbs_mon": types.BoolType,
	}
}

func DSCDiagnosticsAttributeValue(diagnostics map[string]interface{}) map[string]attr.Value {
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

func DSCConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"relative_dpo": types.Int64Type,
		"diagnostics":  types.ObjectType{AttrTypes: DSCDiagnosticsAttributeType()},
	}
}

func DSCConfigAttributeValue(dscConfig map[string]interface{}) map[string]attr.Value {
	relativeDPO := types.Int64Null()
	if dscConfig["relativeDPO"] != nil {
		relativeDPO = types.Int64Value(int64(dscConfig["relativeDPO"].(float64)))
	}
	diagnostics := types.ObjectNull(DSCDiagnosticsAttributeType())
	if (dscConfig["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(DSCDiagnosticsAttributeType(), DSCDiagnosticsAttributeValue(dscConfig["diagnostics"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"relative_dpo": relativeDPO,
		"diagnostics":  diagnostics,
	}
}

func DSCStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":       types.StringType,
		"dsc_aid":          types.StringType,
		"cdsc":             types.Int64Type,
		"tx_status":        types.StringType,
		"rx_status":        types.StringType,
		"relative_dpo":     types.Int64Type,
		"diagnostics":      types.ObjectType{AttrTypes: DSCDiagnosticsAttributeType()},
		"life_cycle_state": types.StringType,
	}
}

func DSCStateAttributeValue(dscState map[string]interface{}) map[string]attr.Value {
	dscAid := types.StringNull()
	if dscState["dscAid"] != nil {
		dscAid = types.StringValue(dscState["dscAid"].(string))
	}
	parentAid := types.StringNull()
	if dscState["parentAid"] != nil {
		parentAids := dscState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	cDsc := types.Int64Null()
	if dscState["cDsc"] != nil {
		cDsc = types.Int64Value(int64(dscState["cDsc"].(float64)))
	}
	txStatus := types.StringNull()
	if dscState["txStatus"] != nil {
		txStatus = types.StringValue(dscState["txStatus"].(string))
	}
	rxStatus := types.StringNull()
	if dscState["rxStatus"] != nil {
		rxStatus = types.StringValue(dscState["rxStatus"].(string))
	}
	relativeDPO := types.Int64Null()
	if dscState["relativeDPO"] != nil {
		relativeDPO = types.Int64Value(int64(dscState["relativeDPO"].(float64)))
	}
	diagnostics := types.ObjectNull(DSCDiagnosticsAttributeType())
	if (dscState["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(DSCDiagnosticsAttributeType(), DSCDiagnosticsAttributeValue(dscState["diagnostics"].(map[string]interface{})))
	}
	lifecycleState := types.StringNull()
	if dscState["lifecycleState"] != nil {
		lifecycleState = types.StringValue(dscState["lifecycleState"].(string))
	}

	return map[string]attr.Value{
		"parent_aid":       parentAid,
		"dsc_aid":          dscAid,
		"cdsc":             cDsc,
		"tx_status":        txStatus,
		"rx_status":        rxStatus,
		"relative_dpo":     relativeDPO,
		"diagnostics":      diagnostics,
		"life_cycle_state": lifecycleState,
	}
}
