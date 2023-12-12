package moduleservice

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
	_ resource.Resource                = &OTUResource{}
	_ resource.ResourceWithConfigure   = &OTUResource{}
	_ resource.ResourceWithImportState = &OTUResource{}
)

// NewOTUResource is a helper function to simplify the provider implementation.
func NewOTUResource() resource.Resource {
	return &OTUResource{}
}

type OTUResource struct {
	client *ipm_pf.Client
}
type OTUDiagnostics struct {
	TermLB         types.String `tfsdk:"term_lb"`
	TermLBDuration types.Int64  `tfsdk:"term_lb_duration"`
}

type OTUConfig struct {
	Diagnostics OTUDiagnostics `tfsdk:"diagnostics"`
}

type OTUResourceData struct {
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href       types.String              `tfsdk:"href"`
	ColId      types.Int64               `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config     *OTUConfig                 `tfsdk:"config"`
	State      types.Object              `tfsdk:"state"`
	ODUs       types.List                `tfsdk:"odus"`
}

// Metadata returns the data source type name.
func (r *OTUResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_otu"
}

// Schema defines the schema for the data source.
func (r *OTUResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type OTUResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an ModuleLinePTP",
		Attributes:  OTUResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *OTUResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r OTUResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OTUResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "OTUResource: Create - ", map[string]interface{}{"OTUResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r OTUResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OTUResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "OTUResource: Create - ", map[string]interface{}{"OTUResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r OTUResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OTUResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "OTUResource: Update", map[string]interface{}{"OTUResourceData": data})

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

func (r OTUResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OTUResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "OTUResource: Update", map[string]interface{}{"OTUResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *OTUResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}


func (r *OTUResource) update(plan *OTUResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "OTUResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"OTUResource: Error update OTU",
			"OTUResource: Could not update OTU. Href, ModuleId, or OTU ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	diagnostics := make(map[string]interface{})
	if !plan.Config.Diagnostics.TermLB.IsNull() {
		diagnostics["termLB"] = plan.Config.Diagnostics.TermLB.ValueString()
	}
	if !plan.Config.Diagnostics.TermLBDuration.IsNull() {
		diagnostics["termLBDuration"] = plan.Config.Diagnostics.TermLBDuration.ValueInt64()
	}
	if len(diagnostics) > 0 {
		updateRequest["diagnostics"] = diagnostics
	}

	tflog.Debug(ctx, "OTUResource: update ## ", map[string]interface{}{"Create Request": updateRequest})
	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"OTUResource: update ##: Error Create AC",
				"Create: Could not Marshal OTUResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.ColId.IsNull() && !plan.ParentId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/modules/" + plan.ParentId.ValueString() + "/otus/" +  strconv.FormatInt(plan.ColId.ValueInt64(),10), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/modules/" + plan.Identifier.DeviceId.ValueString() + "/otus/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"OTUResource: update ##: Error update OTU",
				"Update: Could not update OTUResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}
		if err != nil {
			if !strings.Contains(err.Error(), "status: 202") {
				diags.AddError(
					"OTUResource: update ##: Error update OTUResource",
					"Create:Could not update OTUResource, unexpected error: "+err.Error(),
				)
				return
			}
		}

		tflog.Debug(ctx, "OTUResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"OTUResource: Create ##: Error Unmarshal response",
				"Update:Could not Create OTUResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "OTUResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "OTUResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *OTUResource) read(state *OTUResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && state.Identifier.ColId.IsNull() && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull() {
		diags.AddError(
			"Error Read OTUResource",
			"OTUResource: Could not read. Id, and Href, and identifiers are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "OTUResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/otus" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/otus" + queryStr + "&q={\"state.otuAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/otus" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/otus/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		diags.AddError(
			"Error Read OTUResource",
			"OTUResource: Could not read. Id, and Href, and identifiers are not specified.",
		)
		return
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"OTUResource: read ##: Error Read OTUResource",
			"Read:Could not get OTUResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "OTUResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"OTUResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal OTUResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"OTUResource: read ##: Can not get Module",
				"Read:Could not get OTU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "OTUResource: read ## ", map[string]interface{}{"plan": state})
}

func (otuData *OTUResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "OTUResourceData: populate ## ", map[string]interface{}{"plan": data})

	otuData.Id = types.StringValue(data["id"].(string))
	otuData.ParentId = types.StringValue(data["parentId"].(string))
	otuData.Href = types.StringValue(data["href"].(string))
	otuData.ColId = types.Int64Value(int64(data["colid"].(float64)))

	// populate config
	var config = data["config"].(map[string]interface{})
	if otuData.Config == nil {
		otuData.Config = &OTUConfig{}
	}
	for k, v := range config {
		switch k {
		case "diagnostics":
			diagnostics := v.(map[string]interface{})
			if diagnostics["termLB"] != nil && !otuData.Config.Diagnostics.TermLB.IsNull() {
				otuData.Config.Diagnostics.TermLB = types.StringValue(v.(string))
			}
			if diagnostics["termLBDuration"] != nil && !otuData.Config.Diagnostics.TermLBDuration.IsNull() {
				otuData.Config.Diagnostics.TermLBDuration = types.Int64Value(int64(v.(float64)))
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		otuData.State = types.ObjectValueMust(OTUStateAttributeType(), OTUStateAttributeValue(state))
	}
	// populate odus
	otuData.ODUs = types.ListNull(ODUObjectType())
	if data["odus"] != nil {
		otuData.ODUs = types.ListValueMust(ODUObjectType(), ODUObjectsValue(data["odus"].([]interface{})))
	}

	tflog.Debug(ctx, "OTUResourceData: POPULATE ## ", map[string]interface{}{"otuData": otuData})
}

func OTUResourceSchemaAttributes(computeEntity_optional ...bool) map[string]schema.Attribute {
	computeFlag := false
	optionalFlag := true
	if len(computeEntity_optional) > 0 {
		computeFlag = computeEntity_optional[0]
		optionalFlag = !computeFlag
	}
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Module.",
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
			Computed:    computeFlag,
			Optional:    optionalFlag,
			Attributes: map[string]schema.Attribute{
				"diagnostics": schema.SingleNestedAttribute{
					Description: "diagnostics",
					Computed:    computeFlag,
					Optional:    optionalFlag,
					Attributes: map[string]schema.Attribute{
						"term_lb": schema.StringAttribute{
							Description: "term_lb",
							Computed:    computeFlag,
							Optional:    optionalFlag,
						},
						"term_lb_duration": schema.Int64Attribute{
							Description: "term_lb_duration",
							Computed:    computeFlag,
							Optional:    optionalFlag,
						},
					},
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: OTUStateAttributeType(),
		},
		"odus": schema.ListAttribute{
			Computed:    true,
			ElementType: OTUObjectType(),
		},
	}
}

func OTUObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: OTUAttributeType(),
	}
}

func OTUObjectsValue(data []interface{}) []attr.Value {
	otus := []attr.Value{}
	for _, v := range data {
		otu := v.(map[string]interface{})
		if otu != nil {
			otus = append(otus, types.ObjectValueMust(
				OTUAttributeType(),
				OTUAttributeValue(otu)))
		}
	}
	return otus
}

func OTUAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":        types.StringType,
		"href":      types.StringType,
		"col_id":    types.Int64Type,
		"config":    types.ObjectType{AttrTypes: OTUConfigAttributeType()},
		"state":     types.ObjectType{AttrTypes: OTUStateAttributeType()},
		"odus":      types.ListType{ElemType: ODUObjectType()},
	}
}

func OTUAttributeValue(otu map[string]interface{}) map[string]attr.Value {
	col_id := types.Int64Null()
	if otu["colId"] != nil {
		col_id = types.Int64Value(int64(otu["colId"].(float64)))
	}
	href := types.StringNull()
	if otu["href"] != nil {
		href = types.StringValue(otu["href"].(string))
	}
	parentId := types.StringNull()
	if otu["parentId"] != nil {
		parentId = types.StringValue(otu["parentId"].(string))
	}
	id := types.StringNull()
	if otu["id"] != nil {
		id = types.StringValue(otu["id"].(string))
	}
	config := types.ObjectNull(OTUConfigAttributeType())
	if (otu["config"]) != nil {
		config = types.ObjectValueMust(OTUConfigAttributeType(), OTUConfigAttributeValue(otu["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(OTUStateAttributeType())
	if (otu["state"]) != nil {
		state = types.ObjectValueMust(OTUStateAttributeType(), OTUStateAttributeValue(otu["state"].(map[string]interface{})))
	}
	odus := types.ListNull(ODUObjectType())
	if (otu["odus"]) != nil {
		odus = types.ListValueMust(ODUObjectType(), ODUObjectsValue(otu["odus"].([]interface{})))
	}

	return map[string]attr.Value{
		"col_id":    col_id,
		"parent_id": parentId,
		"id":        id,
		"href":      href,
		"config":    config,
		"state":     state,
		"odus":      odus,
	}
}
func OTUConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"diagnostics": types.ObjectType{AttrTypes: OTUDiagnosticsAttributeType()},
	}
}

func OTUConfigAttributeValue(otuConfig map[string]interface{}) map[string]attr.Value {
	diagnostics := types.ObjectNull(OTUDiagnosticsAttributeType())
	if (otuConfig["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(OTUDiagnosticsAttributeType(), OTUDiagnosticsAttributeValue(otuConfig["diagnostics"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"diagnostics": diagnostics,
	}
}

func OTUStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":       types.StringType,
		"otu_aid":          types.StringType,
		"otu_type":         types.StringType,
		"rate":             types.Int64Type,
		"diagnostics":      types.ObjectType{AttrTypes: OTUDiagnosticsAttributeType()},
		"life_cycle_state": types.StringType,
	}
}

func OTUStateAttributeValue(otuState map[string]interface{}) map[string]attr.Value {
	parentAid := types.StringNull()
	if otuState["parentAid"] != nil {
		parentAids := otuState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	otuAid := types.StringNull()
	if otuState["otuAid"] != nil {
		otuAid = types.StringValue(otuState["otuAid"].(string))
	}
	otuType := types.StringNull()
	if otuState["otuType"] != nil {
		otuType = types.StringValue(otuState["otuType"].(string))
	}

	rate := types.Int64Null()
	if otuState["rate"] != nil {
		rate = types.Int64Value(int64(otuState["rate"].(float64)))
	}
	diagnostics := types.ObjectNull(OTUDiagnosticsAttributeType())
	if (otuState["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(OTUDiagnosticsAttributeType(), OTUDiagnosticsAttributeValue(otuState["diagnostics"].(map[string]interface{})))
	}
	lifecycleState := types.StringNull()
	if otuState["lifecycleState"] != nil {
		lifecycleState = types.StringValue(otuState["lifecycleState"].(string))
	}

	return map[string]attr.Value{
		"parent_aid":       parentAid,
		"otu_aid":          otuAid,
		"otu_type":         otuType,
		"rate":             rate,
		"diagnostics":      diagnostics,
		"life_cycle_state": lifecycleState,
	}
}

func OTUDiagnosticsAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"term_lb":          types.StringType,
		"term_lb_duration": types.Int64Type,
	}
}

func OTUDiagnosticsAttributeValue(diagnostics map[string]interface{}) map[string]attr.Value {
	termLB := types.StringNull()
	if diagnostics["termLB"] != nil {
		termLB = types.StringValue(diagnostics["termLB"].(string))
	}
	termLBDuration := types.Int64Null()
	if diagnostics["termLBDuration"] != nil {
		termLBDuration = types.Int64Value(int64(diagnostics["termLBDuration"].(float64)))
	}

	return map[string]attr.Value{
		"term_lb":          termLB,
		"term_lb_duration": termLBDuration,
	}
}
