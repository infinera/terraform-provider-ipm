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
	_ resource.Resource                = &ModuleResource{}
	_ resource.ResourceWithConfigure   = &ModuleResource{}
	_ resource.ResourceWithImportState = &ModuleResource{}
)

// NewModuleResource is a helper function to simplify the provider implementation.
func NewModuleResource() resource.Resource {
	return &ModuleResource{}
}

type ModuleResource struct {
	client *ipm_pf.Client
}

type ModuleConfig struct {
	ModuleName types.String `tfsdk:"module_name"`
	Labels     types.Map    `tfsdk:"labels"`
	MVLANMode  types.String `tfsdk:"m_vlan_mode"`
	DebugPortAccess  types.String `tfsdk:"debug_port_access"`
}

type ModuleResourceData struct {
	Identifier      common.DeviceIdentifier `tfsdk:"identifier"`
	Id              types.String `tfsdk:"id"`
	Href            types.String `tfsdk:"href"`
	Config          *ModuleConfig `tfsdk:"config"`
	State           types.Object `tfsdk:"state"`
	LinePtps        types.List   `tfsdk:"line_ptps"`
	OTUs            types.List   `tfsdk:"otus"`
	EthernetClients types.List   `tfsdk:"ethernet_clients"`
	LCs             types.List   `tfsdk:"lcs"`
}

// Metadata returns the data source type name.
func (r *ModuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module"
}

// Schema defines the schema for the data source.
func (r *ModuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type ModuleResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an XR Module",
		Attributes:  ModuleResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *ModuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r ModuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ModuleResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "ModuleResource: Create - ", map[string]interface{}{"ModuleResourceData": data})

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

func (r ModuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ModuleResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "ModuleResource: Create - ", map[string]interface{}{"ModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r ModuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ModuleResourceData
	var stateData ModuleResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "ModuleResource: Update", map[string]interface{}{"ModuleResourceData": data})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &stateData)
	tflog.Debug(ctx, "ModuleResource: Update", map[string]interface{}{"ModuleResourceData": data})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

  if data.Id.IsNull() || data.Id.ValueString() == "" {
		data.Id = stateData.Id
	}

	if data.Href.IsNull() || data.Href.ValueString() == "" {
		data.Href = stateData.Href
	}
	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r ModuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ModuleResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "ModuleResource: Update", map[string]interface{}{"ModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	//r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ModuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ModuleResource) update(plan *ModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "ModuleResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && plan.Id.IsNull() && plan.Identifier.DeviceId.IsNull()  {
		diags.AddError(
			"ModuleResource: Error update module",
			"ModuleResource: Could not update module. Resource Href, Id adn Identifier are not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get EClientResource config settings
	if !plan.Config.ModuleName.IsNull() {
		updateRequest["moduleName"] = plan.Config.ModuleName.ValueString()
	}
	if !plan.Config.DebugPortAccess.IsNull() {
		updateRequest["debugPortAccess"] = plan.Config.DebugPortAccess.ValueString()
	}
	if !plan.Config.MVLANMode.IsNull() {
		updateRequest["mvlanMode"] = plan.Config.MVLANMode.ValueString()
	}
	if !plan.Config.Labels.IsNull() {
		labels := map[string]string{}
		diag := plan.Config.Labels.ElementsAs(ctx, &labels, true)
		if !diag.HasError() {
			updateRequest["labels"] = labels
		}
	}

	tflog.Debug(ctx, "ModuleResource: update ## ", map[string]interface{}{"Update Request": updateRequest, "ID": plan.Id.ValueString(), "href": plan.Href.ValueString()})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"ModuleResource: update ##: Error Create Module",
				"Update: Could not Marshal Module, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Id.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/modules/" + plan.Id.ValueString(), rb)
		} else {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/modules/" + plan.Identifier.DeviceId.ValueString(), rb)
		}
		if err != nil {
			if !strings.Contains(err.Error(), "status: 202") {
				diags.AddError(
					"ModuleResource: update ##: Error update Module",
					"Update: Could not update Module, unexpected error: "+err.Error(),
				)
				return
			}
		}

		tflog.Debug(ctx, "ModuleResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"ModuleResource: Update ##: Error Unmarshal response",
				"Update: Could not Create Module, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "ModuleResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "ModuleResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *ModuleResource) read(state *ModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Href.IsNull() && state.Id.IsNull() && state.Identifier.DeviceId.IsNull()  {
		diags.AddError(
			"Error Read ModuleResource",
			"ModuleResource: Could not read. Id, Href and Identifier are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "ModuleResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/modules/"+state.Id.ValueString() + queryStr
	} else {
		queryStr = "/modules/"+ state.Identifier.DeviceId.ValueString() + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)

	if err != nil {
		diags.AddError(
			"ModuleResource: read ##: Error Read ModuleResource",
			"Read:Could not get ModuleResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "ModuleResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	//var data = make(map[string]interface{})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"ModuleResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal ModuleResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
		case []interface{}:
			if len(resp.([]interface{})) > 0 {
				state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
			} else {
				diags.AddError(
					"ModuleResource: read ##: Can not get Module",
					"Read:Could not get Module for query: "+ queryStr,
				)
				return
			}
		case map[string]interface{}:
			state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "ModuleResource: read ## ", map[string]interface{}{"plan": state})
}


func (moduleData *ModuleResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "ModuleResourceData: populate ## ", map[string]interface{}{"plan": data})

	moduleData.Href = types.StringValue(data["href"].(string))
	moduleData.Id = types.StringValue(data["id"].(string))
	// populate config
	var config = data["config"].(map[string]interface{})
	if moduleData.Config == nil {
		moduleData.Config = &ModuleConfig{}
	}
	for k, v := range config {
		switch k {
		case "moduleName":
			if !moduleData.Config.ModuleName.IsNull() {
				moduleData.Config.ModuleName = types.StringValue(v.(string))
			}
		case "debugPortAccess":
			if !moduleData.Config.DebugPortAccess.IsNull() {
				moduleData.Config.DebugPortAccess = types.StringValue(v.(string))
			}
		case "mvlanMode":
			if !moduleData.Config.MVLANMode.IsNull() {
				moduleData.Config.MVLANMode = types.StringValue(v.(string))
			}
		case "labels":
			labels := types.MapNull(types.StringType)
			data := make(map[string]attr.Value)
			for k, label := range v.(map[string]interface{}) {
				data[k] = types.StringValue(label.(string))
			}
			labels = types.MapValueMust(types.StringType, data)
			if !moduleData.Config.Labels.IsNull() {
				moduleData.Config.Labels = labels
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		moduleData.State = types.ObjectValueMust(ModuleStateAttributeType(), ModuleStateAttributeValue(state))
	}
	// populate LinePtps
	moduleData.LinePtps = types.ListNull(LinePTPObjectType())
	if data["linePtps"] != nil {
		moduleData.LinePtps = types.ListValueMust(LinePTPObjectType(), LinePTPObjectsValue(data["linePtps"].([]interface{})))
	}
	moduleData.OTUs = types.ListNull(OTUObjectType())
	if data["otus"] != nil {
		moduleData.OTUs = types.ListValueMust(OTUObjectType(), OTUObjectsValue(data["otus"].([]interface{})))
	}
	moduleData.EthernetClients = types.ListNull(EClientObjectType())
	if data["ethernetClients"] != nil {
		moduleData.EthernetClients = types.ListValueMust(EClientObjectType(), EClientObjectsValue(data["ethernetClients"].([]interface{})))
	}
	moduleData.LCs = types.ListNull(LCObjectType())
	if data["localConnections"] != nil {
		moduleData.LCs = types.ListValueMust(LCObjectType(), LCObjectsValue(data["localConnections"].([]interface{})))
	}

	tflog.Debug(ctx, "ModuleResourceData: read ## ", map[string]interface{}{"plan": state})
}

func ModuleResourceSchemaAttributes(computeEntity_optional ...bool) map[string]schema.Attribute {
	computeFlag := false
	optionalFlag := true
	if len(computeEntity_optional) > 0 {
		computeFlag = computeEntity_optional[0]
		optionalFlag = !computeFlag
	}
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID of the Module.",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		"identifier": common.DeviceIdentifierAttribute(),
		"config": schema.SingleNestedAttribute{
			Description: "Module Config",
			Computed:    computeFlag,
			Optional:    optionalFlag,
			Attributes: map[string]schema.Attribute{
				"module_name": schema.StringAttribute{
					Description: "module_name",
					Optional:    true,
				},
				"labels": schema.MapAttribute{
					Description: "labels",
					Optional:    true,
					ElementType: types.StringType,
				},
				"m_vlan_mode": schema.StringAttribute{
					Description: "m_vlan_mode",
					Optional:    true,
				},
				"debug_port_access": schema.StringAttribute{
					Description: "debug_port_access",
					Optional:    true,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: ModuleStateAttributeType(),
		},
		"line_ptps": schema.ListAttribute{
			Computed:    true,
			ElementType: LinePTPObjectType(),
		},
		"otus": schema.ListAttribute{
			Computed:    true,
			ElementType: OTUObjectType(),
		},
		"lcs": schema.ListAttribute{
			Computed:    true,
			ElementType: LCObjectType(),
		},
		"ethernet_clients": schema.ListAttribute{
			Computed:    true,
			ElementType: EClientObjectType(),
		},
	}
}

func ModuleObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: ModuleAttributeType(),
	}
}

func ModuleObjectsValue(data []interface{}) []attr.Value {
	modules := []attr.Value{}
	for _, v := range data {
		module := v.(map[string]interface{})
		if module != nil {
			modules = append(modules, types.ObjectValueMust(
				ModuleAttributeType(),
				ModuleAttributeValue(module)))
		}
	}
	return modules
}

func ModuleAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"href":             types.StringType,
		"config":           types.ObjectType{AttrTypes: ModuleConfigAttributeType()},
		"state":            types.ObjectType{AttrTypes: ModuleStateAttributeType()},
		"otus":             types.ListType{ElemType: OTUObjectType()},
		"line_ptps":        types.ListType{ElemType: LinePTPObjectType()},
		"lcs":              types.ListType{ElemType: LCObjectType()},
		"ethernet_clients": types.ListType{ElemType: EClientObjectType()},
	}
}

func ModuleAttributeValue(module map[string]interface{}) map[string]attr.Value {
	href := types.StringNull()
	if module["href"] != nil {
		moduleHref := module["href"].(string)
		href = types.StringValue(moduleHref)
	}
	id := types.StringNull()
	if module["id"] != nil {
		id = types.StringValue(module["id"].(string))
	}
	config := types.ObjectNull(ModuleConfigAttributeType())
	if (module["config"]) != nil {
		config = types.ObjectValueMust(ModuleConfigAttributeType(), ModuleConfigAttributeValue(module["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(ModuleStateAttributeType())
	if (module["state"]) != nil {
		state = types.ObjectValueMust(ModuleStateAttributeType(), ModuleStateAttributeValue(module["state"].(map[string]interface{})))
	}
	otus := types.ListNull(OTUObjectType())
	if (module["otus"]) != nil {
		otus = types.ListValueMust(OTUObjectType(), OTUObjectsValue(module["otus"].([]interface{})))
	}
	linePTPs := types.ListNull(LinePTPObjectType())
	if (module["linePTPs"]) != nil {
		linePTPs = types.ListValueMust(LinePTPObjectType(), LinePTPObjectsValue(module["linePTPs"].([]interface{})))
	}
	lcs := types.ListNull(LCObjectType())
	if (module["localConnections"]) != nil {
		lcs = types.ListValueMust(LCObjectType(), LCObjectsValue(module["localConnections"].([]interface{})))
	}
	ethernetClients := types.ListNull(EClientObjectType())
	if (module["ethernetClients"]) != nil {
		ethernetClients = types.ListValueMust(EClientObjectType(), EClientObjectsValue(module["ethernetClients"].([]interface{})))
	}

	return map[string]attr.Value{
		"id":               id,
		"href":             href,
		"config":           config,
		"state":            state,
		"otus":             otus,
		"lcs":              lcs,
		"line_ptps":        linePTPs,
		"ethernet_clients": ethernetClients,
	}
}
func ModuleConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_name": types.StringType,
		"labels":      types.MapType{ElemType: types.StringType},
		"m_vlan_mode": types.StringType,
		"debug_port_access": types.StringType,
	}
}

func ModuleConfigAttributeValue(moduleConfig map[string]interface{}) map[string]attr.Value {
	moduleName := types.StringNull()
	if moduleConfig["moduleName"] != nil {
		moduleName = types.StringValue(moduleConfig["moduleName"].(string))
	}
	mvlanMode := types.StringNull()
	if moduleConfig["mvlanMode"] != nil {
		mvlanMode = types.StringValue(moduleConfig["mvlanMode"].(string))
	}
	debugPortAccess := types.StringNull()
	if moduleConfig["debugPortAccess"] != nil {
		debugPortAccess = types.StringValue(moduleConfig["debugPortAccess"].(string))
	}

	labels := types.MapNull(types.StringType)
	if moduleConfig["labels"] != nil {
		data := make(map[string]attr.Value)
		for k, v := range moduleConfig["labels"].(map[string]interface{}) {
			data[k] = types.StringValue(v.(string))
		}
		labels = types.MapValueMust(types.StringType, data)
	}

	return map[string]attr.Value{
		"module_name": moduleName,
		"labels":      labels,
		"m_vlan_mode": mvlanMode,
		"debug_port_access": debugPortAccess,
	}
}

func ModuleStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_aid":           types.StringType,
		"module_name":          types.StringType,
		"m_vlan_mode":          types.StringType,
		"debug_port_access":    types.StringType,
		"labels":               types.MapType{ElemType: types.StringType},
		"hid":                  types.StringType,
		"hport_id":             types.StringType,
		"configured_role":      types.StringType,
		"current_role":         types.StringType,
		"role_status":          types.StringType,
		"traffic_mode":         types.StringType,
		"topology":             types.StringType,
		"config_state":         types.StringType,
		"tc_mode":              types.BoolType,
		"life_cycle_state":     types.StringType,
		"serdes_rate":          types.StringType,
		"connectivity_state":   types.StringType,
		"restart_action":       types.StringType,
		"factory_reset_action": types.BoolType,
		"hw_description":       types.ObjectType{AttrTypes: ModuleHWDescriptionAttributeType()},
	}
}

func ModuleStateAttributeValue(moduleState map[string]interface{}) map[string]attr.Value {
	moduleAid := types.StringNull()
	moduleName := types.StringNull()
	mVlanMode := types.StringNull()
	debugPortAccess := types.StringNull()
	labels := types.MapNull(types.StringType)
	hid := types.StringNull()
	hportId := types.StringNull()
	configuredRole := types.StringNull()
	roleStatus := types.StringNull()
	currentRole := types.StringNull()
	trafficMode := types.StringNull()
	topology := types.StringNull()
	configState := types.StringNull()
	tcMode := types.BoolNull()
	lifeCycleState := types.StringNull()
	serdesRate := types.StringNull()
	connectivityState := types.StringNull()
	restartAction := types.StringNull()
	factoryResetAction := types.BoolNull()
	hwDescription := types.ObjectNull(ModuleHWDescriptionAttributeType())

	for k, v := range moduleState {
		switch k {
		case "moduleAid":
			moduleAid = types.StringValue(v.(string))
		case "moduleName":
			moduleName = types.StringValue(v.(string))
		case "mvlanMode":
			mVlanMode = types.StringValue(v.(string))
		case "debugPortAccess":
			debugPortAccess = types.StringValue(v.(string))
		case "labels":
			data := make(map[string]attr.Value)
			for k, label := range v.(map[string]interface{}) {
				data[k] = types.StringValue(label.(string))
			}
			labels = types.MapValueMust(types.StringType, data)
		case "hid":
			hid = types.StringValue(v.(string))
		case "hportId":
			hportId = types.StringValue(v.(string))
		case "configuredRole":
			configuredRole = types.StringValue(v.(string))
		case "roleStatus":
			roleStatus = types.StringValue(v.(string))
		case "currentRole":
			currentRole = types.StringValue(v.(string))
		case "trafficMode":
			trafficMode = types.StringValue(v.(string))
		case "topology":
			topology = types.StringValue(v.(string))
		case "configState":
			configState = types.StringValue(v.(string))
		case "tcMode":
			tcMode = types.BoolValue(v.(bool))
		case "lifeCycleState":
			lifeCycleState = types.StringValue(v.(string))
		case "serdesRate":
			serdesRate = types.StringValue(v.(string))
		case "connectivityState":
			connectivityState = types.StringValue(v.(string))
		case "restartAction":
			restartAction = types.StringValue(v.(string))
		case "factoryResetAction":
			factoryResetAction = types.BoolValue(v.(bool))
		case "hwDescription":
			hwDescription = types.ObjectValueMust(ModuleHWDescriptionAttributeType(), ModuleHWDescriptionAttributeValue(v.(map[string]interface{})))
		}
	}

	return map[string]attr.Value{
		"module_aid":           moduleAid,
		"module_name":          moduleName,
		"m_vlan_mode":          mVlanMode,
		"debug_port_access":    debugPortAccess,
		"labels":               labels,
		"hid":                  hid,
		"hport_id":             hportId,
		"configured_role":      configuredRole,
		"current_role":         currentRole,
		"role_status":          roleStatus,
		"traffic_mode":         trafficMode,
		"topology":             topology,
		"config_state":         configState,
		"tc_mode":              tcMode,
		"life_cycle_state":     lifeCycleState,
		"serdes_rate":          serdesRate,
		"connectivity_state":   connectivityState,
		"restart_action":       restartAction,
		"factory_reset_action": factoryResetAction,
		"hw_description":       hwDescription,
	}
}

func ModuleHWDescriptionAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"pi":             types.StringType,
		"mnfv":           types.StringType,
		"mnmn":           types.StringType,
		"mnmo":           types.StringType,
		"mnhw":           types.StringType,
		"mndt":           types.StringType,
		"serial_number":  types.StringType,
		"clei":           types.StringType,
		"mac_address":    types.StringType,
		"connector_type": types.StringType,
		"form_factor":    types.StringType,
		"piid":           types.StringType,
		"dmn":            types.ObjectType{AttrTypes: DMNAttributeType()},
		"sv":             types.StringType,
		"capabilities":   types.MapType{ElemType: types.StringType},
	}
}

func ModuleHWDescriptionAttributeValue(moduleHWDescription map[string]interface{}) map[string]attr.Value {
	pi := types.StringNull()
	mnfv := types.StringNull()
	mnmn := types.StringNull()
	mnmo := types.StringNull()
	mnhw := types.StringNull()
	mndt := types.StringNull()
	serialNumber := types.StringNull()
	clei := types.StringNull()
	macAddress := types.StringNull()
	connectorType := types.StringNull()
	formFactor := types.StringNull()
	piid := types.StringNull()
	dmn := types.ObjectNull(DMNAttributeType())
	sv := types.StringNull()
	capabilities := types.MapNull(types.StringType)
	for k, v := range moduleHWDescription {
		switch k {
		case "pi":
			pi = types.StringValue(v.(string))
		case "mnfv":
			mnfv = types.StringValue(v.(string))
		case "mnmn":
			mnmn = types.StringValue(v.(string))
		case "mnmo":
			mnmo = types.StringValue(v.(string))
		case "mnhw":
			mnhw = types.StringValue(v.(string))
		case "mndt":
			mndt = types.StringValue(v.(string))
		case "serialNumber":
			serialNumber = types.StringValue(v.(string))
		case "clei":
			clei = types.StringValue(v.(string))
		case "macAddress":
			macAddress = types.StringValue(v.(string))
		case "connectorType":
			connectorType = types.StringValue(v.(string))
		case "formFactor":
			formFactor = types.StringValue(v.(string))
		case "sv":
			sv = types.StringValue(v.(string))
		case "piid":
			piid = types.StringValue(v.(string))
		case "dmn":
			dmn = types.ObjectValueMust(DMNAttributeType(), DMNAttributeValue(v.(map[string]interface{})))
		case "capabilities":
			capabilities = types.MapValueMust(types.StringType, common.MapAttributeValue(v.(map[string]interface{})))
		}
	}
	return map[string]attr.Value{
		"pi":             pi,
		"mnfv":           mnfv,
		"mnmn":           mnmn,
		"mnmo":           mnmo,
		"mnhw":           mnhw,
		"mndt":           mndt,
		"serial_number":  serialNumber,
		"clei":           clei,
		"mac_address":    macAddress,
		"connector_type": connectorType,
		"form_factor":    formFactor,
		"sv":             sv,
		"piid":           piid,
		"dmn":            dmn,
		"capabilities":   capabilities,
	}
}

func DMNAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"language":           types.StringType,
		"value":          types.StringType,
	}
}

func DMNAttributeValue(dmn map[string]interface{}) map[string]attr.Value {
	language := types.StringNull()
	value := types.StringNull()
	for k, v := range dmn {
		switch k {
		case "language":
			language = types.StringValue(v.(string))
		case "value":
			value = types.StringValue(v.(string))
		}
	}
	return map[string]attr.Value{
		"language":      language,
		"value":         value,
	}
}
