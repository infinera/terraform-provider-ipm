package nduservice

import (
	"context"
	"encoding/json"
	"strconv"

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
	_ resource.Resource                = &EClientResource{}
	_ resource.ResourceWithConfigure   = &EClientResource{}
	_ resource.ResourceWithImportState = &EClientResource{}
)

// NewEClientResource is a helper function to simplify the provider implementation.
func NewEClientResource() resource.Resource {
	return &EClientResource{}
}

type EClientResource struct {
	client *ipm_pf.Client
}

type EClientDiagnostics struct {
	TermLB         types.String `tfsdk:"term_lb"`
	TermLBDuration types.Int64  `tfsdk:"term_lb_duration"`
	FacLB          types.String `tfsdk:"fac_lb"`
	FacLBDuration  types.Int64  `tfsdk:"fac_lb_duration"`
	FacTestingGen  types.Bool   `tfsdk:"fac_testing_gen"`
	FacTestingMon  types.Bool   `tfsdk:"fac_testing_mon"`
	TermTestingGen types.Bool   `tfsdk:"term_testing_gen"`
}

type EClientConfig struct {
	FecMode     types.String       `tfsdk:"fec_mode"`
	Name        types.String       `tfsdk:"name"`
	Diagnostics EClientDiagnostics `tfsdk:"diagnostics"`
}

type EClientResourceData struct {
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href       types.String              `tfsdk:"href"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	ColId      types.Int64               `tfsdk:"colid"`
	Config *EClientConfig `tfsdk:"config"`
	State  types.Object   `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *EClientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_ethernet_client"
}

// Schema defines the schema for the data source.
func (r *EClientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type EClientResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an ModuleLinePTP",
		Attributes:  EClientResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *EClientResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r EClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EClientResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "EClientResource: Create - ", map[string]interface{}{"EClientResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r EClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EClientResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "EClientResource: Create - ", map[string]interface{}{"EClientResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r EClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EClientResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "EClientResource: Update", map[string]interface{}{"EClientResourceData": data})

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

func (r EClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EClientResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "EClientResource: Update", map[string]interface{}{"EClientResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *EClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *EClientResource) create(plan *EClientResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "EClientResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *EClientResource) update(plan *EClientResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "EClientResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"EClientResource: Error update Carrier",
			"EClientResource: Could not update Carrier. Href, NDUId, ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get TC config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.FecMode.IsNull() {
		updateRequest["fecMode"] = plan.Config.FecMode.ValueString()
	}
	diagnostics := make(map[string]interface{})
	if !plan.Config.Diagnostics.TermLB.IsNull() {
		diagnostics["termLB"] = plan.Config.Diagnostics.TermLB.ValueString()
	}
	if !plan.Config.Diagnostics.TermLBDuration.IsNull() {
		diagnostics["termLBDuration"] = plan.Config.Diagnostics.TermLBDuration.ValueInt64()
	}
	if !plan.Config.Diagnostics.FacLB.IsNull() {
		diagnostics["facLB"] = plan.Config.Diagnostics.FacLB.ValueString()
	}
	if !plan.Config.Diagnostics.FacLBDuration.IsNull() {
		diagnostics["facLBDuration"] = plan.Config.Diagnostics.FacLBDuration.ValueInt64()
	}
	if !plan.Config.Diagnostics.FacTestingGen.IsNull() {
		diagnostics["facTestingGen"] = plan.Config.Diagnostics.FacTestingGen.ValueBool()
	}
	if !plan.Config.Diagnostics.FacTestingMon.IsNull() {
		diagnostics["facTestingMon"] = plan.Config.Diagnostics.FacTestingMon.ValueBool()
	}
	if !plan.Config.Diagnostics.TermTestingGen.IsNull() {
		diagnostics["termTestingGen"] = plan.Config.Diagnostics.TermTestingGen.ValueBool()
	}
	if len(diagnostics) > 0 {
		updateRequest["diagnostics"] = diagnostics
	}

	tflog.Debug(ctx, "EClientResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"EClientResource: update ##: Error Create AC",
				"Create: Could not Marshal EClientResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.ColId.IsNull() && !plan.ParentId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.ParentId.ValueString() + "/ethernets/" +  strconv.FormatInt(plan.ColId.ValueInt64(),10), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ethernets/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"PortResource: update ##: Error update porr}",
				"Update: Could not update PortResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "EClientResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"EClientResource: update ##: Error Unmarshal response",
				"Update:Could not update EClientResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "EClientResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "EClientResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *EClientResource) read(state *EClientResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && state.Identifier.ColId.IsNull() && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull() {
		diags.AddError(
			"Error Read EClientResource",
			"EClientResource: Could not read. Id, and Href, and identifier are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "EClientResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/ethernets" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/ethernets" + queryStr + "&q={\"state.clientIfAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/ethernets" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/ethernets/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		diags.AddError(
			"EClientResource: read ##: Error Read EClientResource",
			"Read:Could not get EClientResource, No identifier specified",
		)
		return
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"EClientResource: read ##: Error Read EClientResource",
			"Read:Could not get EClientResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "EClientResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})

	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"EClientResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal EClientResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"EClientResource: read ##: Can not get Module",
				"Read:Could not get EClient for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "EClientResource: read ## ", map[string]interface{}{"plan": state})
}

func (r *EClientResource) delete(plan *EClientResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "EClientResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (eClientData *EClientResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "EClientResourceData: populate ## ", map[string]interface{}{"plan": data})

	eClientData.Id = types.StringValue(data["id"].(string))
	eClientData.ParentId = types.StringValue(data["parentId"].(string))
	eClientData.Href = types.StringValue(data["href"].(string))
	eClientData.ColId = types.Int64Value(int64(data["colid"].(float64)))

	// populate config
	var config = data["config"].(map[string]interface{})
	if config != nil {
		if eClientData.Config == nil {
			eClientData.Config = &EClientConfig{}
		}
		for k, v := range config {
			switch k {
			case "fecMode":
				if !eClientData.Config.FecMode.IsNull() {
					eClientData.Config.FecMode = types.StringValue(v.(string))
				}
			case "name":
				if !eClientData.Config.Name.IsNull() {
					eClientData.Config.Name = types.StringValue(v.(string))
				}
			case "diagnostics":
				diagnostics := v.(map[string]interface{})
				if diagnostics["termLB"] != nil && !eClientData.Config.Diagnostics.TermLB.IsNull() {
					eClientData.Config.Diagnostics.TermLB = types.StringValue(v.(string))
				}
				if diagnostics["termLBDuration"] != nil && !eClientData.Config.Diagnostics.TermLBDuration.IsNull() {
					eClientData.Config.Diagnostics.TermLBDuration = types.Int64Value(int64(v.(float64)))
				}
				if diagnostics["facLB"] != nil && !eClientData.Config.Diagnostics.FacLB.IsNull() {
					eClientData.Config.Diagnostics.FacLB = types.StringValue(v.(string))
				}
				if diagnostics["facLBDuration"] != nil && !eClientData.Config.Diagnostics.FacLBDuration.IsNull() {
					eClientData.Config.Diagnostics.FacLBDuration = types.Int64Value(int64(v.(float64)))
				}
				if diagnostics["facTestingGen"] != nil && !eClientData.Config.Diagnostics.FacTestingGen.IsNull() {
					eClientData.Config.Diagnostics.FacTestingGen = types.BoolValue(v.(bool))
				}
				if diagnostics["facTestingMon"] != nil && !eClientData.Config.Diagnostics.FacTestingMon.IsNull() {
					eClientData.Config.Diagnostics.FacTestingMon = types.BoolValue(v.(bool))
				}
				if diagnostics["termTestingGen"] != nil && !eClientData.Config.Diagnostics.TermTestingGen.IsNull() {
					eClientData.Config.Diagnostics.TermTestingGen = types.BoolValue(v.(bool))
				}
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		eClientData.State = types.ObjectValueMust(EClientStateAttributeType(), EClientStateAttributeValue(state))
	}
	tflog.Debug(ctx, "EClientResourceData: read ## ", map[string]interface{}{"eClientData": eClientData})
}

func EClientResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Module.",
			Computed:    true,
		},
		"parent_id": schema.StringAttribute{
			Description: "module id",
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
			Description: "Network Connection Ethernet Client Config",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"fec_mode": schema.StringAttribute{
					Description: "fec_mode",
					Optional:    true,
				},
				"name": schema.StringAttribute{
					Description: "name",
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
						"fac_testing_gen": schema.StringAttribute{
							Description: "fac_testing_gen",
							Optional:    true,
						},
						"fac_testing_mon": schema.StringAttribute{
							Description: "fac_testing_mon",
							Optional:    true,
						},
						"term_testing_gen": schema.StringAttribute{
							Description: "term_testing_gen",
							Optional:    true,
						},
					},
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: EClientStateAttributeType(),
		},
	}
}

func EClientObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: EClientAttributeType(),
	}
}

func EClientObjectsValue(data []interface{}) []attr.Value {
	eClients := []attr.Value{}
	for _, v := range data {
		eClient := v.(map[string]interface{})
		if eClient != nil {
			eClients = append(eClients, types.ObjectValueMust(
				EClientAttributeType(),
				EClientAttributeValue(eClient)))
		}
	}
	return eClients
}

func EClientAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":        types.StringType,
		"href":      types.StringType,
		"col_id":    types.Int64Type,
		"config":    types.ObjectType{AttrTypes: EClientConfigAttributeType()},
		"state":     types.ObjectType{AttrTypes: EClientStateAttributeType()},
	}
}

func EClientAttributeValue(eClient map[string]interface{}) map[string]attr.Value {
	col_id := types.Int64Null()
	if eClient["colId"] != nil {
		col_id = types.Int64Value(int64(eClient["colId"].(float64)))
	}
	href := types.StringNull()
	if eClient["href"] != nil {
		href = types.StringValue(eClient["href"].(string))
	}
	id := types.StringNull()
	if eClient["id"] != nil {
		id = types.StringValue(eClient["id"].(string))
	}
	parentId := types.StringNull()
	if eClient["parentId"] != nil {
		parentId = types.StringValue(eClient["parentId"].(string))
	}
	config := types.ObjectNull(EClientConfigAttributeType())
	if (eClient["config"]) != nil {
		config = types.ObjectValueMust(EClientConfigAttributeType(), EClientConfigAttributeValue(eClient["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(EClientStateAttributeType())
	if (eClient["state"]) != nil {
		state = types.ObjectValueMust(EClientStateAttributeType(), EClientStateAttributeValue(eClient["state"].(map[string]interface{})))
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
func EClientConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"fec_mode":    types.StringType,
		"name":        types.StringType,
		"diagnostics": types.ObjectType{AttrTypes: EClienDiagnosticsAttributeType()},
	}
}

func EClientConfigAttributeValue(eClientConfig map[string]interface{}) map[string]attr.Value {
	fecMode := types.StringNull()
	if eClientConfig["fecMode"] != nil {
		fecMode = types.StringValue(eClientConfig["fecMode"].(string))
	}
	name := types.StringNull()
	if eClientConfig["name"] != nil {
		name = types.StringValue(eClientConfig["name"].(string))
	}
	diagnostics := types.ObjectNull(EClienDiagnosticsAttributeType())
	if (eClientConfig["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(EClienDiagnosticsAttributeType(), EClientDiagnosticsAttributeValue(eClientConfig["diagnostics"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"fec_mode":    fecMode,
		"name":        name,
		"diagnostics": diagnostics,
	}
}

func EClientStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":       types.StringType,
		"client_if_aid":    types.StringType,
		"name":             types.StringType,
		"fec_mode":         types.StringType,
		"diagnostics":      types.ObjectType{AttrTypes: EClienDiagnosticsAttributeType()},
		"lifecycle_state": types.StringType,
	}
}

func EClientStateAttributeValue(eClientState map[string]interface{}) map[string]attr.Value {
	parentAid := types.StringNull()
	if eClientState["parentAid"] != nil {
		parentAids := eClientState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	clientIfAid := types.StringNull()
	if eClientState["clientIfAid"] != nil {
		clientIfAid = types.StringValue(eClientState["clientIfAid"].(string))
	}
	name := types.StringNull()
	if eClientState["name"] != nil {
		name = types.StringValue(eClientState["name"].(string))
	}
	fecMode := types.StringNull()
	if eClientState["fecMode"] != nil {
		fecMode = types.StringValue(eClientState["fecMode"].(string))
	}

	diagnostics := types.ObjectNull(EClienDiagnosticsAttributeType())
	if (eClientState["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(EClienDiagnosticsAttributeType(), EClientDiagnosticsAttributeValue(eClientState["diagnostics"].(map[string]interface{})))
	}
	lifecycleState := types.StringNull()
	if eClientState["lifecycleState"] != nil {
		lifecycleState = types.StringValue(eClientState["lifecycleState"].(string))
	}

	return map[string]attr.Value{
		"parent_aid":       parentAid,
		"client_if_aid":      clientIfAid,
		"name":             name,
		"fec_mode":             fecMode,
		"diagnostics":      diagnostics,
		"lifecycle_state": lifecycleState,
	}
}

func EClienDiagnosticsAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"term_lb":          types.StringType,
		"term_lb_duration": types.Int64Type,
		"fac_lb":           types.StringType,
		"fac_lb_duration":  types.Int64Type,
		"fac_test_signal_gen":  types.StringType,
		"fac_test_signal_mon":  types.StringType,
		"term_test_signal_gen": types.StringType,
	}
}

func EClientDiagnosticsAttributeValue(diagnostics map[string]interface{}) map[string]attr.Value {
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
	facTestSignalGen := types.StringNull()
	if diagnostics["facTestSignalGen"] != nil {
		facTestSignalGen = types.StringValue(diagnostics["facTestSignalGen"].(string))
	}
	facTestSignalMon := types.StringNull()
	if diagnostics["facTestSignalMon"] != nil {
		facTestSignalMon = types.StringValue(diagnostics["facTestSignalMon"].(string))
	}
	termTestSignalMon := types.StringNull()
	if diagnostics["termTestSignalMon"] != nil {
		termTestSignalMon = types.StringValue(diagnostics["termTestSignalMon"].(string))
	}

	return map[string]attr.Value{
		"term_lb":          termLB,
		"term_lb_duration": termLBDuration,
		"fac_lb":           facLB,
		"fac_lb_duration":  facLBDuration,
		"fac_test_signal_gen":  facTestSignalGen,
		"fac_test_signal_mon":  facTestSignalMon,
		"term_test_signal_gen": termTestSignalMon,
	}
}
