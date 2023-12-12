package host

import (
	"context"
	"encoding/json"
	"strings"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"

	//"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	//"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &HostPortResource{}
	_ resource.ResourceWithConfigure   = &HostPortResource{}
	_ resource.ResourceWithImportState = &HostPortResource{}
)

// NewModuleResource is a helper function to simplify the provider implementation.
func NewHostPortResource() resource.Resource {
	return &HostPortResource{}
}

type HostPortResource struct {
	client *ipm_pf.Client
}


type HostPortConfig struct {
	Name             types.String `tfsdk:"name"`
	ManagedBy        types.String `tfsdk:"managed_by"`
	Selector         common.IfSelector `tfsdk:"selector"`
	Labels           types.Map    `tfsdk:"labels"`
}

type HostPortResourceData struct {
	HostId        types.String `tfsdk:"host_id"`
	Id        types.String `tfsdk:"id"`
	Href      types.String `tfsdk:"href"`
	Config    HostPortConfig   `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
}


// Metadata returns the data source type name.
func (r *HostPortResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host_port"
}

// Schema defines the schema for the data source.
func (r *HostPortResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type HostPortResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an Module",
		Attributes:  HostPortSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *HostPortResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r HostPortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HostPortResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "NetworkResource: Create - ", map[string]interface{}{"NetworkResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)
	resp.State.Set(ctx, &data)

}

func (r HostPortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HostPortResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "HostPortResource: Read - ", map[string]interface{}{"HostPortResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r HostPortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HostPortResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "HostPortResource: Update", map[string]interface{}{"HostPortResourceData": data})

	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r HostPortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HostPortResourceData

	diags := req.State.Get(ctx, &data)
	diags.AddError(
		"HostPortResource: Error Delete Host Port",
		"Could not delete Host Port. Can get host port",
	)
	resp.Diagnostics.Append(diags...)
	resp.State.RemoveResource(ctx)
}

func (r *HostPortResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *HostPortResource) create(plan *HostPortResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "HostPortResource: create ## ", map[string]interface{}{"plan": plan})

	if plan.HostId.IsNull() {
		diags.AddError(
			"HostPortResource: Error create Host Port",
			"Could not create Hub Module. Host Id is not specified",
		)
	}

	var createRequest = make(map[string]interface{})

	// get Network config settings
	if !plan.Config.Name.IsNull() {
		createRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.ManagedBy.IsNull() {
		createRequest["managedBy"] = plan.Config.ManagedBy.ValueString()
	}
	if !plan.Config.Labels.IsNull() {
		labels := map[string]string{}
		diag := plan.Config.Labels.ElementsAs(ctx, &labels, true)
		if !diag.HasError() {
			createRequest["labels"] = labels
		}
	}
	var selector = make(map[string]interface{})
	aSelector := make(map[string]interface{})
	if plan.Config.Selector.ModuleIfSelectorByModuleId != nil {
		aSelector["moduleId"] = plan.Config.Selector.ModuleIfSelectorByModuleId.ModuleId.ValueString()
		aSelector["moduleClientIfAid"] = plan.Config.Selector.ModuleIfSelectorByModuleId.ModuleClientIfAid.ValueString()
		selector["ModuleIfSelectorByModuleId"] = aSelector
	} else if plan.Config.Selector.ModuleIfSelectorByModuleName != nil {
		aSelector["moduleName"] = plan.Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString()
		aSelector["moduleClientIfAid"] = plan.Config.Selector.ModuleIfSelectorByModuleId.ModuleClientIfAid.ValueString()
		selector["ModuleIfSelectorByModuleName"] = aSelector
	} else if plan.Config.Selector.ModuleIfSelectorByModuleMAC != nil {
		aSelector["moduleMAC"] = plan.Config.Selector.ModuleIfSelectorByModuleMAC.ModuleMAC.ValueString()
		aSelector["moduleClientIfAid"] = plan.Config.Selector.ModuleIfSelectorByModuleId.ModuleClientIfAid.ValueString()
		selector["ModuleIfSelectorByModuleMAC"] = aSelector
	} else if plan.Config.Selector.ModuleIfSelectorByModuleSerialNumber != nil {
		aSelector["moduleSerialNumber"] = plan.Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleSerialNumber.ValueString()
		aSelector["moduleClientIfAid"] = plan.Config.Selector.ModuleIfSelectorByModuleId.ModuleClientIfAid.ValueString()
		selector["ModuleIfSelectorByModuleSerialNumber"] = aSelector
	} else if plan.Config.Selector.HostPortSelectorByName != nil {
		aSelector["hostName"] = plan.Config.Selector.HostPortSelectorByName.HostName.ValueString()
		aSelector["hostPortName"] = plan.Config.Selector.HostPortSelectorByName.HostPortName.ValueString()
		selector["HostPortPortSelectorByName"] = aSelector
	} else if plan.Config.Selector.HostPortSelectorByPortId != nil {
		aSelector["chassisId"] = plan.Config.Selector.HostPortSelectorByPortId.ChassisId.ValueString()
		aSelector["chassisIdSubtype"] = plan.Config.Selector.HostPortSelectorByPortId.ChassisIdSubtype.ValueString()
		aSelector["portId"] = plan.Config.Selector.HostPortSelectorByPortId.PortId.ValueString()
		aSelector["portIdSubtype"] = plan.Config.Selector.HostPortSelectorByPortId.PortIdSubtype.ValueString()
		selector["hostPortSelectorByPortId"] = aSelector
	} else if plan.Config.Selector.HostPortSelectorBySysName != nil {
		aSelector["sysName"] = plan.Config.Selector.HostPortSelectorBySysName.SysName.ValueString()
		aSelector["portId"] = plan.Config.Selector.HostPortSelectorBySysName.PortId.ValueString()
		aSelector["portIdSubtype"] = plan.Config.Selector.HostPortSelectorBySysName.PortIdSubtype.ValueString()
		selector["hostPortSelectorBySysName"] = aSelector
	} else if plan.Config.Selector.HostPortSelectorByPortSourceMAC != nil {
		aSelector["portSourceMAC"] = plan.Config.Selector.HostPortSelectorByPortSourceMAC.PortSourceMAC.ValueString()
		selector["hostPortSelectorByPortSourceMAC"] = aSelector
	} else {
		diags.AddError(
			"Error Create HostPortResource",
			"Create: Could not create HostPortResource, No hub module selector specified",
		)
		return
	}
	createRequest["selector"] = selector

	tflog.Debug(ctx, "HostPortResource: create ## ", map[string]interface{}{"Create Request": createRequest})

	// send create request to server
	var request []map[string]interface{}
	request = append(request, createRequest)
	rb, err := json.Marshal(request)
	if err != nil {
		diags.AddError(
			"HostPortResource: create ##: Error Create AC",
			"Create: Could not Marshal HostPortResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/hosts/"+ plan.HostId.ValueString()+"/ports", rb)
	if err != nil {
		if !strings.Contains(err.Error(), "status: 202") {
			diags.AddError(
				"HostPortResource: create ##: Error create HostPortResource",
				"Create:Could not create HostPortResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "HostPortResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HostPortResource: Create ##: Error Unmarshal response",
			"Update:Could not Create HostPortResource, unexpected error: "+err.Error(),
		)
		return
	}

	result := data[0].(map[string]interface{})

	href := result["href"].(string)
	splits := strings.Split(href, "/")
	id := splits[len(splits)-1]
	plan.Href = types.StringValue(href)
	plan.Id = types.StringValue(id)

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "HostPortResource: create failed. Can't find the created network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "HostPortResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *HostPortResource) update(plan *HostPortResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "HostPortResource: update - plan", map[string]interface{}{"plan": plan})
	if plan.HostId.IsNull() || plan.Id.IsNull(){
		diags.AddError(
			"HostPortResource: Error Update HostPort",
			"Update: Could not Update HostPort. Host Id or Port Id is not specified",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get Network config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.ManagedBy.IsNull() {
		updateRequest["managedBy"] = plan.Config.ManagedBy.ValueString()
	}
	if !plan.Config.Labels.IsNull() {
		labels := map[string]string{}
		diag := plan.Config.Labels.ElementsAs(ctx, &labels, true)
		if !diag.HasError() {
			updateRequest["labels"] = labels
		}
	}

	// send Update request to server
	rb, err := json.Marshal(updateRequest)
	if err != nil {
		diags.AddError(
			"HostPortResource: Update ##: Error Update AC",
			"Update: Could not Marshal HostPortResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HostPortResource: update - rb", map[string]interface{}{"rb": rb})

	body, err := r.client.ExecuteIPMHttpCommand("PUT", "/hosts/"+ plan.HostId.ValueString()+"/ports/"+plan.Id.ValueString(), rb)

	if err != nil {
		diags.AddError(
			"HostPortResource: Update ##: Error Update HostPortResource",
			"Update:Could not Update HostPortResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HostPortResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HostPortResource: Update ##: Error Unmarshal HostPortResource",
			"Update:Could not Update HostPortResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HostPortResource: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": data})

	r.read(plan, ctx, diags)

	tflog.Debug(ctx, "HostPortResource: update SUCCESS ", map[string]interface{}{"plan": plan})
}

func (r *HostPortResource) read(state *HostPortResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.HostId.IsNull() || state.Id.IsNull() {
		diags.AddError(
			"HostPortResource: Error Read HostPort",
			"Could not Read, Host Id or HostPort ID is not specified.",
		)
		return
	}
	tflog.Debug(ctx, "HostPortResource: read ", map[string]interface{}{"state": state})

	body, err := r.client.ExecuteIPMHttpCommand("GET", "/hosts/"+ state.HostId.ValueString()+"/ports/"+state.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"HostPortResource: read ##: Error Update HostPortResource",
			"Update:Could not read HostPortResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "HostPortResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HostPortResource: read ##: Error Unmarshal response",
			"Update:Could not read HostPortResource, unexpected error: "+err.Error(),
		)
		return
	}

	// populate state
	HostPortData := data[0].(map[string]interface{})
	state.Populate(HostPortData, ctx, diags)

	tflog.Debug(ctx, "HostPortResource: read ## ", map[string]interface{}{"plan": state})
}

func (hpData *HostPortResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}

	tflog.Debug(ctx, "HostPortResourceData: populate ## ", map[string]interface{}{"computeFlag": computeFlag, "data": data})
	if computeFlag {
		if data["parentId"] != nil {
			hpData.HostId = types.StringValue(data["parentId"].(string))
		}
		hpData.Id = types.StringValue(data["id"].(string))
	}
	hpData.Href = types.StringValue(data["href"].(string))

	tflog.Debug(ctx, "HostPortResourceData: populate Config## ")
	// populate Config
	if data["config"] != nil {
		HostPortConfig := data["config"].(map[string]interface{})

		labels := types.MapNull(types.StringType)
		if HostPortConfig["labels"] != nil {
			data := make(map[string]attr.Value)
			for k, v := range HostPortConfig["labels"].(map[string]interface{}) {
				data[k] = types.StringValue(v.(string))
			}
			labels = types.MapValueMust(types.StringType, data)
		} 
		if !hpData.Config.Labels.IsNull() || computeFlag {
			hpData.Config.Labels = labels
		}
		for k, v := range HostPortConfig {
			switch k {
			case "name": 
				if !hpData.Config.Name.IsNull() || computeFlag {
					hpData.Config.Name = types.StringValue(v.(string))
				}
			case "managedBy": 
				if !hpData.Config.ManagedBy.IsNull() || computeFlag {
					hpData.Config.ManagedBy = types.StringValue(v.(string))
				}
			case "selector":
				common.IfSelectorPopulate(v.(map[string]interface{}), &hpData.Config.Selector, computeFlag)
			}
		}
	}
	tflog.Debug(ctx, "HostPortResourceData: populate State## ")
	// populate state
	if data["state"] != nil {
		hpData.State =types.ObjectValueMust(
			HostPortStateAttributeType(),HostPortStateAttributeValue(data["state"].(map[string]interface{})))
	}
	
	tflog.Debug(ctx, "HostPortResourceData: populate SUCCESS ")
}

func HostPortSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"host_id": schema.StringAttribute{
			Description: "Numeric identifier of the Host.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the port",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"href": schema.StringAttribute{
			Description: "href of the network module",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		//Config    NodeConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "name",
					Optional:    true,
				},
				"managed_by": schema.StringAttribute{
					Description: "managed_by",
					Optional:    true,
				},
				"selector": common.IfSelectorSchema(),
				"labels": schema.MapAttribute{
					Description: "labels",
					Optional:    true,
					ElementType: types.StringType,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: HostPortStateAttributeType(),
		},
	}
}


func HostPortStateAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"name":    types.StringType,
		"host_name":    types.StringType,
		"chassis_id_subtype" :types.StringType,
		"chassis_id"  : types.StringType,
		"sys_name" : types.StringType,
		"port_id_sub_type": types.StringType,
		"port_id"  : types.StringType,
		"port_source_mac": types.StringType,
		"port_descr": types.StringType,
		"managed_by" : types.StringType,
		"lldp_state": types.StringType,
		"module_if": types.ObjectType{AttrTypes: common.ModuleIfAttributeType()},
		"labels": types.MapType {ElemType:  types.StringType},
	}
}

func HostPortStateAttributeValue(hostPorState map[string]interface{}) (map[string]attr.Value) {
	name := types.StringNull()
	if hostPorState["name"] != nil {
		name = types.StringValue(hostPorState["name"].(string))
	}
	hostName := types.StringNull()
	if hostPorState["hosName"] != nil {
		hostName = types.StringValue(hostPorState["hostName"].(string))
	}
	chassisIdSubtype := types.StringNull()
	if hostPorState["chassisIdSubtype"] != nil {
		chassisIdSubtype = types.StringValue(hostPorState["chassisIdSubtype"].(string))
	}
	chassisId := types.StringNull()
	if hostPorState["chassisId"] != nil {
		chassisId = types.StringValue(hostPorState["chassisId"].(string))
	}
	sysName := types.StringNull()
	if hostPorState["sysName"] != nil {
		sysName = types.StringValue(hostPorState["sysName"].(string))
	}
	portIdSubtype := types.StringNull()
	if hostPorState["portIdSubtype"] != nil {
		portIdSubtype = types.StringValue(hostPorState["portIdSubtype"].(string))
	}
	portId := types.StringNull()
	if hostPorState["portId"] != nil {
		portId = types.StringValue(hostPorState["portId"].(string))
	}
	portSourceMAC := types.StringNull()
	if hostPorState["portSourceMAC"] != nil {
		portSourceMAC = types.StringValue(hostPorState["portSourceMAC"].(string))
	}
	portDescr := types.StringNull()
	if hostPorState["portDescr"] != nil {
		portDescr = types.StringValue(hostPorState["portDescr"].(string))
	}
	lldpState := types.StringNull()
	if hostPorState["lldpState"] != nil {
		lldpState = types.StringValue(hostPorState["lldpState"].(string))
	}
	managedBy := types.StringNull()
	if hostPorState["managedBy"] != nil {
		managedBy = types.StringValue(hostPorState["managedBy"].(string))
	}
	moduleIf := types.ObjectNull(common.ModuleIfAttributeType())
	if hostPorState["moduleIf"] != nil {
		moduleIf =  types.ObjectValueMust(common.ModuleIfAttributeType(),
		common.ModuleIfAttributeValue(hostPorState["moduleIf"].(map[string]interface{})))
	}
	labels := types.MapNull(types.StringType)
	if hostPorState["labels"] != nil {
		data := make(map[string]attr.Value)
		for k, v := range hostPorState["labels"].(map[string]interface{}) {
			data[k] = types.StringValue(v.(string))
		}
		labels = types.MapValueMust(types.StringType, data)
	}
	return map[string]attr.Value{
		"name": name,
		"host_name": hostName,
		"chassis_id_subtype": chassisIdSubtype,
		"chassis_id": chassisId,
		"sys_name": sysName,
		"port_id_sub_type": portIdSubtype,
		"port_id": portId,
		"port_source_mac": portSourceMAC,
		"port_descr": portDescr,
		"managed_by": managedBy,
		"lldp_state": lldpState,
		"module_if": moduleIf,
		"labels": labels,
	}
}

func HostPortObjectType() (types.ObjectType) {
	return types.ObjectType{	
					AttrTypes: HostPortAttributeType(),
				}	
}

func HostPortObjectsValue(data []interface{}) ([]attr.Value) {
	ports := []attr.Value{}
	for _, v := range data {
		port := v.(map[string]interface{})
		if port != nil {
			ports = append(ports, types.ObjectValueMust(
														HostPortAttributeType(),
														HostPortAttributeValue(port)))
		}
	}
	return ports
}

func HostPortAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"href":    types.StringType, 
		"id" : types.StringType,
		"config" : types.ObjectType{AttrTypes: HostPortConfigAttributeType()},
		"state" : types.ObjectType{AttrTypes: HostPortStateAttributeType()},
	}
}

func HostPortAttributeValue(hostPort map[string]interface{}) (map[string]attr.Value) {
	href := types.StringNull()
	if hostPort["href"] != nil {
		href = types.StringValue(hostPort["href"].(string))
	}
	id := types.StringNull()
	if hostPort["id"] != nil {
		id = types.StringValue(hostPort["id"].(string))
	}
	config := types.ObjectNull(HostPortConfigAttributeType())
	if hostPort["config"] != nil {
		config =  types.ObjectValueMust( HostPortConfigAttributeType(),
		HostPortConfigAttributeValue(hostPort["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(HostPortStateAttributeType())
	if hostPort["state"] != nil {
		state =  types.ObjectValueMust( HostPortStateAttributeType(),
		HostPortStateAttributeValue(hostPort["state"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"href": href,
		"id": id,
		"config": config,
		"state": state,
	}
}

func HostPortConfigAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"name":    types.StringType,
		"managed_by" : types.StringType,
		"labels": types.MapType {ElemType:  types.StringType},
		"selector": types.ObjectType{AttrTypes: common.IfSelectorAttributeType()},
	}
}

func HostPortConfigAttributeValue(hostPortConfig map[string]interface{}) (map[string]attr.Value) {
	name := types.StringNull()
	if hostPortConfig["name"] != nil {
		name = types.StringValue(hostPortConfig["name"].(string))
	}
	managedBy := types.StringNull()
	if hostPortConfig["managedBy"] != nil {
		managedBy = types.StringValue(hostPortConfig["managedBy"].(string))
	}
	labels := types.MapNull(types.StringType)
	if hostPortConfig["labels"] != nil {
		data := make(map[string]attr.Value)
		for k, v := range hostPortConfig["labels"].(map[string]interface{}) {
			data[k] = types.StringValue(v.(string))
		}
		labels, _ = types.MapValue(types.StringType, data)
	}
	selector := types.ObjectNull(common.IfSelectorAttributeType())
	if hostPortConfig["selector"] != nil {
		selector =  types.ObjectValueMust( common.IfSelectorAttributeType(),
		common.IfSelectorAttributeValue(hostPortConfig["selector"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"name": name,
		"managed_by": managedBy,
		"labels": labels,
		"selector": selector,
	}
}
