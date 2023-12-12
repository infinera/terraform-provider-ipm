package network

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"
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
	_ resource.Resource                = &NetworkResource{}
	_ resource.ResourceWithConfigure   = &NetworkResource{}
	_ resource.ResourceWithImportState = &NetworkResource{}
)

// NewNetworkResource is a helper function to simplify the provider implementation.
func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

type NetworkResource struct {
	client *ipm_pf.Client
}

type NetworkConfig struct {
	Name                   types.String `tfsdk:"name"`
	ConstellationFrequency types.Int64  `tfsdk:"constellation_frequency"`
	Modulation             types.String `tfsdk:"modulation"`
	TcMode                 types.Bool   `tfsdk:"tc_mode"`
	Topology               types.String `tfsdk:"topology"`
	ManagedBy              types.String `tfsdk:"managed_by"`
}

type NetworkResourceData struct {
	Id               types.String                  `tfsdk:"id"`
	Href             types.String                  `tfsdk:"href"`
	Config           *NetworkConfig                 `tfsdk:"config"`
	State            types.Object                  `tfsdk:"state"`
	HubModule        *ModuleResourceData            `tfsdk:"hub_module"`
	LeafModules      types.List                    `tfsdk:"leaf_modules"`
	ReachableModules types.List                    `tfsdk:"reachable_modules"`
}

// Metadata returns the data source type name.
func (r *NetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_constellation_network"
}

// Schema defines the schema for the data source.
func (r *NetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type NetworkResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an Network",
		Attributes:  NetworkSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *NetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "NetworkResource: Create - ", map[string]interface{}{"NetworkResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)
	resp.State.Set(ctx, &data)

}

func (r NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "NetworkResource: Read - ", map[string]interface{}{"NetworkResourceData": data})

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "NetworkResource: Update", map[string]interface{}{"NetworkResourceData": data})

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

func (r NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NetworkResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "NetworkResource: Delete", map[string]interface{}{"NetworkResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NetworkResource) create(plan *NetworkResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "NetworkResource: create ## ", map[string]interface{}{"plan": plan})

	if plan.Config.ConstellationFrequency.IsNull() {
		diags.AddError(
			"Error Create NetworkResource",
			"Create: Could not create NetworkResource, ConstellationFrequency is not specified",
		)
		return
	}

	var createRequest = make(map[string]interface{})

	// get Network config settings
	var configRequest = make(map[string]interface{})
	if !plan.Config.Name.IsNull() {
		configRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.ConstellationFrequency.IsNull() {
		configRequest["constellationFrequency"] = plan.Config.ConstellationFrequency.ValueInt64()
	}
	if !plan.Config.Modulation.IsNull() {
		configRequest["modulation"] = plan.Config.Modulation.ValueString()
	}
	if !plan.Config.Topology.IsNull() {
		configRequest["topology"] = plan.Config.Topology.ValueString()
	}
	if !plan.Config.TcMode.IsNull() {
		configRequest["tcMode"] = plan.Config.TcMode.ValueBool()
	}
	createRequest["config"] = configRequest
	tflog.Debug(ctx, "NetworkResource: create ## ", map[string]interface{}{"configRequest": configRequest})

	// get hubModule setting
	var hubModuleRequest = make(map[string]interface{})
	var selector = make(map[string]interface{})
	aSelector := make(map[string]interface{})
	if plan.HubModule.Config.Selector.ModuleSelectorByModuleId != nil {
		aSelector["moduleId"] = plan.HubModule.Config.Selector.ModuleSelectorByModuleId.ModuleId.ValueString()
		selector["moduleSelectorByModuleId"] = aSelector
	} else if plan.HubModule.Config.Selector.ModuleSelectorByModuleName != nil {
		aSelector["moduleName"] = plan.HubModule.Config.Selector.ModuleSelectorByModuleName.ModuleName.ValueString()
		selector["moduleSelectorByModuleName"] = aSelector
	} else if plan.HubModule.Config.Selector.ModuleSelectorByModuleMAC != nil {
		aSelector["moduleMAC"] = plan.HubModule.Config.Selector.ModuleSelectorByModuleMAC.ModuleMAC.ValueString()
		selector["moduleSelectorByModuleMAC"] = aSelector
	} else if plan.HubModule.Config.Selector.ModuleSelectorByModuleSerialNumber != nil {
		aSelector["moduleSerialNumber"] = plan.HubModule.Config.Selector.ModuleSelectorByModuleSerialNumber.ModuleSerialNumber.ValueString()
		selector["moduleSelectorByModuleSerialNumber"] = aSelector
	} else if plan.HubModule.Config.Selector.HostPortSelectorByName != nil {
		aSelector["hostName"] = plan.HubModule.Config.Selector.HostPortSelectorByName.HostName.ValueString()
		aSelector["hostPortName"] = plan.HubModule.Config.Selector.HostPortSelectorByName.HostPortName.ValueString()
		selector["hostPortSelectorByName"] = aSelector
	} else if plan.HubModule.Config.Selector.HostPortSelectorByPortId != nil {
		aSelector["chassisId"] = plan.HubModule.Config.Selector.HostPortSelectorByPortId.ChassisId.ValueString()
		aSelector["chassisIdSubtype"] = plan.HubModule.Config.Selector.HostPortSelectorByPortId.ChassisIdSubtype.ValueString()
		aSelector["portId"] = plan.HubModule.Config.Selector.HostPortSelectorByPortId.PortId.ValueString()
		aSelector["portIdSubtype"] = plan.HubModule.Config.Selector.HostPortSelectorByPortId.PortIdSubtype.ValueString()
		selector["hostPortSelectorByPortId"] = aSelector
	} else if plan.HubModule.Config.Selector.HostPortSelectorBySysName != nil {
		aSelector["sysName"] = plan.HubModule.Config.Selector.HostPortSelectorBySysName.SysName.ValueString()
		aSelector["portId"] = plan.HubModule.Config.Selector.HostPortSelectorByPortId.PortId.ValueString()
		aSelector["portIdSubtype"] = plan.HubModule.Config.Selector.HostPortSelectorByPortId.PortIdSubtype.ValueString()
		selector["hostPortSelectorBySysName"] = aSelector
	} else if plan.HubModule.Config.Selector.HostPortSelectorByPortSourceMAC != nil {
		aSelector["portSourceMAC"] = plan.HubModule.Config.Selector.HostPortSelectorByPortSourceMAC.PortSourceMAC.ValueString()
		selector["hostPortSelectorByPortSourceMAC"] = aSelector
	} else {
		diags.AddError(
			"Error Create NetworkResource",
			"Create: Could not create NetworkResource, No hub module selector specified",
		)
		return
	}
	hubModuleRequest["selector"] = selector
	var module = make(map[string]interface{})
	if !plan.HubModule.Config.Module.TrafficMode.IsNull() {
		module["trafficMode"] = plan.HubModule.Config.Module.TrafficMode.ValueString()
	}
	if !plan.HubModule.Config.Module.FecIterations.IsNull() {
		module["fecIterations"] = plan.HubModule.Config.Module.FecIterations.ValueString()
	}
	if !plan.HubModule.Config.Module.FiberConnectionMode.IsNull() {
		module["fiberConnectionMode"] = plan.HubModule.Config.Module.FiberConnectionMode.ValueString()
	}
	if !plan.HubModule.Config.Module.RequestedNominalPsdOffset.IsNull() {
		module["requestedNominalPsdOffset"] = plan.HubModule.Config.Module.RequestedNominalPsdOffset.ValueString()
	}
	if !plan.HubModule.Config.Module.TxCLPtarget.IsNull() {
		module["txCLPtarget"] = plan.HubModule.Config.Module.TxCLPtarget.ValueInt64()
	}
	if !plan.HubModule.Config.Module.PlannedCapacity.IsNull() {
		module["plannedCapacity"] = plan.HubModule.Config.Module.PlannedCapacity.ValueString()
	}
	hubModuleRequest["module"] = module
	createRequest["hubModule"] = hubModuleRequest

	tflog.Debug(ctx, "NetworkResource: create ## ", map[string]interface{}{"Create Request": createRequest})

	// send create request to server
	var request []map[string]interface{}
	request = append(request, createRequest)
	rb, err := json.Marshal(request)
	if err != nil {
		diags.AddError(
			"NetworkResource: create ##: Error Create AC",
			"Create: Could not Marshal NetworkResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/xr-networks", rb)
	if err != nil {
		if !strings.Contains(err.Error(), "status: 202") {
			diags.AddError(
				"NetworkResource: create ##: Error create NetworkResource",
				"Create:Could not create NetworkResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "NetworkResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NetworkResource: Create ##: Error Unmarshal response",
			"Update:Could not Create NetworkResource, unexpected error: "+err.Error(),
		)
		return
	}

	result := data[0].(map[string]interface{})

	href := result["href"].(string)
	splits := strings.Split(href, "/")
	id := splits[len(splits)-1]
	plan.Href = types.StringValue(href)
	plan.Id = types.StringValue(id)

	r.read(plan, ctx, diags, 3)
	if diags.HasError() {
		tflog.Debug(ctx, "NetworkResource: create failed. Can't find the created network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "NetworkResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *NetworkResource) update(plan *NetworkResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.Id.IsNull() || plan.Id.IsUnknown() {
		diags.AddError(
			"Error Update NetworkResource",
			"Update: Could not Update. NetworkResource ID is not specified",
		)
		return
	}
// update network
	var updateRequest = make(map[string]interface{})
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.ConstellationFrequency.IsNull() {
		updateRequest["constellationFrequency"] = plan.Config.ConstellationFrequency.ValueInt64()
	}
	if !plan.Config.Modulation.IsNull() {
		updateRequest["modulation"] = plan.Config.Modulation.ValueString()
	}
	if !plan.Config.TcMode.IsNull() {
		updateRequest["tcMode"] = plan.Config.TcMode.ValueBool()
	}
	if !plan.Config.ManagedBy.IsNull() {
		updateRequest["managedBy"] = plan.Config.ManagedBy.ValueString()
	}

	tflog.Debug(ctx, "NetworkResource: update ## ", map[string]interface{}{"id": plan.Id.ValueString(),"Update Request": updateRequest})

	if len(updateRequest) >0 {
		// send Update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"NetworkResource: Update ##: Error Update AC",
				"Update: Could not Marshal NetworkResource, unexpected error: "+err.Error(),
			)
			return
		}
		body, err := r.client.ExecuteIPMHttpCommand("PUT", "/xr-networks/"+plan.Id.ValueString(), rb)
		if err != nil {
			diags.AddError(
				"NetworkResource: Update ##: Error Update NetworkResource",
				"Update:Could not Update NetworkResource, unexpected error: "+err.Error(),
			)
			return
		}

		tflog.Debug(ctx, "NetworkResource: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	}

	/*// update hub module 
	updateRequest = make(map[string]interface{})
	module := make(map[string]interface{})

	if !plan.HubModule.Config.Module.TrafficMode.IsNull() {
		module["trafficMode"] = plan.HubModule.Config.Module.TrafficMode.ValueString()
	}
	if !plan.HubModule.Config.Module.FecIterations.IsNull() {
		module["fecIterations"] = plan.HubModule.Config.Module.FecIterations.ValueString()
	}
	if !plan.HubModule.Config.Module.FiberConnectionMode.IsNull() {
		module["fiberConnectionMode"] = plan.HubModule.Config.Module.FiberConnectionMode.ValueString()
	}
	if !plan.HubModule.Config.ManagedBy.IsNull() {
		module["managedBy"] = plan.HubModule.Config.ManagedBy.ValueString()
	}
	if !plan.HubModule.Config.Module.PlannedCapacity.IsNull() {
		module["plannedCapacity"] = plan.HubModule.Config.Module.PlannedCapacity.ValueString()
	}
	if !plan.HubModule.Config.Module.RequestedNominalPsdOffset.IsNull() {
		module["requestedNominalPsdOffset"] = plan.HubModule.Config.Module.RequestedNominalPsdOffset.ValueString()
	}
	if !plan.HubModule.Config.Module.TxCLPtarget.IsNull() {
		module["txCLPtarget"] = plan.HubModule.Config.Module.TxCLPtarget.ValueInt64()
	}
	updateRequest["module"] = module

	// send Update request to server
	rb, err := json.Marshal(updateRequest)
	if err != nil {
		diags.AddError(
			"NetworkResource: Update ##: Error Update Network Hub Module",
			"Update:Could not Update Network Hub Module, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "NetworkResource Hub module: update - rb", map[string]interface{}{"rb": rb})

	body, err := r.client.ExecuteIPMHttpCommand("PUT", "/xr-networks/"+plan.Id.ValueString()+"/hubModule", rb)

	if err != nil {
		diags.AddError(
			"NetworkResource: Update ##: Error Update Network Hub Module",
			"Update:Could not Update Network Hub Module, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HubModuleResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NetworkResource: Update ##: Error Unmarshal Network Hub Module",
			"Update:Could not Update NetworkResource, unexpected error: "+err.Error(),
		)
		return
	}
*/
	r.read(plan, ctx, diags, 2)

	tflog.Debug(ctx, "NetworkResource: update ## ", map[string]interface{}{"plan": plan})
}

func (r *NetworkResource) read(state *NetworkResourceData, ctx context.Context, diags *diag.Diagnostics, retryCount ...int) {

	numRetry := 1
  if len(retryCount) > 0 {
    numRetry = retryCount[0]
  }

	tflog.Debug(ctx, "NetworkResource: read ##", map[string]interface{}{"state": state})
	queryString := "?content=expanded"
	if state.Id.IsNull() {
		queryString = queryString + "&q={\"hubModule.state.module.moduleName\":\"" + state.HubModule.Config.Selector.ModuleSelectorByModuleName.ModuleName.ValueString() + "\"}"
	} else {
		queryString = "/" + state.Id.ValueString() + queryString
	}
	var err error
	body := []byte{}
	for i := 1; i <= numRetry; i++ {
		body, err = r.client.ExecuteIPMHttpCommand("GET", "/xr-networks"+queryString, nil)
		if err == nil {
			break
		} else if i < numRetry {
			time.Sleep(2 * time.Second)
		}
	} 

	if err != nil {
		diags.AddError(
			"NetworkResource: read ##: Error Get NetworkResource",
			"Read:Could not get Network, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "NetworkResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NetworkResource: read ##: Error Get Network",
			"Read:Could not get Network, unexpected error: "+err.Error(),
		)
		return
	}
	// populate network state
	state.Populate(data, ctx, diags)
	tflog.Debug(ctx, "NetworkResource: read SUCCESS ")
}

func (r *NetworkResource) delete(plan *NetworkResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.Id.IsNull() {
		diags.AddError(
			"NetworkResource: Error Delete NetworkResource",
			"Delete: Could not delete. NetworkResource Id is not specified",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", "/xr-networks/"+plan.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"NetworkResource: delete ##: Error Update NetworkResource",
			"Update:Could not delete NetworkResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "NetworkResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (nwData *NetworkResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}
	tflog.Debug(ctx, "NetworkResource: populate ## ", map[string]interface{}{"NW ID ": data["id"]})
	//populate state
	if computeFlag {
		nwData.Id = types.StringValue(data["id"].(string))
	}
	nwData.Href = types.StringValue(data["href"].(string))
	if (data["state"]) != nil {
		nwData.State = types.ObjectValueMust(NetworkStateAttributeType(), NetworkStateAttributeValue(data["state"].(map[string]interface{})))
	}
	//populate config
	if data["config"] != nil {
		if  nwData.Config == nil {
			nwData.Config = &NetworkConfig{}
		}
		networkConfig := data["config"].(map[string]interface{})
		for k, v := range networkConfig {
			switch k {
			case "name":
				if !nwData.Config.Name.IsNull() || computeFlag {
					nwData.Config.Name = types.StringValue(v.(string))
				}
			case "constellationFrequency":
				if !nwData.Config.ConstellationFrequency.IsNull() || computeFlag {
					nwData.Config.ConstellationFrequency = types.Int64Value(int64(v.(float64)))
				}
			case "modulation":
				if !nwData.Config.Modulation.IsNull() || computeFlag {
					nwData.Config.Modulation = types.StringValue(v.(string))
				}
			case "tcMode":
				if !nwData.Config.TcMode.IsNull() || computeFlag {
					nwData.Config.TcMode = types.BoolValue(v.(bool))
				}
			case "topology":
				if !nwData.Config.Topology.IsNull() || computeFlag {
					nwData.Config.Topology = types.StringValue(v.(string))
				}
			case "managedBy":
				if !nwData.Config.ManagedBy.IsNull() || computeFlag {
					nwData.Config.ManagedBy = types.StringValue(v.(string))
				}
			}
		}
	}
	// populate network hubModule
	hubModule := data["hubModule"].(map[string]interface{})
	if nwData.HubModule == nil {
		nwData.HubModule = &ModuleResourceData{}
	}
	nwData.HubModule.Populate(hubModule, ctx, diags, computeFlag)

	tflog.Debug(ctx, "NetworkResource: populate network HubModules SUCCESS ")

	// populate network leafModules
	nwData.LeafModules = types.ListNull(NWModuleObjectType())
	if data["leafModules"] != nil && len(data["leafModules"].([]interface{})) > 0 {
		nwData.LeafModules = types.ListValueMust(NWModuleObjectType(), NWModulesValue(data["leafModules"].([]interface{})))
	}

	tflog.Debug(ctx, "NetworkResource: populate network leafModules SUCCESS ")
	// populate network reachable Modules
	nwData.ReachableModules = types.ListNull(NWReachableModuleObjectType())
	if data["reachableModules"] != nil && len(data["reachableModules"].([]interface{})) > 0{
		nwData.ReachableModules = types.ListValueMust(NWReachableModuleObjectType(), NWReachableModulesValue(data["reachableModules"].([]interface{})))
	}
	
	tflog.Debug(ctx, "NetworkResource: populate network reachable module SUCCESS ")

	tflog.Debug(ctx, "NetworkResource: Populated SUCCESS ", map[string]interface{}{"NW ID ": data["id"]})
}

func NetworkSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the Network.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		//Config           NetworkConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Description: "config",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "Network Name",
					Optional:    true,
				},
				"constellation_frequency": schema.Int64Attribute{
					Description: "constellation_frequency",
					Optional:    true,
				},
				"modulation": schema.StringAttribute{
					Description: "modulation",
					Optional:    true,
				},
				"tc_mode": schema.BoolAttribute{
					Description: "tc_mode",
					Optional:    true,
				},
				"topology": schema.StringAttribute{
					Description: "topology",
					Optional:    true,
				},
				"managed_by": schema.StringAttribute{
					Description: "managed_by",
					Optional:    true,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: NetworkStateAttributeType(),
		},
		//HubModule        Module `tfsdk:"hub_module"`
		"hub_module": schema.SingleNestedAttribute{
			Description: "hub_module",
			Optional:    true,
			Attributes:  HubModuleSchemaAttributes(),
		},
		//LeafModules      types.List `tfsdk:"leaf_modules"`
		"leaf_modules":schema.ListAttribute{
			Computed: true,
			ElementType: NWModuleObjectType(),
		},
		"reachable_modules":schema.ListAttribute{
			Computed: true,
			ElementType: NWReachableModuleObjectType(),
		},
	}
}

func NetworkControlLinkObjectType() (types.ObjectType) {
	return types.ObjectType{	
					AttrTypes: NetworkControlLinkAttributeType(),
				}	
}

func NetworkControlLinkAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
						"source_module_id": types.StringType,
						"destination_module_id": types.StringType,
						"con_state": types.StringType,
						"last_con_state_change": types.StringType,
				}	
}

func NetworkControlLinksValue(data []interface{}) ([]attr.Value) {
	controlLinks := []attr.Value{}
	for _, v := range data {
		controlLink := v.(map[string]interface{})
		sourceModuleId := types.StringNull()
		if controlLink["sourceModuleId"] != nil {
			sourceModuleId = types.StringValue(controlLink["sourceModuleId"].(string))
		}
		destinationModuleId := types.StringNull()
		if controlLink["destinationModuleId"] != nil {
			destinationModuleId = types.StringValue(controlLink["destinationModuleId"].(string))
		}
		conState := types.StringNull()
		if controlLink["conState"] != nil {
			conState = types.StringValue(controlLink["conState"].(string))
		}
		lastConStateChange := types.StringNull()
		if controlLink["lastConStateChange"] != nil {
			lastConStateChange = types.StringValue(controlLink["lastConStateChange"].(string))
		}
		controlLinks = append(controlLinks, types.ObjectValueMust(
													NetworkControlLinkAttributeType(),
													map[string]attr.Value{
														"source_module_id": sourceModuleId,
														"destination_module_id": destinationModuleId,
														"con_state": conState,
														"last_con_state_change": lastConStateChange,
													}))
	}
	return controlLinks
}

func NetworkAvailableServiceObjectType() (types.ObjectType) {
	return types.ObjectType{	
					AttrTypes: NetworkAvailableServiceAttributeType(),
				}	
}

func NetworkAvailableServiceAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
						"type": types.StringType,
						"maximum": types.Int64Type,
						"available": types.Int64Type,
						"used": types.Int64Type,
				}	
}

func NetworkAvailableServicesValue(data []interface{}) ([]attr.Value) {
	availableServices := []attr.Value{}
	for _, v := range data {
		availableService := v.(map[string]interface{})
		atype := types.StringNull()
		if availableService["type"] != nil {
			atype = types.StringValue(availableService["type"].(string))
		}
		maximum := types.Int64Null()
		if availableService["maximum"] != nil {
			maximum = types.Int64Value(int64(availableService["maximum"].(float64)))
		}
		available := types.Int64Null()
		if availableService["available"] != nil {
			available = types.Int64Value(int64(availableService["available"].(float64)))
		}
		used := types.Int64Null()
		if availableService["used"] != nil {
			used = types.Int64Value(int64(availableService["used"].(float64)))
		}
		availableServices = append(availableServices, types.ObjectValueMust(
													NetworkAvailableServiceAttributeType(),
													map[string]attr.Value{
														"type": atype,
														"maximum": maximum,
														"available": available,
														"used": used,
													}))
	}
	return availableServices
}

func NetworkObjectType() (types.ObjectType) {
	return types.ObjectType{	
					AttrTypes: NetworkAttributeType(),
				}	
}

func NetworkAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"id":     types.StringType,
		"href":   types.StringType,
		"config": types.ObjectType{AttrTypes: NetworkConfigAttributeType()},
		"state":  types.ObjectType{AttrTypes: NetworkStateAttributeType()},
		"hub_module":  types.ObjectType{AttrTypes: NWModuleAttributeType()},
		"leaf_modules": types.ListType{ElemType:  NWModuleObjectType()},
		"reachable_modules": types.ListType{ElemType:  NWReachableModuleObjectType()},
	}
}

func NetworkAttributeValue(network map[string]interface{}) map[string]attr.Value {
	id := types.StringNull()
	if network["id"] != nil {
		id = types.StringValue(network["id"].(string))
	}
	href := types.StringNull()
	if network["href"] != nil {
		href = types.StringValue(network["href"].(string))
	}
	config := types.ObjectNull(NetworkConfigAttributeType())
	if (network["config"]) != nil {
		config = types.ObjectValueMust(NetworkConfigAttributeType(), NetworkConfigAttributeValue(network["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(NetworkStateAttributeType())
	if (network["state"]) != nil {
		state = types.ObjectValueMust(NetworkStateAttributeType(), NetworkStateAttributeValue(network["state"].(map[string]interface{})))
	}
	hubModule := types.ObjectNull(NWModuleAttributeType())
	if (network["hubModule"]) != nil {
		hubModule = types.ObjectValueMust(NWModuleAttributeType(), NWModuleAttributeValue(network["hubModule"].(map[string]interface{})))
	}
	leafModules := types.ListNull(NWModuleObjectType())
	if network["leafModules"] != nil {
		leafModules = types.ListValueMust(NWModuleObjectType(), NWModulesValue(network["leafModules"].([]interface{})))
	}
	reachableModules := types.ListNull(NWReachableModuleObjectType())
	if network["reachableModules"] != nil {
		reachableModules = types.ListValueMust(NWReachableModuleObjectType(), NWReachableModulesValue(network["reachableModules"].([]interface{})))
	}

	return map[string]attr.Value{
		"id":     id,
		"href":   href,
		"config": config,
		"state":  state,
		"hub_module": hubModule,
		"leaf_modules": leafModules,
		"reachable_modules": reachableModules,
	}
}

func NetworksValue(data []interface{}) []attr.Value {
	networks := []attr.Value{}
	for _, v := range data {
		network := v.(map[string]interface{})
		if network != nil {
			networks = append(networks, types.ObjectValueMust(NetworkAttributeType(),NetworkAttributeValue(network)))
		}
	}
	return networks
}

func NetworkStateObjectType() (types.ObjectType) {
	return types.ObjectType{	
					AttrTypes: NetworkStateAttributeType(),
				}	
}

func NetworkStateAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"name": types.StringType,
		"constellation_frequency": types.Int64Type,
		"actual_constellation_frequency": types.Int64Type,
		"operating_frequency" : types.Int64Type,
		"modulation": types.StringType,
		"operating_modulation": types.StringType,
		"tc_mode": types.BoolType,
		"topology":  types.StringType,
		"managed_by": types.StringType,
		"lifecycle_state": types.StringType,
		"lifecycle_state_cause": types.ObjectType{AttrTypes: common.LifecycleStateCauseAttributeType() },
		"control_links": types.ListType{ElemType:  NetworkControlLinkObjectType()},
		"available_services": types.ListType{ElemType: NetworkAvailableServiceObjectType()},	
	}
}

func NetworkStateAttributeValue(state map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	if state["name"] != nil {
		name = types.StringValue(state["name"].(string))
	}
	constellationFrequency := types.Int64Null()
	if state["constellationFrequency"] != nil {
		constellationFrequency = types.Int64Value(int64(state["constellationFrequency"].(float64)))
	}
	actualConstellationFrequency := types.Int64Null()
	if state["actualConstellationFrequency"] != nil {
		actualConstellationFrequency = types.Int64Value(int64(state["actualConstellationFrequency"].(float64)))
	}
	operatingFrequency := types.Int64Null()
	if state["operatingFrequency"] != nil {
		operatingFrequency = types.Int64Value(int64(state["operatingFrequency"].(float64)))
	}
	tcMode := types.BoolNull()
	if state["tcMode"] != nil {
		tcMode = types.BoolValue(state["tcMode"].(bool))
	}
	modulation := types.StringNull()
	if state["modulation"] != nil {
		modulation = types.StringValue(state["modulation"].(string))
	}
	operatingModulation := types.StringNull()
	if state["operatingModulation"] != nil {
		operatingModulation = types.StringValue(state["operatingModulation"].(string))
	}
	topology := types.StringNull()
	if state["topology"] != nil {
		topology = types.StringValue(state["topology"].(string))
	}
	managedBy := types.StringNull()
	if state["managedBy"] != nil {
		managedBy = types.StringValue(state["managedBy"].(string))
	}
	lifecycleState := types.StringNull()
	if state["lifecycleState"] != nil {
		lifecycleState = types.StringValue(state["lifecycleState"].(string))
	}
	lifecycleStateCause := types.ObjectNull(common.LifecycleStateCauseAttributeType())
	if state["lifecycleStateCause"] != nil {
		lifecycleStateCause = types.ObjectValueMust(common.LifecycleStateCauseAttributeType(),
			common.LifecycleStateCauseAttributeValue(state["lifecycleStateCause"].(map[string]interface{})))
	}
	controlLinks := types.ListNull(NetworkControlLinkObjectType())
	if state["controlLinks"] != nil {
		controlLinks = types.ListValueMust(NetworkControlLinkObjectType(), NetworkControlLinksValue(state["controlLinks"].([]interface{})))
	}
	availableServices := types.ListNull(NetworkAvailableServiceObjectType())
	if state["availableServices"] != nil {
		availableServices = types.ListValueMust(NetworkAvailableServiceObjectType(), NetworkAvailableServicesValue(state["availableServices"].([]interface{})))
	}

	return map[string]attr.Value{
		"name":       name,
		"constellation_frequency": constellationFrequency,
		"actual_constellation_frequency": actualConstellationFrequency,
		"operating_frequency":     operatingFrequency,
		"operating_modulation":    operatingModulation,
		"modulation":              modulation,
		"tc_mode":                 tcMode,
		"topology":                topology,
		"managed_by":              managedBy,
		"lifecycle_state":         lifecycleState,
		"lifecycle_state_cause":   lifecycleStateCause,
		"control_links"  :         controlLinks,
		"available_services":      availableServices,
	}
}

func NetworkConfigObjectType() (types.ObjectType) {
	return types.ObjectType{	
					AttrTypes: NetworkConfigAttributeType(),
				}	
}

func NetworkConfigAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"name": types.StringType,
		"constellation_frequency": types.Int64Type,
		"modulation": types.StringType,
		"tc_mode": types.BoolType,
		"topology":  types.StringType,
		"managed_by": types.StringType,
	}
}

func NetworkConfigAttributeValue(config map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	if config["name"] != nil {
		name = types.StringValue(config["name"].(string))
	}
	constellationFrequency := types.Int64Null()
	if config["constellationFrequency"] != nil {
		constellationFrequency = types.Int64Value(int64(config["constellationFrequency"].(float64)))
	}
	tcMode := types.BoolNull()
	if config["tcMode"] != nil {
		tcMode = types.BoolValue(config["tcMode"].(bool))
	}
	modulation := types.StringNull()
	if config["modulation"] != nil {
		modulation = types.StringValue(config["modulation"].(string))
	}
	topology := types.StringNull()
	if config["topology"] != nil {
		topology = types.StringValue(config["topology"].(string))
	}
	managedBy := types.StringNull()
	if config["managedBy"] != nil {
		managedBy = types.StringValue(config["managedBy"].(string))
	}

	return map[string]attr.Value{
		"name":                    name,
		"constellation_frequency": constellationFrequency,
		"modulation":              modulation,
		"tc_mode":                 tcMode,
		"topology":                topology,
		"managed_by":              managedBy,
	}
}

