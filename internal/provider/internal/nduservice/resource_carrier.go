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
	_ resource.Resource                = &CarrierResource{}
	_ resource.ResourceWithConfigure   = &CarrierResource{}
	_ resource.ResourceWithImportState = &CarrierResource{}
)

// NewCarrierResource is a helper function to simplify the provider implementation.
func NewCarrierResource() resource.Resource {
	return &CarrierResource{}
}

type CarrierResource struct {
	client *ipm_pf.Client
}

type CarrierDiagnostics struct {
	TermLB         types.String `tfsdk:"term_lb"`
	TermLBDuration types.Int64  `tfsdk:"term_lb_duration"`
	FacLB         types.String `tfsdk:"fac_lb"`
	FacLBDuration types.Int64  `tfsdk:"fac_lb_duration"`
}

type CarrierConfig struct {
	Name        types.String       `tfsdk:"name"`
	Frequency   types.Int64        `tfsdk:"frequency"`
	Modulation  types.String       `tfsdk:"modulation"`
	TxCLPtarget types.Int64        `tfsdk:"tx_clp_target"`
	Diagnostics CarrierDiagnostics `tfsdk:"diagnostics"`
}

type CarrierResourceData struct {
	Identifier   common.ResourceIdentifier `tfsdk:"identifier"`
	Id           types.String   `tfsdk:"id"`
	ParentId     types.String   `tfsdk:"parent_id"`
	Href         types.String   `tfsdk:"href"`
	ColId        types.Int64    `tfsdk:"col_id"`
	Config       *CarrierConfig `tfsdk:"config"`
	State        types.Object   `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *CarrierResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_carrier"
}

// Schema defines the schema for the data source.
func (r *CarrierResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type CarrierResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages NDU Carrier",
		Attributes:  CarrierResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *CarrierResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r CarrierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CarrierResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "CarrierResource: Create - ", map[string]interface{}{"CarrierResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r CarrierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CarrierResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "CarrierResource: Create - ", map[string]interface{}{"CarrierResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r CarrierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CarrierResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "CarrierResource: Update", map[string]interface{}{"CarrierResourceData": data})

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

func (r CarrierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CarrierResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "CarrierResource: Update", map[string]interface{}{"CarrierResourceData": data})

	resp.Diagnostics.Append(diags...)


	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *CarrierResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *CarrierResource) update(plan *CarrierResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "CarrierResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.GrandParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"CarrierResource: Error update Carrier",
			"CarrierResource: Could not update Carrier. Href, NDUId, PortColId, LinePTPColId, Carrier ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get TC config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.Frequency.IsNull() {
		updateRequest["frequency"] = plan.Config.Frequency.ValueInt64()
	}
	if !plan.Config.Modulation.IsNull() {
		updateRequest["modulation"] = plan.Config.Modulation.ValueString()
	}
	if !plan.Config.TxCLPtarget.IsNull() {
		updateRequest["txCLPtarget"] = plan.Config.TxCLPtarget.ValueInt64()
	}

	diagnostics := make(map[string]interface{})
	if !plan.Config.Diagnostics.TermLB.IsNull() {
		diagnostics["termLB"] = plan.Config.Diagnostics.TermLB.ValueString()
	}
	if !plan.Config.Diagnostics.TermLBDuration.IsNull() {
		diagnostics["termLBDuration"] = plan.Config.Diagnostics.TermLBDuration.ValueInt64()
	}
	if !plan.Config.Diagnostics.FacPRBSGen.IsNull() {
		diagnostics["facPRBSGen"] = plan.Config.Diagnostics.FacPRBSGen.ValueBool()
	}
	if !plan.Config.Diagnostics.FacPRBSMon.IsNull() {
		diagnostics["facPRBSMon"] = plan.Config.Diagnostics.FacPRBSMon.ValueBool()
	}
	if len(diagnostics) > 0 {
		updateRequest["diagnostics"] = diagnostics
	}

	tflog.Debug(ctx, "CarrierResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"CarrierResource: update ##: Error Create AC",
				"Create: Could not Marshal CarrierResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ParentColId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" +  plan.Identifier.GrandParentColId.ValueString()  + "/linePtps/" +  plan.Identifier.ParentColId.ValueString() + "/carrier" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"CarrierResource: update ##: Error update CarrierResource",
				"Update: Could not update CarrierResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "CarrierResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"CarrierResource: Create ##: Error Unmarshal response",
				"Update:Could not Create CarrierResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "CarrierResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "CarrierResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *CarrierResource) read(state *CarrierResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull() || state.Identifier.GrandParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read CarrierResource",
			"CarrierResource: Could not read. Id, and Href, and identifiers are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "CarrierResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.GrandParentColId.String() + "/linePtps/" + state.Identifier.ParentColId.ValueString() + "/carrier" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.GrandParentColId.String() + "/linePtps/" + state.Identifier.ParentColId.ValueString() + "/carrier" + queryStr + "&q={\"state.carrierAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.GrandParentColId.String() + "/linePtps/" + state.Identifier.ParentColId.ValueString() + "/carrier" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.GrandParentColId.String() + "/linePtps/" + state.Identifier.ParentColId.ValueString() + "/carriers/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.GrandParentColId.String() + "/linePtps" + state.Identifier.ParentColId.ValueString() + "/carrier" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"CarrierResource: read ##: Error Read CarrierResource",
			"Read:Could not get CarrierResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "CarrierResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"CarrierResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal CarrierResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"CarrierResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "CarrierResource: read ## ", map[string]interface{}{"plan": state})
}

func (r *CarrierResource) delete(plan *CarrierResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "CarrierResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (carrierData *CarrierResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "CarrierResourceData: populate ## ", map[string]interface{}{"plan": data})

	carrierData.Id = types.StringValue(data["id"].(string))
	carrierData.ColId = types.Int64Value(int64(data["colid"].(float64)))
	carrierData.Href = types.StringValue(data["href"].(string))
	carrierData.ParentId = types.StringValue(data["parentId"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if carrierData.Config == nil {
		carrierData.Config = &CarrierConfig{}
	}
	for k, v := range config {
		switch k {
		case "name":
			if !carrierData.Config.Name.IsNull() {
				carrierData.Config.Name = types.StringValue(v.(string))
			}
		case "frequency":
			if !carrierData.Config.Frequency.IsNull() {
				carrierData.Config.Frequency = types.Int64Value(int64(v.(float64)))
			}
		case "modulation":
			if !carrierData.Config.Modulation.IsNull() {
				carrierData.Config.Modulation = types.StringValue(v.(string))
			}
		case "tx_clp_target":
			if !carrierData.Config.TxCLPtarget.IsNull() {
				carrierData.Config.TxCLPtarget = types.Int64Value(int64(v.(float64)))
			}
		case "diagnostics":
			diagnostics := v.(map[string]interface{})
			if diagnostics["termLB"] != nil && !carrierData.Config.Diagnostics.TermLB.IsNull() {
				carrierData.Config.Diagnostics.TermLB = types.StringValue(v.(string))
			}
			if diagnostics["termLBDuration"] != nil && !carrierData.Config.Diagnostics.TermLBDuration.IsNull() {
				carrierData.Config.Diagnostics.TermLBDuration = types.Int64Value(int64(v.(float64)))
			}
			if diagnostics["facPRBSGen"] != nil && !carrierData.Config.Diagnostics.FacPRBSGen.IsNull() {
				carrierData.Config.Diagnostics.FacPRBSGen = types.BoolValue(v.(bool))
			}
			if diagnostics["facPRBSMon"] != nil && !carrierData.Config.Diagnostics.FacPRBSMon.IsNull() {
				carrierData.Config.Diagnostics.FacPRBSMon = types.BoolValue(v.(bool))
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		carrierData.State = types.ObjectValueMust(CarrierStateAttributeType(), CarrierStateAttributeValue(state))
	}

	tflog.Debug(ctx, "CarrierResourceData: read ## ", map[string]interface{}{"carrierData": carrierData})
}

func CarrierResourceSchemaAttributes() map[string]schema.Attribute {
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
				"modulation": schema.StringAttribute{
					Description: "modulation",
					Optional:    true,
				},
				"tx_clp_target": schema.Int64Attribute{
					Description: "txCLPtarget",
					Optional:    true,
				},
				"frequency": schema.Int64Attribute{
					Description: "frequency",
					Optional:    true,
				},
				"diagnostics": schema.SingleNestedAttribute{
					Description: "diagnostics",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"term_lb": schema.StringAttribute{
							Description: "term_lb",
							Optional:    true,
						},
						"term_lb_duration": schema.Int64Attribute{
							Description: "term_lb_duration",
							Optional:    true,
						},
						"fac_lb": schema.StringAttribute{
							Description: "fac_lb",
							Optional:    true,
						},
						"fac_lb_duration": schema.Int64Attribute{
							Description: "fac_lb_duration",
							Optional:    true,
						},
					},
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: CarrierStateAttributeType(),
		},
	}
}

func CarrierObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: CarrierAttributeType(),
	}
}

func CarrierObjectsValue(data []interface{}) []attr.Value {
	carriers := []attr.Value{}
	for _, v := range data {
		carrier := v.(map[string]interface{})
		if carrier != nil {
			carriers = append(carriers, types.ObjectValueMust(
				CarrierAttributeType(),
				CarrierAttributeValue(carrier)))
		}
	}
	return carriers
}

func CarrierAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.StringType,
		"parent_id": types.StringType,
		"href":      types.StringType,
		"col_id":    types.Int64Type,
		"config":          types.ObjectType{AttrTypes: CarrierConfigAttributeType()},
		"state":           types.ObjectType{AttrTypes: CarrierStateAttributeType()},
	}
}

func CarrierAttributeValue(carrier map[string]interface{}) map[string]attr.Value {
	colId := types.Int64Null()
	if carrier["colId"] != nil {
		colId = types.Int64Value(int64(carrier["colId"].(float64)))
	}
	href := types.StringNull()
	if carrier["href"] != nil {
		href = types.StringValue(carrier["href"].(string))
	}
	id := types.StringNull()
	if carrier["id"] != nil {
		id = types.StringValue(carrier["id"].(string))
	}
	parentId := types.StringNull()
	if carrier["parentId"] != nil {
		parentId = types.StringValue(carrier["parentId"].(string))
	}
	config := types.ObjectNull(CarrierConfigAttributeType())
	if (carrier["config"]) != nil {
		config = types.ObjectValueMust(CarrierConfigAttributeType(), CarrierConfigAttributeValue(carrier["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(CarrierStateAttributeType())
	if (carrier["state"]) != nil {
		state = types.ObjectValueMust(CarrierStateAttributeType(), CarrierStateAttributeValue(carrier["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"col_id":          colId,
		"parent_id": parentId,
		"id":              id,
		"href":            href,
		"config":          config,
		"state":           state,
	}
}

func CarrierConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":          types.StringType,
		"frequency":     types.Int64Type,
		"modulation":    types.StringType,
		"tx_clp_target": types.Int64Type,
		"diagnostics":   types.ObjectType{AttrTypes: CarrierDiagnosticsAttributeType()},
	}
}

func CarrierConfigAttributeValue(carrierConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	modulation := types.StringNull()
	frequency := types.Int64Null()
	txCLPtarget := types.Int64Null()
	diagnostics := types.ObjectNull(CarrierDiagnosticsAttributeType())

	for k, v := range carrierConfig {
		switch k {
		case "name":
			name = types.StringValue(v.(string))
		case "modulation":
			modulation = types.StringValue(v.(string))
		case "frequency":
			frequency = types.Int64Value(int64(v.(float64)))
		case "txCLPtarget":
			txCLPtarget = types.Int64Value(int64(v.(float64)))
		case "diagnostics":
			diagnostics = types.ObjectValueMust(CarrierDiagnosticsAttributeType(), CarrierDiagnosticsAttributeValue(carrierConfig["diagnostics"].(map[string]interface{})))
		}
	}

	return map[string]attr.Value{
		"name":          name,
		"modulation":    modulation,
		"frequency":     frequency,
		"tx_clp_target": txCLPtarget,
		"diagnostics":   diagnostics,
	}
}

func CarrierStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"parent_aid":        types.StringType,
		"carrier_aid":        types.StringType,
		"modulation":         types.StringType,
		"frequency":          types.Int64Type,
		"operating_frequency": types.Int64Type,
		"tx_clp_target":      types.Int64Type,
		"diagnostics":        types.ObjectType{AttrTypes: CarrierDiagnosticsAttributeType()},
		"lifecycle_state":    types.StringType,
	}
}

func CarrierStateAttributeValue(carrierState map[string]interface{}) map[string]attr.Value {
	carrierAid := types.StringNull()
	parentAid := types.StringNull()
	name := types.StringNull()
	modulation := types.StringNull()
	frequency := types.Int64Null()
	operatingFrequency := types.Int64Null()
	txCLPtarget := types.Int64Null()
	diagnostics := types.ObjectNull(CarrierDiagnosticsAttributeType())
	lifecycleState := types.StringNull()

	for k, v := range carrierState {
		switch k {
		case "carrierAid":
			carrierAid = types.StringValue(v.(string))
		case "parentAid":
			parentAids := v.([]interface{})
			parentAid = types.StringValue(parentAids[0].(string))
		case "name":
			name = types.StringValue(v.(string))
		case "modulation":
			modulation = types.StringValue(v.(string))
		case "frequency":
			frequency = types.Int64Value(int64(v.(float64)))
		case "operatingFrequency":
			operatingFrequency = types.Int64Value(int64(v.(float64)))
		case "txCLPtarget":
			txCLPtarget = types.Int64Value(int64(v.(float64)))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		case "diagnostics":
			diagnostics = types.ObjectValueMust(CarrierDiagnosticsAttributeType(), CarrierDiagnosticsAttributeValue(v.(map[string]interface{})))
		}
	}

	return map[string]attr.Value{
		"carrier_aid":         carrierAid,
		"parent_aid":         parentAid,
		"name":                name,
		"modulation":          modulation,
		"frequency":           frequency,
		"operating_frequency": operatingFrequency,
		"tx_clp_target":       txCLPtarget,
		"lifecycle_state":    lifecycleState,
		"diagnostics":         diagnostics,
	}
}

func CarrierDiagnosticsAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"term_lb":          types.StringType,
		"term_lb_duration": types.Int64Type,
		"fac_lb":          types.StringType,
		"fac_lb_duration": types.Int64Type,
	}
}

func CarrierDiagnosticsAttributeValue(diagnostics map[string]interface{}) map[string]attr.Value {
	termLB := types.StringNull()
	if diagnostics["termLB"] != nil {
		termLB = types.StringValue(diagnostics["termLB"].(string))
	}
	termLBDuration := types.Int64Null()
	if diagnostics["termLBDuration"] != nil {
		termLBDuration = types.Int64Value(int64(diagnostics["termLBDuration"].(float64)))
	}

	facLB := types.StringNull()
	if diagnostics["facLB"] != nil {
		facLB = types.StringValue(diagnostics["facLB"].(string))
	}
	facLBDuration := types.Int64Null()
	if diagnostics["facLBDuration"] != nil {
		facLBDuration = types.Int64Value(int64(diagnostics["facLBDuration"].(float64)))
	}

	return map[string]attr.Value{
		"term_lb":          termLB,
		"term_lb_duration": termLBDuration,
		"fac_lb":     facLB,
		"fac_lb_duration":     facLBDuration,
	}
}
