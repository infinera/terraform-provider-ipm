package network

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"terraform-provider-ipm/internal/ipm_pf"

	//	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                = &LeafModuleResource{}
	_ resource.ResourceWithConfigure   = &LeafModuleResource{}
	_ resource.ResourceWithImportState = &LeafModuleResource{}
)

// NewLeafModuleResource is a helper function to simplify the provider implementation.
func NewLeafModuleResource() resource.Resource {
	return &LeafModuleResource{}
}

type LeafModuleResource struct {
	client *ipm_pf.Client
}


// Metadata returns the data source type name.
func (r *LeafModuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_leaf_module"
}

// Schema defines the schema for the data source.
func (r *LeafModuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type ModuleResourceData struct
	resp.Schema = schema.Schema{
		Description: "Leaf Module",
		Attributes:  LeafModuleSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *LeafModuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r LeafModuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ModuleResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "LeafModuleResource: Create - ", map[string]interface{}{"ModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LeafModuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ModuleResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LeafModuleResource: Create - ", map[string]interface{}{"ModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LeafModuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ModuleResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "LeafModuleResource: Update", map[string]interface{}{"ModuleResourceData": data})

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

func (r LeafModuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ModuleResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LeafModuleResource: Delete", map[string]interface{}{"ModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *LeafModuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *LeafModuleResource) create(plan *ModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.NetworkId.IsNull() {
		diags.AddError(
			"LeafModuleResource: Error Create Module",
			"Create: Could not create Module for network, NetworkId is not specified.",
		)
		return
	}

	if plan.Config.Selector.ModuleSelectorByModuleId == nil && plan.Config.Selector.ModuleSelectorByModuleName == nil &&
		plan.Config.Selector.ModuleSelectorByModuleMAC == nil && plan.Config.Selector.ModuleSelectorByModuleSerialNumber == nil &&
		plan.Config.Selector.HostPortSelectorByName == nil && plan.Config.Selector.HostPortSelectorByPortId == nil &&
		plan.Config.Selector.HostPortSelectorBySysName == nil && plan.Config.Selector.HostPortSelectorByPortSourceMAC == nil {
		diags.AddError(
			"LeafModuleResource: Error Create LeafModuleResource",
			"Create: Could not create LeafModuleResource, No hub module selector specified",
		)
		return
	}

	var createRequest = make(map[string]interface{})
	// get Module setting
	var selector = make(map[string]interface{})
	var module = make(map[string]interface{})
	aSelector := make(map[string]interface{})
	if plan.Config.Selector.ModuleSelectorByModuleId != nil {
		aSelector["moduleId"] = plan.Config.Selector.ModuleSelectorByModuleId.ModuleId.ValueString()
		selector["moduleSelectorByModuleId"] = aSelector
	} else if plan.Config.Selector.ModuleSelectorByModuleName != nil {
		aSelector["moduleName"] = plan.Config.Selector.ModuleSelectorByModuleName.ModuleName.ValueString()
		selector["moduleSelectorByModuleName"] = aSelector
	} else if plan.Config.Selector.ModuleSelectorByModuleMAC != nil {
		aSelector["moduleMAC"] = plan.Config.Selector.ModuleSelectorByModuleMAC.ModuleMAC.ValueString()
		selector["moduleSelectorByModuleMAC"] = aSelector
	} else if plan.Config.Selector.ModuleSelectorByModuleSerialNumber != nil {
		aSelector["moduleSerialNumber"] = plan.Config.Selector.ModuleSelectorByModuleSerialNumber.ModuleSerialNumber.ValueString()
		selector["moduleSelectorByModuleSerialNumber"] = aSelector
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
	}
	createRequest["selector"] = selector

	if !plan.Config.Module.TrafficMode.IsNull() {
		module["trafficMode"] = plan.Config.Module.TrafficMode.ValueString()
	}
	if !plan.Config.Module.FecIterations.IsNull() {
		module["fecIterations"] = plan.Config.Module.FecIterations.ValueString()
	}
	if !plan.Config.Module.FiberConnectionMode.IsNull() {
		module["fiberConnectionMode"] = plan.Config.Module.FiberConnectionMode.ValueString()
	}
	if !plan.Config.ManagedBy.IsNull() {
		module["managedBy"] = plan.Config.ManagedBy.ValueString()
	}
	if !plan.Config.Module.RequestedNominalPsdOffset.IsNull() {
		module["requestedNominalPsdOffset"] = plan.Config.Module.RequestedNominalPsdOffset.ValueString()
	}
	if !plan.Config.Module.MaxDSCs.IsNull() {
		module["maxDSCs"] = plan.Config.Module.MaxDSCs.ValueInt64()
	}
	if !plan.Config.Module.MaxTxDSCs.IsNull() {
		module["maxTxDSCs"] = plan.Config.Module.MaxTxDSCs.ValueInt64()
	}
	if !plan.Config.Module.TxCLPtarget.IsNull() {
		module["txCLPtarget"] = plan.Config.Module.TxCLPtarget.ValueInt64()
	}
	if !plan.Config.Module.PlannedCapacity.IsNull() {
		module["plannedCapacity"] = plan.Config.Module.PlannedCapacity.ValueString()
	}
	createRequest["module"] = module
	tflog.Debug(ctx, "LeafModuleResource: create ## ", map[string]interface{}{"Create Reauest": createRequest})

	var request []map[string]interface{}
	request = append(request, createRequest)

	// send create request to server
	rb, err := json.Marshal(request)
	if err != nil {
		diags.AddError(
			"LeafModuleResource: create ##: Error Create LeafModule",
			"Create: Could not Marshal LeafModuleResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/xr-networks/"+plan.NetworkId.ValueString()+"/leafModules", rb)
	if err != nil {
		diags.AddError(
			"LeafModuleResource: create ##: Error create LeafModuleResource",
			"Create:Could not create LeafModuleResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "LeafModuleResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	// if create fails, can't get the network
	//r.read(plan, ctx, diags)
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"LeafModuleResource: create ##: Error Unmarshal response",
			"Create:Could not Create LeafModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	result := data[0].(map[string]interface{})
	
	href := result["href"].(string)
	splits := strings.Split(href, "/")
	id := splits[len(splits)-1]
	plan.Href = types.StringValue(href)
	plan.Id = types.StringValue(id)

	r.read(plan, ctx, diags, 10)
	if diags.HasError() {
		tflog.Debug(ctx, "LeafModuleResource: create failed. Can't find the created LeafModuleResource")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "LeafModuleResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *LeafModuleResource) update(plan *ModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.NetworkId.IsNull() || plan.Id.IsNull() {
		diags.AddError(
			"LeafModuleResource: Error Update Module",
			"Update: Could not Update. Both NetworkId and Module ID are required.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})
	module := make(map[string]interface{})

	if !plan.Config.Module.PlannedCapacity.IsNull() {
		module["plannedCapacity"] = plan.Config.Module.PlannedCapacity.ValueString()
	}
	if !plan.Config.Module.FiberConnectionMode.IsNull() {
		module["fiberConnectionMode"] = plan.Config.Module.FiberConnectionMode.ValueString()
	}
	if !plan.Config.Module.TrafficMode.IsNull() {
		module["trafficMode"] = plan.Config.Module.TrafficMode.ValueString()
	}
	if !plan.Config.Module.FecIterations.IsNull() {
		module["fecIterations"] = plan.Config.Module.FecIterations.ValueString()
	}
	if !plan.Config.Module.RequestedNominalPsdOffset.IsNull() {
		module["requestedNominalPsdOffset"] = plan.Config.Module.RequestedNominalPsdOffset.ValueString()
	}
	if !plan.Config.Module.TxCLPtarget.IsNull() {
		module["txCLPtarget"] = plan.Config.Module.TxCLPtarget.ValueInt64()
	}
	if !plan.Config.ManagedBy.IsNull() {
		module["managedBy"] = plan.Config.ManagedBy.ValueString()
	}
	updateRequest["module"] = module

	// send Update request to server
	rb, err := json.Marshal(updateRequest)
	if err != nil {
		diags.AddError(
			"LeafModuleResource: Update ##: Error Update LeafModule",
			"Update: Could not Marshal LeafModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	body, err := r.client.ExecuteIPMHttpCommand("PUT", "/xr-networks/"+plan.NetworkId.ValueString()+"/leafModules/"+plan.Id.ValueString(), rb)
	if err != nil {
		diags.AddError(
			"LeafModuleResource: Update ##: Error Update LeafModule",
			"Update: Could not Update LeafModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "LeafModuleResource: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"LeafModuleResource: Update ##: Error Unmarshal LeafModuleResource",
			"Update:Could not Update LeafModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "LeafModuleResource: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": data})

	r.read(plan, ctx, diags,2)

	tflog.Debug(ctx, "LeafModuleResource: update ## ", map[string]interface{}{"plan": plan})
}

func (r *LeafModuleResource) read(state *ModuleResourceData, ctx context.Context, diags *diag.Diagnostics, retryCount ...int) {

	numRetry := 1
  if len(retryCount) > 0 {
    numRetry = retryCount[0]
  }

	if state.NetworkId.IsNull() || state.Id.IsNull() {
		diags.AddError(
			"LeafModuleResource: Error read Leaf Module",
			"Read: Could not read leaf module , Network ID and Module ID are required.",
		)
		return
	}
	var err error
	body := []byte{}
	for i := 1; i <= numRetry; i++ {
		body, err = r.client.ExecuteIPMHttpCommand("GET", "/xr-networks/"+state.NetworkId.ValueString()+"/leafModules/"+state.Id.ValueString()+"?content=expanded", nil)
		if err == nil {
			break
		} else if i < numRetry {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		diags.AddError(
			"LeafModuleResource: read ##: Error Update LeafModuleResource",
			"Read:Could not read LeafModuleResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "LeafModuleResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"LeafModuleResource: read ##: Error Unmarshal response",
			"Read: Could not read LeafModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	// populate state
	state.Populate(data, ctx, diags)

	tflog.Debug(ctx, "LeafModuleResource: read ## ", map[string]interface{}{"plan": state})
}

func (r *LeafModuleResource) delete(plan *ModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.NetworkId.IsNull() || plan.Id.IsNull() {
		diags.AddError(
			"LeafModuleResource: Error Delete Leaf Module",
			"Delete: Could not delete. Network ID or Module Id is not specified",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", "/xr-networks/"+plan.NetworkId.ValueString()+"/leafModules/"+plan.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"LeafModuleResource: delete ##: Error Delete LeafModule",
			"Delete:Could not delete LeafModule, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "LeafModuleResource: delete ## ", map[string]interface{}{"plan": plan})
}
