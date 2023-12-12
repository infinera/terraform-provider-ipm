package moduleservice

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

type EClientConfigLLDP struct {
	AdminStatus types.String `tfsdk:"admin_status"`
	GccFwd      types.Bool   `tfsdk:"gcc_fwd"`
	HostRxDrop  types.Bool   `tfsdk:"host_rx_drop"`
	TTLUsage    types.Bool   `tfsdk:"ttl_usage"`
}

type EClientConfigDiagnostics struct {
	TermLB         types.String `tfsdk:"term_lb"`
	TermLBDuration types.Int64  `tfsdk:"term_lb_duration"`
	FacPRBSGen     types.Bool   `tfsdk:"fac_prbs_gen"`
	FacPRBSMon     types.Bool   `tfsdk:"fac_prbs_mon"`
	TermPRBSGen    types.Bool   `tfsdk:"term_prbs_gen"`
}

type EClientConfig struct {
	FecMode     types.String             `tfsdk:"fec_mode"`
	LLDP        EClientConfigLLDP        `tfsdk:"lldp"`
	Diagnostics EClientConfigDiagnostics `tfsdk:"diagnostics"`
}

type EClientResourceData struct {
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href       types.String              `tfsdk:"href"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	ColId      types.Int64               `tfsdk:"colid"`
	Config     *EClientConfig             `tfsdk:"config"`
	State      types.Object              `tfsdk:"state"`
	ACs        types.List                `tfsdk:"acs"`
}

// Metadata returns the data source type name.
func (r *EClientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ethernet_client"
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
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *EClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *EClientResource) update(plan *EClientResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "EClientResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) {
		diags.AddError(
			"EClientResource: Error update ethernetClient",
			"EClientResource: Could not update ethernetClient. Href, ModuleId, or ethernetClient ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get EClientResource config settings
	if !plan.Config.FecMode.IsNull() {
		updateRequest["fecMode"] = plan.Config.FecMode.ValueString()
	}

	// get LLDP
	var lldp = make(map[string]interface{})

	if !plan.Config.LLDP.AdminStatus.IsNull() {
		lldp["adminStatus"] = plan.Config.LLDP.AdminStatus.ValueString()
	}
	if !plan.Config.LLDP.GccFwd.IsNull() {
		lldp["gccFwd"] = plan.Config.LLDP.GccFwd.ValueBool()
	}
	if !plan.Config.LLDP.HostRxDrop.IsNull() {
		lldp["hostRxDrop"] = plan.Config.LLDP.HostRxDrop.ValueBool()
	}
	if !plan.Config.LLDP.TTLUsage.IsNull() {
		lldp["TTLUsage"] = plan.Config.LLDP.TTLUsage.ValueBool()
	}

	if len(lldp) > 0 {
		updateRequest["lldp"] = lldp
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
	if !plan.Config.Diagnostics.TermPRBSGen.IsNull() {
		diagnostics["termPRBSGen"] = plan.Config.Diagnostics.TermPRBSGen.ValueBool()
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
				"EClientResource: update ##: Error Create EClient",
				"Create: Could not Marshal EClient, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.ColId.IsNull() && !plan.ParentId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/modules/" + plan.ParentId.ValueString() + "/ethernetClients/" +  strconv.FormatInt(plan.ColId.ValueInt64(),10), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/modules/" + plan.Identifier.DeviceId.ValueString() + "/ethernetClients/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"EClientResource: update ##: Error update EClient",
				"Update: Could not update EClientResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}


		tflog.Debug(ctx, "EClientResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"EClientResource: Update ##: Error Unmarshal response",
				"Update:Could not Create EClient, unexpected error: "+err.Error(),
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
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/ethernetClients" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/ethernetClients" + queryStr + "&q={\"state.clientIfAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/ethernetClients" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/ethernetClients/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/modules" + state.Identifier.DeviceId.ValueString() + "/ethernetClients" + queryStr
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
	if eClientData.Config == nil {
		eClientData.Config = &EClientConfig{}
	}
	for k, v := range config {
		switch k {
		case "fecMode":
			if !eClientData.Config.FecMode.IsNull() {
				eClientData.Config.FecMode = types.StringValue(v.(string))
			}
		case "lldp":
			lldp := v.(map[string]interface{})
			if lldp["adminStatus"] != nil && !eClientData.Config.LLDP.AdminStatus.IsNull() {
				eClientData.Config.LLDP.AdminStatus = types.StringValue(v.(string))
			}
			if lldp["gccFwd"] != nil && !eClientData.Config.LLDP.GccFwd.IsNull() {
				eClientData.Config.LLDP.GccFwd = types.BoolValue(v.(bool))
			}
			if lldp["hostRxDrop"] != nil && !eClientData.Config.LLDP.HostRxDrop.IsNull() {
				eClientData.Config.LLDP.HostRxDrop = types.BoolValue(v.(bool))
			}
			if lldp["TTLUsage"] != nil && !eClientData.Config.LLDP.TTLUsage.IsNull() {
				eClientData.Config.LLDP.TTLUsage = types.BoolValue(v.(bool))
			}
		case "diagnostics":
			diagnostics := v.(map[string]interface{})
			if diagnostics["termLB"] != nil && !eClientData.Config.Diagnostics.TermLB.IsNull() {
				eClientData.Config.Diagnostics.TermLB = types.StringValue(v.(string))
			}
			if diagnostics["termLBDuration"] != nil && !eClientData.Config.Diagnostics.TermLBDuration.IsNull() {
				eClientData.Config.Diagnostics.TermLBDuration = types.Int64Value(int64(v.(float64)))
			}
			if diagnostics["facPRBSGen"] != nil && !eClientData.Config.Diagnostics.FacPRBSGen.IsNull() {
				eClientData.Config.Diagnostics.FacPRBSGen = types.BoolValue(v.(bool))
			}
			if diagnostics["facPRBSMon"] != nil && !eClientData.Config.Diagnostics.FacPRBSMon.IsNull() {
				eClientData.Config.Diagnostics.FacPRBSMon = types.BoolValue(v.(bool))
			}
			if diagnostics["termPRBSGen"] != nil && !eClientData.Config.Diagnostics.TermPRBSGen.IsNull() {
				eClientData.Config.Diagnostics.TermPRBSGen = types.BoolValue(v.(bool))
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		eClientData.State = types.ObjectValueMust(EClientStateAttributeType(), EClientStateAttributeValue(state))
	}
	// populate ACs
	eClientData.ACs = types.ListNull(ACObjectType())
	if data["acs"] != nil {
		eClientData.ACs = types.ListValueMust(ACObjectType(), ACObjectsValue(data["acs"].([]interface{})))
	}

	tflog.Debug(ctx, "EClientResourceData: read ## ", map[string]interface{}{"plan": state})
}

func EClientResourceSchemaAttributes(computeEntity_optional ...bool) map[string]schema.Attribute {
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
			Computed:    computeFlag,
			Optional:    optionalFlag,
			Attributes: map[string]schema.Attribute{
				"fec_mode": schema.StringAttribute{
					Description: "fec_mode",
					Computed:    computeFlag,
					Optional:    optionalFlag,
				},
				"lldp": schema.SingleNestedAttribute{
					Description: "diagnostics",
					Computed:    computeFlag,
					Optional:    optionalFlag,
					Attributes: map[string]schema.Attribute{
						"admin_status": schema.StringAttribute{
							Description: "admin_status",
							Computed:    computeFlag,
							Optional:    optionalFlag,
						},
						"gcc_fwd": schema.BoolAttribute{
							Description: "gcc_fwd",
							Computed:    computeFlag,
							Optional:    optionalFlag,
						},
						"host_rx_drop": schema.BoolAttribute{
							Description: "hostRxDrop",
							Computed:    computeFlag,
							Optional:    optionalFlag,
						},
						"ttl_usage": schema.BoolAttribute{
							Description: "TTLUsage",
							Computed:    computeFlag,
							Optional:    optionalFlag,
						},
					},
				},
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
						"fac_prbs_gen": schema.StringAttribute{
							Description: "facPRBSGen",
							Computed:    computeFlag,
							Optional:    optionalFlag,
						},
						"fac_prbs_mon": schema.Int64Attribute{
							Description: "fac_prbs_mon",
							Computed:    computeFlag,
							Optional:    optionalFlag,
						},
						"term_prbs_gen": schema.StringAttribute{
							Description: "term_prbs_gen",
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
			AttributeTypes: EClientStateAttributeType(),
		},
		"acs": schema.ListAttribute{
			Computed:    true,
			ElementType: ACObjectType(),
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
		"acs":       types.ListType{ElemType: ACObjectType()},
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
	acs := types.ListNull(ACObjectType())
	if (eClient["acs"]) != nil {
		acs = types.ListValueMust(ACObjectType(), ACObjectsValue(eClient["acs"].([]interface{})))
	}

	return map[string]attr.Value{
		"col_id":    col_id,
		"parent_id": parentId,
		"id":        id,
		"href":      href,
		"config":    config,
		"state":     state,
		"acs":       acs,
	}
}
func EClientConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"fec_mode":    types.StringType,
		"lldp":        types.ObjectType{AttrTypes: EClientConfigLLDPAttributeType()},
		"diagnostics": types.ObjectType{AttrTypes: EClientConfigDiagnosticsAttributeType()},
	}
}

func EClientConfigAttributeValue(eClientConfig map[string]interface{}) map[string]attr.Value {
	fecMode := types.StringNull()
	if eClientConfig["fecMode"] != nil {
		fecMode = types.StringValue(eClientConfig["fecMode"].(string))
	}
	lldp := types.ObjectNull(EClientConfigLLDPAttributeType())
	if (eClientConfig["lldp"]) != nil {
		lldp = types.ObjectValueMust(EClientConfigLLDPAttributeType(), EClientConfigLLDPAttributeValue(eClientConfig["lldp"].(map[string]interface{})))
	}
	diagnostics := types.ObjectNull(EClientConfigDiagnosticsAttributeType())
	if (eClientConfig["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(EClientConfigDiagnosticsAttributeType(), EClientConfigDiagnosticsAttributeValue(eClientConfig["diagnostics"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"fec_mode":    fecMode,
		"lldp":        lldp,
		"diagnostics": diagnostics,
	}
}

func EClientConfigLLDPAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"admin_status": types.StringType,
		"gcc_fwd":      types.BoolType,
		"host_rx_drop": types.BoolType,
		"ttl_usage":    types.BoolType,
	}
}

func EClientConfigLLDPAttributeValue(lldp map[string]interface{}) map[string]attr.Value {
	adminStatus := types.StringNull()
	if lldp["adminStatus"] != nil {
		adminStatus = types.StringValue(lldp["adminStatus"].(string))
	}
	gccFwd := types.BoolNull()
	if lldp["gccFwd"] != nil {
		gccFwd = types.BoolValue(lldp["gccFwd"].(bool))
	}
	hostRxDrop := types.BoolNull()
	if lldp["hostRxDrop"] != nil {
		hostRxDrop = types.BoolValue(lldp["hostRxDrop"].(bool))
	}

	TTLUsage := types.BoolNull()
	if lldp["TTLUsage"] != nil {
		TTLUsage = types.BoolValue(lldp["TTLUsage"].(bool))
	}

	return map[string]attr.Value{
		"admin_status": adminStatus,
		"gcc_fwd":      gccFwd,
		"host_rx_drop": hostRxDrop,
		"ttl_usage":    TTLUsage,
	}
}

func EClientStateLLDPAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"lldp_config_aid":    types.StringType,
		"admin_status":       types.StringType,
		"gcc_fwd":            types.BoolType,
		"host_rx_drop":       types.BoolType,
		"ttl_usage":          types.BoolType,
		"too_many_neighbors": types.BoolType,
		"clr_stats":          types.BoolType,
		"flush_host_db":      types.BoolType,
		"neighbors":          types.ListType{ElemType: EClientLLDPNeighborObjectType()},
	}
}

func EClientStateLLDPAttributeValue(lldp map[string]interface{}) map[string]attr.Value {
	lldpconfigAid := types.StringNull()
	if lldp["lldpconfigAid"] != nil {
		lldpconfigAid = types.StringValue(lldp["lldpconfigAid"].(string))
	}
	adminStatus := types.StringNull()
	if lldp["adminStatus"] != nil {
		adminStatus = types.StringValue(lldp["adminStatus"].(string))
	}
	gccFwd := types.BoolNull()
	if lldp["gccFwd"] != nil {
		gccFwd = types.BoolValue(lldp["gccFwd"].(bool))
	}
	hostRxDrop := types.BoolNull()
	if lldp["hostRxDrop"] != nil {
		hostRxDrop = types.BoolValue(lldp["hostRxDrop"].(bool))
	}
	TTLUsage := types.BoolNull()
	if lldp["TTLUsage"] != nil {
		TTLUsage = types.BoolValue(lldp["TTLUsage"].(bool))
	}
	tooManyNeighbors := types.BoolNull()
	if lldp["tooManyNeighbors"] != nil {
		tooManyNeighbors = types.BoolValue(lldp["tooManyNeighbors"].(bool))
	}
	clrStats := types.BoolNull()
	if lldp["clrStats"] != nil {
		clrStats = types.BoolValue(lldp["clrStats"].(bool))
	}
	flushHostDb := types.BoolNull()
	if lldp["flushHostDb"] != nil {
		flushHostDb = types.BoolValue(lldp["flushHostDb"].(bool))
	}
	neighbors := types.ListNull(EClientLLDPNeighborObjectType())
	if (lldp["neighbors"]) != nil {
		neighbors = types.ListValueMust(EClientLLDPNeighborObjectType(), EClientLLDPNeighborObjectsValue(lldp["neighbors"].([]interface{})))
	}

	return map[string]attr.Value{
		"lldp_config_aid":    lldpconfigAid,
		"admin_status":       adminStatus,
		"gcc_fwd":            gccFwd,
		"host_rx_drop":       hostRxDrop,
		"ttl_usage":          TTLUsage,
		"too_many_neighbors": tooManyNeighbors,
		"clr_stats":          clrStats,
		"flush_host_db":      flushHostDb,
		"neighbors":          neighbors,
	}
}

func EClientStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":           types.StringType,
		"client_if_aid":        types.StringType,
		"client_if_port_speed": types.Int64Type,
		"fec_type":             types.StringType,
		"fec_mode":             types.StringType,
		"lldp":                 types.ObjectType{AttrTypes: EClientStateLLDPAttributeType()},
		"diagnostics":          types.ObjectType{AttrTypes: EClientStateDiagnosticsAttributeType()},
		"life_cycle_state":     types.StringType,
	}
}

func EClientStateAttributeValue(eClientState map[string]interface{}) map[string]attr.Value {
	clientIfAid := types.StringNull()
	if eClientState["clientIfAid"] != nil {
		clientIfAid = types.StringValue(eClientState["clientIfAid"].(string))
	}
	parentAid := types.StringNull()
	if eClientState["parentAid"] != nil {
		parentAids := eClientState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	fecType := types.StringNull()
	if eClientState["fecType"] != nil {
		fecType = types.StringValue(eClientState["fecType"].(string))
	}

	fecMode := types.StringNull()
	if eClientState["fecMode"] != nil {
		fecMode = types.StringValue(eClientState["fecMode"].(string))
	}

	clientIfPortSpeed := types.Int64Null()
	if eClientState["clientIfPortSpeed"] != nil {
		clientIfPortSpeed = types.Int64Value(int64(eClientState["clientIfPortSpeed"].(float64)))
	}
	lldp := types.ObjectNull(EClientStateLLDPAttributeType())
	if eClientState["lldp"] != nil {
		lldp = types.ObjectValueMust(EClientStateLLDPAttributeType(), EClientStateLLDPAttributeValue(eClientState["lldp"].(map[string]interface{})))
	}
	diagnostics := types.ObjectNull(EClientStateDiagnosticsAttributeType())
	if (eClientState["diagnostics"]) != nil {
		diagnostics = types.ObjectValueMust(EClientStateDiagnosticsAttributeType(), EClientStateDiagnosticsAttributeValue(eClientState["diagnostics"].(map[string]interface{})))
	}
	lifecycleState := types.StringNull()
	if eClientState["lifecycleState"] != nil {
		lifecycleState = types.StringValue(eClientState["lifecycleState"].(string))
	}

	return map[string]attr.Value{
		"parent_aid":           parentAid,
		"client_if_aid":        clientIfAid,
		"client_if_port_speed": clientIfPortSpeed,
		"fec_type":             fecType,
		"fec_mode":             fecMode,
		"lldp":                 lldp,
		"diagnostics":          diagnostics,
		"life_cycle_state":     lifecycleState,
	}
}

func EClientConfigDiagnosticsAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"term_lb":          types.StringType,
		"term_lb_duration": types.Int64Type,
		"fac_prbs_gen":     types.BoolType,
		"fac_prbs_mon":     types.BoolType,
		"term_prbs_gen":    types.BoolType,
	}
}

func EClientConfigDiagnosticsAttributeValue(diagnostics map[string]interface{}) map[string]attr.Value {
	termLB := types.StringNull()
	if diagnostics["termLB"] != nil {
		termLB = types.StringValue(diagnostics["termLB"].(string))
	}
	termLBDuration := types.Int64Null()
	if diagnostics["termLBDuration"] != nil {
		termLBDuration = types.Int64Value(int64(diagnostics["termLBDuration"].(float64)))
	}
	facPRBSGen := types.BoolNull()
	if diagnostics["facPRBSGen"] != nil {
		facPRBSGen = types.BoolValue(diagnostics["facPRBSGen"].(bool))
	}
	facPRBSMon := types.BoolNull()
	if diagnostics["facPRBSMon"] != nil {
		facPRBSMon = types.BoolValue(diagnostics["facPRBSMon"].(bool))
	}
	termPRBSGen := types.BoolNull()
	if diagnostics["termPRBSGen"] != nil {
		termPRBSGen = types.BoolValue(diagnostics["termPRBSGen"].(bool))
	}

	return map[string]attr.Value{
		"term_lb":          termLB,
		"term_lb_duration": termLBDuration,
		"fac_prbs_gen":     facPRBSGen,
		"fac_prbs_mon":     facPRBSMon,
		"term_prbs_gen":    termPRBSGen,
	}
}

func EClientStateDiagnosticsAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"term_lb":          types.StringType,
		"term_lb_duration": types.Int64Type,
		"fac_lb":           types.StringType,
		"fac_lb_duration":  types.Int64Type,
		"fac_prbs_gen":     types.BoolType,
		"fac_prbs_mon":     types.BoolType,
		"term_prbs_gen":    types.BoolType,
	}
}

func EClientStateDiagnosticsAttributeValue(diagnostics map[string]interface{}) map[string]attr.Value {
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
	facPRBSGen := types.BoolNull()
	if diagnostics["facPRBSGen"] != nil {
		facPRBSGen = types.BoolValue(diagnostics["facPRBSGen"].(bool))
	}
	facPRBSMon := types.BoolNull()
	if diagnostics["facPRBSMon"] != nil {
		facPRBSMon = types.BoolValue(diagnostics["facPRBSMon"].(bool))
	}
	termPRBSGen := types.BoolNull()
	if diagnostics["termPRBSGen"] != nil {
		termPRBSGen = types.BoolValue(diagnostics["termPRBSGen"].(bool))
	}

	return map[string]attr.Value{
		"term_lb":          termLB,
		"term_lb_duration": termLBDuration,
		"fac_lb":           facLB,
		"fac_lb_duration":  facLBDuration,
		"fac_prbs_gen":     facPRBSGen,
		"fac_prbs_mon":     facPRBSMon,
		"term_prbs_gen":    termPRBSGen,
	}
}

func EClientLLDPNeighborObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: EClientLLDPNeighborAttributeType(),
	}
}

func EClientLLDPNeighborObjectsValue(data []interface{}) []attr.Value {
	eClients := []attr.Value{}
	for _, v := range data {
		eClient := v.(map[string]interface{})
		if eClient != nil {
			eClients = append(eClients, types.ObjectValueMust(
				EClientLLDPNeighborAttributeType(),
				EClientLLDPNeighborAttributeValue(eClient)))
		}
	}
	return eClients
}

func EClientLLDPNeighborAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"host_neighbor_aid":  types.StringType,
		"chassis_id_subtype": types.StringType,
		"chassis_id":         types.StringType,
		"port_id_subtype":    types.StringType,
		"port_id":            types.StringType,
		"port_descr":         types.StringType,
		"sys_name":           types.StringType,
		"sys_descr":          types.StringType,
		"sys_ttl":            types.Int64Type,
		"port_source_mac":    types.StringType,
	}
}

func EClientLLDPNeighborAttributeValue(neighbors map[string]interface{}) map[string]attr.Value {
	hostneighborAid := types.StringNull()
	chassisIdSubtype := types.StringNull()
	chassisId := types.StringNull()
	sysName := types.StringNull()
	portId := types.StringNull()
	portDescr := types.StringNull()
	portSourceMAC := types.StringNull()
	portIdSubtype := types.StringNull()
	sysDescr := types.StringNull()
	sysTTL := types.Int64Null()

	for k, v := range neighbors {
		switch k {
		case "hostneighborAid":
			hostneighborAid = types.StringValue(v.(string))
		case "chassisIdSubtype":
			chassisIdSubtype = types.StringValue(v.(string))
		case "chassisId":
			chassisId = types.StringValue(v.(string))
		case "sysName":
			sysName = types.StringValue(v.(string))
		case "portIdSubtype":
			portIdSubtype = types.StringValue(v.(string))
		case "portSourceMAC":
			portSourceMAC = types.StringValue(v.(string))
		case "portDescr":
			portDescr = types.StringValue(v.(string))
		case "sysDescr":
			sysDescr = types.StringValue(v.(string))
		case "sysTTL":
			sysTTL = types.Int64Value(int64(v.(float64)))
		}
	}

	return map[string]attr.Value{
		"host_neighbor_aid":  hostneighborAid,
		"chassis_id_subtype": chassisIdSubtype,
		"chassis_id":         chassisId,
		"sys_name":           sysName,
		"port_id":            portId,
		"port_descr":         portDescr,
		"port_source_mac":    portSourceMAC,
		"port_id_subtype":    portIdSubtype,
		"sys_descr":          sysDescr,
		"sys_ttl":            sysTTL,
	}
}
