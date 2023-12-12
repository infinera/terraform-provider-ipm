package networkconnection

import (
	"context"
	"encoding/json"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	//"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &NCEndpointResource{}
	_ resource.ResourceWithConfigure   = &NCEndpointResource{}
	_ resource.ResourceWithImportState = &NCEndpointResource{}
)

// NewNCEndpointResource is a helper function to simplify the provider implementation.
func NewNCEndpointResource() resource.Resource {
	return &NCEndpointResource{}
}

type NCEndpointResource struct {
	client *ipm_pf.Client
}

type NCEndpointConfig struct {
	Selector common.IfSelector  `tfsdk:"selector"`
	Capacity types.Int64 `tfsdk:"capacity"` // 100, 400
}


type NCEndpointResourceData struct {
	NCId   types.String     `tfsdk:"nc_id"`
	Id     types.String     `tfsdk:"id"`
	Href   types.String     `tfsdk:"href"`
	Config NCEndpointConfig `tfsdk:"config"`
	State  types.Object     `tfsdk:"state"`
	ACs    types.List       `tfsdk:"acs"`
}

// Metadata returns the data source type name.
func (r *NCEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nc_endpoint"
}

// Schema defines the schema for the data source.
func (r *NCEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type NCEndpointResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an NC endpoint",
		Attributes:  NCEndpointSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *NCEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r NCEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NCEndpointResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "NCEndpointResource: Create - ", map[string]interface{}{"NCEndpointResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r NCEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NCEndpointResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "NCEndpointResource: Create - ", map[string]interface{}{"NCEndpointResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r NCEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NCEndpointResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"NCEndpointResourceData": data})

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

func (r NCEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NCEndpointResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"NCEndpointResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *NCEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NCEndpointResource) create(plan *NCEndpointResourceData, ctx context.Context, diags *diag.Diagnostics) {

	plan.Create(r.client, ctx, diags)

	if !diags.HasError() {
		r.read(plan, ctx, diags)
	}

	tflog.Debug(ctx, "NCEndpointResource: create ##", map[string]interface{}{"plan": plan})
}

func (plan *NCEndpointResourceData) Create(client *ipm_pf.Client, ctx context.Context, diags *diag.Diagnostics) {

	if plan.NCId.IsNull() || plan.Config.Capacity.IsNull() {
		diags.AddError(
			"NCEndpointResourceData: Error Create NCEndpoint",
			"Create: Could not create NCEndpointResourceData, NC Id or Endpoint Capacity is not specified",
		)
		return
	}

	var createRequest = make(map[string]interface{})
	createRequest["capacity"] = plan.Config.Capacity.ValueInt64()
	// get Module setting
	var selector = make(map[string]interface{})
	aSelector := make(map[string]interface{})
	if plan.Config.Selector.ModuleIfSelectorByModuleId != nil {
		aSelector["moduleId"] = plan.Config.Selector.ModuleIfSelectorByModuleId.ModuleId.ValueString()
		aSelector["moduleClientIfAid"] = plan.Config.Selector.ModuleIfSelectorByModuleId.ModuleClientIfAid.ValueString()
		selector["ModuleIfSelectorByModuleId"] = aSelector
	} else if plan.Config.Selector.ModuleIfSelectorByModuleName != nil {
		aSelector["moduleName"] = plan.Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString()
		aSelector["moduleClientIfAid"] = plan.Config.Selector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString()
		selector["ModuleIfSelectorByModuleName"] = aSelector
	} else if plan.Config.Selector.ModuleIfSelectorByModuleMAC != nil {
		aSelector["moduleMAC"] = plan.Config.Selector.ModuleIfSelectorByModuleMAC.ModuleMAC.ValueString()
		aSelector["moduleClientIfAid"] = plan.Config.Selector.ModuleIfSelectorByModuleMAC.ModuleClientIfAid.ValueString()
		selector["ModuleIfSelectorByModuleMAC"] = aSelector
	} else if plan.Config.Selector.ModuleIfSelectorByModuleSerialNumber != nil {
		aSelector["moduleSerialNumber"] = plan.Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleSerialNumber.ValueString()
		aSelector["moduleClientIfAid"] = plan.Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleClientIfAid.ValueString()
		selector["ModuleIfSelectorByModuleSerialNumber"] = aSelector
	} else if plan.Config.Selector.HostPortSelectorByName != nil {
		aSelector["hostName"] = plan.Config.Selector.HostPortSelectorByName.HostName.ValueString()
		aSelector["hostPortName"] = plan.Config.Selector.HostPortSelectorByName.HostPortName.ValueString()
		selector["hostPortSelectorByName"] = aSelector
	} else if plan.Config.Selector.HostPortSelectorByPortId != nil {
		aSelector["chassisId"] = plan.Config.Selector.HostPortSelectorByPortId.ChassisId.ValueString()
		aSelector["chassisIdSubtype"] = plan.Config.Selector.HostPortSelectorByPortId.ChassisIdSubtype.ValueString()
		aSelector["portId"] = plan.Config.Selector.HostPortSelectorByPortId.PortId.ValueString()
		aSelector["portIdSubtype"] = plan.Config.Selector.HostPortSelectorByPortId.PortIdSubtype.ValueString()
		selector["hostPortSelectorByPortId"] = aSelector
	} else if plan.Config.Selector.HostPortSelectorBySysName != nil {
		aSelector["sysName"] = plan.Config.Selector.HostPortSelectorBySysName.SysName.ValueString()
		aSelector["portId"] = plan.Config.Selector.HostPortSelectorByPortId.PortId.ValueString()
		aSelector["portIdSubtype"] = plan.Config.Selector.HostPortSelectorByPortId.PortIdSubtype.ValueString()
		selector["hostPortSelectorBySysName"] = aSelector
	} else if plan.Config.Selector.HostPortSelectorByPortSourceMAC != nil {
		aSelector["portSourceMAC"] = plan.Config.Selector.HostPortSelectorByPortSourceMAC.PortSourceMAC.ValueString()
		selector["hostPortSelectorByPortSourceMAC"] = aSelector
	} else {
		diags.AddError(
			"NCEndpointResourceData: Error Create NC Endpoint",
			"Create: Could not create NCEndpointResource, No IfSelector specified",
		)
		return
	}
	createRequest["selector"] = selector

	if !plan.Config.Capacity.IsNull() {
		createRequest["capacity"] = plan.Config.Capacity.ValueInt64()
	}

	tflog.Debug(ctx, "NCEndpointResourceData: create ## ", map[string]interface{}{"Create Request": createRequest})

	// send create request to server
	rb, err := json.Marshal(createRequest)
	if err != nil {
		diags.AddError(
			"NCEndpointResourceData: create ##: Error Create AC",
			"Create: Could not Marshal NCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := client.ExecuteIPMHttpCommand("POST", "/network-connections/"+plan.NCId.ValueString()+"/endpoints", rb)
	if err != nil {
		diags.AddError(
			"NCEndpointResourceData: create ##: Error create NCEndpoint",
			"Create:Could not create NCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "NCEndpointResourceData: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})

	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NCEndpointResourceData: read ##: Error Unmarshal response",
			"Update:Could not Create NCEndpoint, unexpected error: "+err.Error(),
		)
		return
	}

	var content = data["content"].(map[string]interface{})
	plan.Id = types.StringValue(content["id"].(string))

	plan.Href = types.StringValue(content["href"].(string))

	tflog.Debug(ctx, "NCEndpointResourceData: create ##", map[string]interface{}{"plan": plan})
}

func (r *NCEndpointResource) update(plan *NCEndpointResourceData, ctx context.Context, diags *diag.Diagnostics) {

	plan.Update(r.client, ctx, diags)

	if !diags.HasError() {
		r.read(plan, ctx, diags)
	}

	tflog.Debug(ctx, "NCEndpointResource: create ##", map[string]interface{}{"plan": plan})
}
func (plan *NCEndpointResourceData) Update(client *ipm_pf.Client, ctx context.Context, diags *diag.Diagnostics) {

	if plan.NCId.IsNull() || plan.Id.IsNull() {
		diags.AddError(
			"NCEndpointResourceData: Error Update NCEndpoint",
			"Update: Could not Update NCEndpoint. NC ID or NC endpoint ID is not specified",
		)
		return
	}

	var updateRequest = make(map[string]interface{})
	if !plan.Config.Capacity.IsNull() {
		updateRequest["capacity"] = plan.Config.Capacity.ValueInt64()
	}

	// send Update request to server
	rb, err := json.Marshal(updateRequest)
	if err != nil {
		diags.AddError(
			"NCEndpointResourceData: Update ##: Error Update",
			"Update: Could not Marshal NCEndpoint , unexpected error: "+err.Error(),
		)
		return
	}

	body, err := client.ExecuteIPMHttpCommand("PUT", "/network-connections/"+plan.NCId.ValueString()+"/endpoints/"+plan.Id.ValueString(), rb)
	if err != nil {
		diags.AddError(
			"NCEndpointResourceData: Update ##: Error Update NCEndpointResource",
			"Update:Could not Update NCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "NCEndpointResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})

	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NCEndpointResourceData: Update ##: Error Unmarshal NCEndpointResource",
			"Update:Could not Update NCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "NCEndpointResourceData: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": data, "plan": plan})

}

func (r *NCEndpointResource) read(state *NCEndpointResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.NCId.IsNull() || state.Id.IsNull() {
		diags.AddError(
			"NCEndpointResource: Error Get",
			"Update: Could not Update Module. NCID and either NCEndpoint ID or Href are not specified.",
		)
		return
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", "/network-connections/"+state.NCId.ValueString()+"/endpoints/"+state.Id.ValueString()+"?content=expanded", nil)
	if err != nil {
		diags.AddError(
			"NCEndpointResource: Error Get NCEndpointResource",
			"Update:Could not Get NCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "NCEndpointResource: Get ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})

	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NCEndpointResource: Get ##: Error Unmarshal NCEndpointResource",
			"Update:Could not Get NCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "NCEndpointResource: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": data})
	state.Populate(data, ctx, diags)

	tflog.Debug(ctx, "NCEndpointResource: read ## ", map[string]interface{}{"plan": state})
}

func (r *NCEndpointResource) delete(plan *NCEndpointResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.NCId.IsNull() || (plan.Id.IsNull() && plan.Href.IsNull()) {
		diags.AddError(
			"NCEndpointResource: Error Delete",
			"Update: Could not Delete NC Endpoint. NCID and either NCEndpoint ID or Href are not specified.",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", "/network-connections/"+plan.NCId.ValueString()+"/endpoints/"+plan.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"NCEndpointResource: delete ##: Error Delete NCEndpointResource",
			"Update:Could not delete NCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "NCEndpointResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (ep *NCEndpointResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}

	tflog.Debug(ctx, "NCEndpointResourceData: populate ## ", map[string]interface{}{"computeFlag": computeFlag, "data": data})
	if computeFlag {
		ep.Id = types.StringValue(data["id"].(string))
		ep.NCId = types.StringValue(data["parentId"].(string))
	}
	ep.Href = types.StringValue(data["href"].(string))
	// populate Config
	if data["config"] != nil {
		config := data["config"].(map[string]interface{})
		for k, v := range config {
			switch k {
			case "capacity":
				if !ep.Config.Capacity.IsNull() || computeFlag {
					ep.Config.Capacity = types.Int64Value(int64(v.(float64)))
				}
			case "selector":
				selectorData := v.(map[string]interface{})
				common.IfSelectorPopulate(selectorData, &ep.Config.Selector, computeFlag)
			}
		}
	}

	// popuplate state
	if data["state"] != nil {
		ep.State = types.ObjectValueMust(
			NCEndpointStateAttributeType(), NCEndpointStateAttributeValue(data["state"].(map[string]interface{})))
	}

	// populate ACs
	ep.ACs = types.ListNull(ACObjectType())
	if data["acs"] != nil {
		ep.ACs = types.ListValueMust(ACObjectType(), ACObjectsValue(data["acs"].([]interface{})))
	}
}

func NCEndpointSchemaAttributes() map[string]schema.Attribute {

	return map[string]schema.Attribute{
		"nc_id": schema.StringAttribute{
			Description: "Identifier of the Network Connection",
			Computed:    true,
		},
		"id": schema.StringAttribute{
			Description: "Identifier of the Network Connection's Endpoint",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href  of the Network Connection's Endpoint",
			Computed:    true,
		},
		//Config           LCConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Description: "Network Connection LC Config",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"selector": common.IfSelectorSchema(),
				"capacity": schema.Int64Attribute{
					Description: "capacity",
					Optional:    true,
				},
			},
		},
		//State           LCConfig `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: NCEndpointStateAttributeType(),
		},
		"acs":schema.ListAttribute{
			Computed: true,
			ElementType: ACObjectType(),
		},
	}
}

func NCEndpointObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: NCEndpointAttributeType(),
	}
}

func NCEndpointObjectsValue(data []interface{}) []attr.Value {
	endpoints := []attr.Value{}
	for _, v := range data {
		endpoint := v.(map[string]interface{})
		if endpoint != nil {
			endpoints = append(endpoints, types.ObjectValueMust(
				NCEndpointAttributeType(),
				NCEndpointAttributeValue(endpoint)))
		}
	}
	return endpoints
}

func NCEndpointAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":     types.StringType,
		"href":   types.StringType,
		"config": types.ObjectType{AttrTypes: NCEndpointConfigAttributeType()},
		"state":  types.ObjectType{AttrTypes: NCEndpointStateAttributeType()},
		"acs":    types.ListType{ElemType: ACObjectType()},
	}
}

func NCEndpointAttributeValue(endpoint map[string]interface{}) map[string]attr.Value {
	id := types.StringNull()
	if endpoint["id"] != nil {
		id = types.StringValue(endpoint["id"].(string))
	}
	href := types.StringNull()
	if endpoint["href"] != nil {
		href = types.StringValue(endpoint["href"].(string))
	}
	config := types.ObjectNull(NCEndpointConfigAttributeType())
	if (endpoint["config"]) != nil {
		config = types.ObjectValueMust(NCEndpointConfigAttributeType(), NCEndpointConfigAttributeValue(endpoint["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(NCEndpointStateAttributeType())
	if (endpoint["state"]) != nil {
		state = types.ObjectValueMust(NCEndpointStateAttributeType(), NCEndpointStateAttributeValue(endpoint["state"].(map[string]interface{})))
	}
	acs := types.ListNull(ACObjectType())
	if (endpoint["acs"]) != nil {
		acs = types.ListValueMust(ACObjectType(), ACObjectsValue(endpoint["acs"].([]interface{})))
	} 

	return map[string]attr.Value{
		"id":     id,
		"href":   href,
		"config": config,
		"state":  state,
		"acs" :  acs,
	}
}

func NCEndpointStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"host_port":     common.EndpointHostPortObjectType(),
		"module_if":     types.ObjectType{AttrTypes: common.ModuleIfAttributeType()},
		"capacity":      types.Int64Type,
	}
}

func NCEndpointStateAttributeValue(state map[string]interface{}) map[string]attr.Value {
	capacity := types.Int64Null()
	if state["capacity"] != nil {
		capacity = types.Int64Value(int64(state["capacity"].(float64)))
	}
	hostPort := types.ObjectNull(common.EndpointHostPortAttributeType())
	if state["hostPort"] != nil {
		hostPort = types.ObjectValueMust(common.EndpointHostPortAttributeType(), common.EndpointHostPortAttributeValue(state["hostPort"].(map[string]interface{})))
	}
	moduleIf := types.ObjectNull(common.ModuleIfAttributeType())
	if state["moduleIf"] != nil {
		moduleIf = types.ObjectValueMust(common.ModuleIfAttributeType(), common.ModuleIfAttributeValue(state["moduleIf"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"capacity":      capacity,
		"host_port":     hostPort,
		"module_if": moduleIf,
	}
}

func NCEndpointConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"selector":      types.ObjectType{AttrTypes: common.IfSelectorAttributeType()},
		"capacity":      types.Int64Type,
	}
}

func NCEndpointConfigAttributeValue(config map[string]interface{}) map[string]attr.Value {
	capacity := types.Int64Null()
	if config["capacity"] != nil {
		capacity = types.Int64Value(int64(config["capacity"].(float64)))
	}
	selector := types.ObjectNull(common.IfSelectorAttributeType())
	if (config["selector"]) != nil {
		selector = types.ObjectValueMust(common.IfSelectorAttributeType(), common.IfSelectorAttributeValue(config["selector"].(map[string]interface{})))
	}
	
	return map[string]attr.Value{
		"capacity":      capacity,
		"selector":      selector,
	}
}
