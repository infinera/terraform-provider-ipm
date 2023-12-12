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
}

type CarrierConfig struct {
	Diagnostics CarrierDiagnostics `tfsdk:"diagnostics"`
}

type CarrierResourceData struct {
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href       types.String              `tfsdk:"href"`
	ColId      types.Int64               `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config     *CarrierConfig             `tfsdk:"config"`
	State      types.Object              `tfsdk:"state"`
	DSCGs      types.List                `tfsdk:"dscgs"`
	DSCs       types.List                `tfsdk:"dscs"`
}

// Metadata returns the data source type name.
func (r *CarrierResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carrier"
}

// Schema defines the schema for the data source.
func (r *CarrierResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type CarrierResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages Carrier",
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

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"CarrierResource: Error update Carrier",
			"CarrierResource: Could not update Carrier. Href or Carrier ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get TC config settings
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
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/modules/" + plan.Identifier.DeviceId.ValueString() + "/linePtps/" +  plan.Identifier.ParentColId.ValueString()  + "/carriers/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"CarrierResource: update ##: Error update Carrier",
				"Update: Could not update CarrierResource, Identfier (DeviceID, parentColID or ColId) is not specified: ",
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

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read CarrierResource",
			"CarrierResource: Could not read. Id, and Href, and identifiers are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "CarrierResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.ParentColId.ValueString() + "/carriers" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.ParentColId.ValueString() + "/carriers" + queryStr + "&q={\"state.carrierAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.ParentColId.ValueString() + "/carriers" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.ParentColId.ValueString() + "/carriers/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/modules" + state.Identifier.DeviceId.ValueString() + "/linePtps" + state.Identifier.ParentColId.ValueString() + "/carriers" + queryStr
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

func (carrier *CarrierResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "CarrierResourceData: populate ## ", map[string]interface{}{"plan": data})

	carrier.Id = types.StringValue(data["id"].(string))
	carrier.Href = types.StringValue(data["href"].(string))
	carrier.ColId = types.Int64Value(int64(data["colId"].(float64)))
	carrier.ParentId = types.StringValue(data["parentId"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if carrier.Config == nil {
		carrier.Config = &CarrierConfig{}
	}
	for k, v := range config {
		switch k {
		case "diagnostics":
			diagnostics := v.(map[string]interface{})
			if diagnostics["termLB"] != nil && !carrier.Config.Diagnostics.TermLB.IsNull() {
				carrier.Config.Diagnostics.TermLB = types.StringValue(v.(string))
			}
			if diagnostics["termLBDuration"] != nil && !carrier.Config.Diagnostics.TermLBDuration.IsNull() {
				carrier.Config.Diagnostics.TermLBDuration = types.Int64Value(int64(v.(float64)))
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		carrier.State = types.ObjectValueMust(CarrierStateAttributeType(), CarrierStateAttributeValue(state))
	}

	// populate DSCGs
	carrier.DSCGs = types.ListNull(DSCGObjectType())
	if data["dscgs"] != nil {
		carrier.DSCGs = types.ListValueMust(DSCGObjectType(), DSCGObjectsValue(data["dscgs"].([]interface{})))
	}

	// populate DSCs
	carrier.DSCs = types.ListNull(DSCObjectType())
	if data["dscgs"] != nil {
		carrier.DSCs = types.ListValueMust(DSCObjectType(), DSCObjectsValue(data["dscs"].([]interface{})))
	}

	tflog.Debug(ctx, "CarrierResourceData: read ## ", map[string]interface{}{"plan": state})
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
		//Config           NetworkConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Description: "config",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"relative_dpo": schema.Int64Attribute{
					Description: "relative DPO",
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
					},
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: CarrierStateAttributeType(),
		},
		"dscgs": schema.ListAttribute{
			Computed:    true,
			ElementType: DSCGObjectType(),
		},
		"dscs": schema.ListAttribute{
			Computed:    true,
			ElementType: DSCObjectType(),
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
		"config":    types.ObjectType{AttrTypes: CarrierConfigAttributeType()},
		"state":     types.ObjectType{AttrTypes: CarrierStateAttributeType()},
		"dscgs":     types.ListType{ElemType: DSCGObjectType()},
		"dscs":      types.ListType{ElemType: DSCObjectType()},
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

	dscgs := types.ListNull(DSCGObjectType())
	if (carrier["dscgs"]) != nil {
		dscgs = types.ListValueMust(DSCGObjectType(), DSCGObjectsValue(carrier["dscgs"].([]interface{})))
	}

	dscs := types.ListNull(DSCObjectType())
	if (carrier["dscs"]) != nil {
		dscs = types.ListValueMust(DSCObjectType(), DSCObjectsValue(carrier["dscs"].([]interface{})))
	}

	return map[string]attr.Value{
		"parent_id": parentId,
		"col_id":    colId,
		"id":        id,
		"href":      href,
		"config":    config,
		"state":     state,
		"dscgs":     dscgs,
		"dscs":      dscs,
	}
}

func CarrierDiagnosticsAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"term_lb":          types.StringType,
		"term_lb_duration": types.Int64Type,
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

	return map[string]attr.Value{
		"term_lb":          termLB,
		"term_lb_duration": termLBDuration,
	}
}

func CarrierConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"diagnostics": types.ObjectType{AttrTypes: CarrierDiagnosticsAttributeType()},
	}
}

func CarrierConfigAttributeValue(carrierConfig map[string]interface{}) map[string]attr.Value {
	diagnostics := types.ObjectNull(CarrierDiagnosticsAttributeType())
	if (carrierConfig["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(CarrierDiagnosticsAttributeType(), CarrierDiagnosticsAttributeValue(carrierConfig["diagnostics"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"diagnostics": diagnostics,
	}
}

func CarrierStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"carrier_aid":                    types.StringType,
		"parent_aid":                     types.StringType,
		"constellation_frequency":        types.Int64Type,
		"host_frequency":                 types.Int64Type,
		"actual_constellation_frequency": types.Int64Type,
		"operating_frequency":            types.Int64Type,
		"modulation":                     types.StringType,
		"host_modulation":                types.StringType,
		"operating_modulation":           types.StringType,
		"fec_iterations":                 types.StringType,
		"host_fec_iterations":            types.StringType,
		"operating_fec_iterations":       types.StringType,
		"spectral_bandwidth":             types.Int64Type,
		"client_port_mode":               types.StringType,
		"baud_rate":                      types.Int64Type,
		"tx_clp_target":                  types.Int64Type,
		"host_tx_clp_target":             types.Int64Type,
		"actual_tx_clp_target":           types.Int64Type,
		"adv_line_ctrl":                  types.StringType,
		"max_dscs":                       types.Int64Type,
		"host_max_dscs":                  types.Int64Type,
		"operating_max_dscs":             types.Int64Type,
		"allowed_rx_cdscs":               types.ListType{ElemType: types.Int64Type},
		"allowed_tx_cdscs":               types.ListType{ElemType: types.Int64Type},
		"host_allowed_tx_cdscs":          types.ListType{ElemType: types.Int64Type},
		"actual_allowed_tx_cdscs":        types.ListType{ElemType: types.Int64Type},
		"host_allowed_rx_cdscs":          types.ListType{ElemType: types.Int64Type},
		"actual_allowed_rx_cdscs":        types.ListType{ElemType: types.Int64Type},
		"capabilities":                   types.MapType{ElemType: types.StringType},
		"diagnostics":                    types.ObjectType{AttrTypes: CarrierDiagnosticsAttributeType()},
		"life_cycle_state":               types.StringType,
	}
}

func CarrierStateAttributeValue(carrierState map[string]interface{}) map[string]attr.Value {
	carrierAid := types.StringNull()
	if carrierState["carrierAid"] != nil {
		carrierAid = types.StringValue(carrierState["carrierAid"].(string))
	}
	parentAid := types.StringNull()
	if carrierState["parentAid"] != nil {
		parentAids := carrierState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	constellationFrequency := types.Int64Null()
	if carrierState["constellationFrequency"] != nil {
		constellationFrequency = types.Int64Value(int64(carrierState["constellationFrequency"].(float64)))
	}
	hostFrequency := types.Int64Null()
	if carrierState["hostFrequency"] != nil {
		hostFrequency = types.Int64Value(int64(carrierState["hostFrequency"].(float64)))
	}
	actualConstellationFrequency := types.Int64Null()
	if carrierState["actualConstellationFrequency"] != nil {
		actualConstellationFrequency = types.Int64Value(int64(carrierState["actualConstellationFrequency"].(float64)))
	}
	operatingFrequency := types.Int64Null()
	if carrierState["operatingFrequency"] != nil {
		operatingFrequency = types.Int64Value(int64(carrierState["operatingFrequency"].(float64)))
	}
	modulation := types.StringNull()
	if carrierState["modulation"] != nil {
		modulation = types.StringValue(carrierState["modulation"].(string))
	}
	hostModulation := types.StringNull()
	if carrierState["hostModulation"] != nil {
		hostModulation = types.StringValue(carrierState["hostModulation"].(string))
	}
	operatingModulation := types.StringNull()
	if carrierState["operatingModulation"] != nil {
		operatingModulation = types.StringValue(carrierState["operatingModulation"].(string))
	}
	fecIterations := types.StringNull()
	if carrierState["fecIterations"] != nil {
		fecIterations = types.StringValue(carrierState["fecIterations"].(string))
	}
	hostFecIterations := types.StringNull()
	if carrierState["hostFecIterations"] != nil {
		hostFecIterations = types.StringValue(carrierState["hostFecIterations"].(string))
	}
	operatingFecIterations := types.StringNull()
	if carrierState["operatingFecIterations"] != nil {
		operatingFecIterations = types.StringValue(carrierState["operatingFecIterations"].(string))
	}
	spectralBandwidth := types.Int64Null()
	if carrierState["spectralBandwidth"] != nil {
		spectralBandwidth = types.Int64Value(int64(carrierState["spectralBandwidth"].(float64)))
	}
	clientPortMode := types.StringNull()
	if carrierState["clientPortMode"] != nil {
		clientPortMode = types.StringValue(carrierState["clientPortMode"].(string))
	}
	baudRate := types.Int64Null()
	if carrierState["baudRate"] != nil {
		baudRate = types.Int64Value(int64(carrierState["baudRate"].(float64)))
	}
	txCLPtarget := types.Int64Null()
	if carrierState["txClpTarget"] != nil {
		txCLPtarget = types.Int64Value(int64(carrierState["txCLPtarget"].(float64)))
	}
	hostTxCLPtarget := types.Int64Null()
	if carrierState["hostTxCLPtarget"] != nil {
		hostTxCLPtarget = types.Int64Value(int64(carrierState["hostTxCLPtarget"].(float64)))
	}
	actualTxCLPtarget := types.Int64Null()
	if carrierState["actualTxClpTarget"] != nil {
		actualTxCLPtarget = types.Int64Value(int64(carrierState["actualTxCLPtarget"].(float64)))
	}
	advLineCtrl := types.StringNull()
	if carrierState["advLineCtrl"] != nil {
		advLineCtrl = types.StringValue(carrierState["advLineCtrl"].(string))
	}
	maxDSCs := types.Int64Null()
	if carrierState["maxDSCs"] != nil {
		maxDSCs = types.Int64Value(int64(carrierState["maxDSCs"].(float64)))
	}
	hostMaxDSCs := types.Int64Null()
	if carrierState["hostMaxDSCs"] != nil {
		hostMaxDSCs = types.Int64Value(int64(carrierState["hostMaxDSCs"].(float64)))
	}
	operatingMaxDSCs := types.Int64Null()
	if carrierState["operatingMaxDSCs"] != nil {
		operatingMaxDSCs = types.Int64Value(int64(carrierState["operatingMaxDSCs"].(float64)))
	}
	allowedTxCDSCs := types.ListNull(types.Int64Type)
	if carrierState["allowedTxCDSCs"] != nil {
		allowedTxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(carrierState["allowedTxCDSCs"].([]interface{})))
	}
	hostAllowedTxCDSCs := types.ListNull(types.Int64Type)
	if carrierState["hostAllowedTxCDSCs"] != nil {
		hostAllowedTxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(carrierState["hostAllowedTxCDSCs"].([]interface{})))
	}
	actualAllowedTxCDSCs := types.ListNull(types.Int64Type)
	if carrierState["actualAllowedTxCDSCs"] != nil {
		actualAllowedTxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(carrierState["actualAllowedTxCDSCs"].([]interface{})))
	}
	allowedRxCDSCs := types.ListNull(types.Int64Type)
	if carrierState["allowedRxCDSCs"] != nil {
		allowedRxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(carrierState["allowedRxCDSCs"].([]interface{})))
	}
	hostAllowedRxCDSCs := types.ListNull(types.Int64Type)
	if carrierState["hostAllowedRxCDSCs"] != nil {
		hostAllowedRxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(carrierState["hostAllowedRxCDSCs"].([]interface{})))
	}
	actualAllowedRxCDSCs := types.ListNull(types.Int64Type)
	if carrierState["actualAllowedRxCDSCs"] != nil {
		actualAllowedRxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(carrierState["actualAllowedRxCDSCs"].([]interface{})))
	}
	capabilities := types.MapNull(types.StringType)
	if carrierState["capabilities"] != nil {
		capabilities = types.MapValueMust(types.StringType, common.MapAttributeValue(carrierState["capabilities"].(map[string]interface{})))
	}
	diagnostics := types.ObjectNull(CarrierDiagnosticsAttributeType())
	if (carrierState["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(CarrierDiagnosticsAttributeType(), CarrierDiagnosticsAttributeValue(carrierState["diagnostics"].(map[string]interface{})))
	}
	lifecycleState := types.StringNull()
	if carrierState["lifecycleState"] != nil {
		lifecycleState = types.StringValue(carrierState["lifecycleState"].(string))
	}

	return map[string]attr.Value{
		"carrier_aid":                    carrierAid,
		"parent_aid":                     parentAid,
		"constellation_frequency":        constellationFrequency,
		"host_frequency":                 hostFrequency,
		"actual_constellation_frequency": actualConstellationFrequency,
		"operating_frequency":            operatingFrequency,
		"modulation":                     modulation,
		"host_modulation":                hostModulation,
		"operating_modulation":           operatingModulation,
		"fec_iterations":                 fecIterations,
		"host_fec_iterations":            hostFecIterations,
		"operating_fec_iterations":       operatingFecIterations,
		"spectral_bandwidth":             spectralBandwidth,
		"client_port_mode":               clientPortMode,
		"baud_rate":                      baudRate,
		"tx_clp_target":                  txCLPtarget,
		"host_tx_clp_target":             hostTxCLPtarget,
		"actual_tx_clp_target":           actualTxCLPtarget,
		"adv_line_ctrl":                  advLineCtrl,
		"max_dscs":                       maxDSCs,
		"host_max_dscs":                  hostMaxDSCs,
		"operating_max_dscs":             operatingMaxDSCs,
		"allowed_tx_cdscs":               allowedTxCDSCs,
		"host_allowed_tx_cdscs":          hostAllowedTxCDSCs,
		"actual_allowed_tx_cdscs":        actualAllowedTxCDSCs,
		"host_allowed_rx_cdscs":          hostAllowedRxCDSCs,
		"allowed_rx_cdscs":               allowedRxCDSCs,
		"actual_allowed_rx_cdscs":        actualAllowedRxCDSCs,
		"capabilities":                   capabilities,
		"diagnostics":                    diagnostics,
		"life_cycle_state":               lifecycleState,
	}
}
