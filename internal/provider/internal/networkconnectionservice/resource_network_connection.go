package networkconnection

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"

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
	_ resource.Resource                = &NetworkConnectionResource{}
	_ resource.ResourceWithConfigure   = &NetworkConnectionResource{}
	_ resource.ResourceWithImportState = &NetworkConnectionResource{}
)

// NewNetworkConnectionResource is a helper function to simplify the provider implementation.
func NewNetworkConnectionResource() resource.Resource {
	return &NetworkConnectionResource{}
}

type NetworkConnectionResource struct {
	client *ipm_pf.Client
}

type NetworkConnectionConfig struct {
	Name                      types.String `tfsdk:"name"`
	ServiceMode               types.String `tfsdk:"service_mode"`                // XR-L1, XR-VTI-P2P, none
	MC                        types.String `tfsdk:"mc"`                          //matchAll, matchOuterVID,,
	OuterVID                  types.String `tfsdk:"outer_vid"`                   // "10,20,50..100"
	ImplicitTransportCapacity types.String `tfsdk:"implicit_transport_capacity"` // none, portMode, dedicatedDownlinkSymmetric, dedicatedDownlinkAsymmetric, sharedDownlink
	Labels                    types.Map    `tfsdk:"labels"`
}

type NetworkConnectionResourceData struct {
	Id        types.String             `tfsdk:"id"`
	Href      types.String             `tfsdk:"href"`
	Config    *NetworkConnectionConfig `tfsdk:"config"`
	State     types.Object             `tfsdk:"state"`
	Endpoints []NCEndpointResourceData `tfsdk:"end_points"`
	LCs       types.List               `tfsdk:"lcs"`
}

// Metadata returns the data source type name.
func (r *NetworkConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_connection"
}

// Schema defines the schema for the data source.
func (r *NetworkConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type NetworkConnectionResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an NC",
		Attributes:  NetworkConnectionSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *NetworkConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r NetworkConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkConnectionResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "NetworkConnectionResource: Create - ", map[string]interface{}{"NetworkConnectionResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)

	resp.State.Set(ctx, &data)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r NetworkConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkConnectionResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "NetworkConnectionResource: Read - ", map[string]interface{}{"NetworkConnectionResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r NetworkConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkConnectionResourceData
	var data2 NetworkConnectionResourceData

	diags := req.Plan.Get(ctx, &data)
	count := 0
	if data.Endpoints != nil {
		count = len(data.Endpoints)
	}
	tflog.Debug(ctx, "NetworkConnectionResource: Update 222: ", map[string]interface{}{"Plan id": data.Id.ValueString(), "plan href": data.Href.ValueString(), "Plan endpoints count": count})

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &data2)
	count = 0
	if data2.Endpoints != nil {
		count = len(data2.Endpoints)
	}
	tflog.Debug(ctx, "NetworkConnectionResource: Update 444: ", map[string]interface{}{"State id": data2.Id.ValueString(), "state href": data2.Href.ValueString(), "Config endpoints count": count})

	data.Id = data2.Id
	data.Href = data2.Href
	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r NetworkConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NetworkConnectionResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "NetworkConnectionResource: delete", map[string]interface{}{"NetworkConnectionResourceData": data})

	resp.Diagnostics.Append(diags...)	

	// DELAY TO MAKE SURE tc IS DELETED
	time.Sleep(1 * time.Second)
	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *NetworkConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NetworkConnectionResource) create(plan *NetworkConnectionResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.Config.Name.IsNull() || plan.Config.ServiceMode.IsNull() {
		diags.AddError(
			"Error Create NC",
			"Create: Could not create NC, NC Name or Service mode is not specified.",
		)
		return
	}
	if len(plan.Endpoints) < 2. {
		diags.AddError(
			"Error Create NC",
			"Create: Could not create NC, at least 2 endpoint must be specified.",
		)
		return
	}

	// Check to see if the related constellastion's lif cycle state is in' 'configured' state. Assume hub is the the first endpoint
	queryString := "/transport-capacities?content=expanded"
	if plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleId != nil {
		queryString = queryString + "&q={\"$and\":[{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleId.moduleId\":\"" + plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleId.ModuleId.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleId.moduleClientIfAid\":\"" + plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleId.ModuleClientIfAid.ValueString() + "\"}}},{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleId.moduleId\":\"" + plan.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleId.ModuleId.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleId.moduleClientIfAid\":\"" + plan.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleId.ModuleClientIfAid.ValueString() + "\"}}}]}"
	} else if plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleName != nil {
			queryString = queryString + "&q={\"$and\":[{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleName.moduleName\":\"" + plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleName.moduleClientIfAid\":\"" + plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString() + "\"}}},{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleName.moduleName\":\"" + plan.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleName.moduleClientIfAid\":\"" + plan.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString() + "\"}}}]}"
	} else if plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleMAC != nil {
			queryString = queryString + "&q={\"$and\":[{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleMAC.moduleMAC\":\"" + plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleMAC.ModuleMAC.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleMAC.moduleClientIfAid\":\"" + plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleMAC.ModuleClientIfAid.ValueString() + "\"}}},{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleMAC.moduleMAC\":\"" + plan.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleMAC.ModuleMAC.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleMAC.moduleClientIfAid\":\"" + plan.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleMAC.ModuleClientIfAid.ValueString() + "\"}}}]}"
	} else if plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleSerialNumber != nil {
			queryString = queryString + "&q={\"$and\":[{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleSerialNumber.moduleSerialNumber\":\"" + plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleSerialNumber.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleSerialNumber.moduleClientIfAid\":\"" + plan.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleClientIfAid.ValueString() + "\"}}},{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleSerialNumber.moduleSerialNumber\":\"" + plan.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleSerialNumber.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleSerialNumber.moduleClientIfAid\":\"" + plan.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleClientIfAid.ValueString() + "\"}}}]}"
	} 
	tflog.Debug(ctx, "NetworkConnectionResource: TC QueryString ## ", map[string]interface{}{"QueryString": queryString})

	_, err := common.CheckResourceState(ctx, r.client, queryString, 5) 
	if err != nil {
		diags.AddError(
			"Error Creating NC",
			"Create: Could not create NC: CheckResourceState: " + err.Error(),
		)
		return
	}

	var createRequest = make(map[string]interface{})
	createRequest["name"] = plan.Config.Name.ValueString()
	createRequest["serviceMode"] = plan.Config.ServiceMode.ValueString()

	if !plan.Config.MC.IsNull() {
		createRequest["mc"] = plan.Config.MC.ValueString()
	}
	if !plan.Config.OuterVID.IsNull() {
		createRequest["outerVID"] = plan.Config.OuterVID.ValueString()
	}
	if !plan.Config.ImplicitTransportCapacity.IsNull() {
		createRequest["implicitTransportCapacity"] = plan.Config.ImplicitTransportCapacity.ValueString()
	}

	if !plan.Config.Labels.IsNull() {
		labels := map[string]string{}
		diag := plan.Config.Labels.ElementsAs(ctx, &labels, true)
		if !diag.HasError() {
			createRequest["labels"] = labels
		}
	}
	endpoints := []interface{}{}
	for _, v := range plan.Endpoints {
		endpoint := make(map[string]interface{})
		endpoint["capacity"] = v.Config.Capacity.ValueInt64()
		selector := make(map[string]interface{})
		aSelector := make(map[string]interface{})
		if v.Config.Selector.ModuleIfSelectorByModuleId != nil {
			aSelector["moduleId"] = v.Config.Selector.ModuleIfSelectorByModuleId.ModuleId.ValueString()
			aSelector["moduleClientIfAid"] = v.Config.Selector.ModuleIfSelectorByModuleId.ModuleClientIfAid.ValueString()
			selector["moduleIfSelectorByModuleId"] = aSelector
		} else if v.Config.Selector.ModuleIfSelectorByModuleName != nil {
			aSelector["moduleName"] = v.Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString()
			aSelector["moduleClientIfAid"] = v.Config.Selector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString()
			tflog.Debug(ctx, "TransportCapacityResource: create ## moduleName", map[string]interface{}{"ModuleClientIfAid": aSelector["moduleClientIfAid"]})
			selector["moduleIfSelectorByModuleName"] = aSelector
		} else if v.Config.Selector.ModuleIfSelectorByModuleMAC != nil {
			aSelector["moduleMAC"] = v.Config.Selector.ModuleIfSelectorByModuleMAC.ModuleMAC.ValueString()
			aSelector["moduleClientIfAid"] = v.Config.Selector.ModuleIfSelectorByModuleMAC.ModuleClientIfAid.ValueString()
			selector["moduleIfSelectorByModuleMAC"] = aSelector
		} else if v.Config.Selector.ModuleIfSelectorByModuleSerialNumber != nil {
			aSelector["moduleSerialNumber"] = v.Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleSerialNumber.ValueString()
			aSelector["moduleClientIfAid"] = v.Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleClientIfAid.ValueString()
			selector["moduleIfSelectorByModuleSerialNumber"] = aSelector
		} else if v.Config.Selector.HostPortSelectorByName != nil {
			aSelector["hostName"] = v.Config.Selector.HostPortSelectorByName.HostName.ValueString()
			aSelector["hostPortName"] = v.Config.Selector.HostPortSelectorByName.HostPortName.ValueString()
			selector["hostPortSelectorByName"] = aSelector
		} else if v.Config.Selector.HostPortSelectorByPortId != nil {
			aSelector["chassisIdSubtype"] = v.Config.Selector.HostPortSelectorByPortId.ChassisIdSubtype.ValueString()
			aSelector["chassisId"] = v.Config.Selector.HostPortSelectorByPortId.ChassisId.ValueString()
			aSelector["portIdSubtype"] = v.Config.Selector.HostPortSelectorByPortId.PortIdSubtype.ValueString()
			aSelector["portId"] = v.Config.Selector.HostPortSelectorByPortId.PortId.ValueString()
			selector["hostPortSelectorByPortId"] = aSelector
		} else if v.Config.Selector.HostPortSelectorBySysName != nil {
			aSelector["sysName"] = v.Config.Selector.HostPortSelectorBySysName.SysName.ValueString()
			aSelector["portIdSubtype"] = v.Config.Selector.HostPortSelectorBySysName.PortIdSubtype.ValueString()
			aSelector["portId"] = v.Config.Selector.HostPortSelectorBySysName.PortId.ValueString()
			selector["hostPortSelectorByPortId"] = aSelector
		} else if v.Config.Selector.HostPortSelectorByPortSourceMAC != nil {
			aSelector["portSourceMAC"] = v.Config.Selector.HostPortSelectorByPortSourceMAC.PortSourceMAC.ValueString()
			selector["hostPortSelectorByName"] = aSelector
		} else {
			diags.AddError(
				"NetworkConnectionResource: Error Create TC. No selector specify for Endpoint",
				"Create: Could not create NetworkConnectionResource, No selector specify for Endpoint",
			)
			return
		}
		tflog.Debug(ctx, "NetworkConnectionResource: create ## ", map[string]interface{}{"Create Request selector": selector})
		endpoint["selector"] = selector
		endpoints = append(endpoints, endpoint)
	}
	createRequest["endpoints"] = endpoints

	tflog.Debug(ctx, "NetworkConnectionResource: create ## ", map[string]interface{}{"Create Request": createRequest})

	// send create request to server
	var request []map[string]interface{}
	request = append(request, createRequest)
	rb, err := json.Marshal(request)
	if err != nil {
		diags.AddError(
			"NetworkConnectionResource: create ##: Error Create LC",
			"Create: Could not Marshal NetworkConnectionResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/network-connections", rb)
	if err != nil {
		diags.AddError(
			"NetworkConnectionResource: create ##: Error create NetworkConnectionResource",
			"Create:Could not create NetworkConnectionResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "NetworkConnectionResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NCEndpointResource: create ##: Error Unmarshal response",
			"Create:Could not Create NCEndpointResource, unexpected error: "+err.Error(),
		)
		return
	}
	result := data[0].(map[string]interface{})

	href := result["href"].(string)
	splits := strings.Split(href, "/")
	id := splits[len(splits)-1]
	plan.Href = types.StringValue(href)
	plan.Id = types.StringValue(id)

	r.read(plan, ctx, diags, 5)
	if diags.HasError() {
		tflog.Debug(ctx, "NetworkResource: create failed. Can't find the created network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "NetworkConnectionResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *NetworkConnectionResource) update(plan *NetworkConnectionResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.Id.IsNull() {
		diags.AddError(
			"Error Update Module",
			"Update: Could not Update NC. ID and Href is not specified.",
		)
		return
	}
	tflog.Debug(ctx, "NetworkConnectionResource: Update ## PLAN##..", map[string]interface{}{"Id": plan.Id.ValueString(), "href": plan.Href.ValueString()})
	//tflog.Debug(ctx, "NetworkConnectionResource: Update ## STATE##..", map[string]interface{}{"Id": state.Id.ValueString(), "href": state.Href.ValueString()})

	if len(plan.Endpoints) < 2 {
		diags.AddError(
			"Error Update Module",
			"Update: Could not Update NC. There must be at least two endpoints.",
		)
		return
	}

	updateRequest := make(map[string]interface{})
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.ServiceMode.IsNull() {
		updateRequest["serviceMode"] = plan.Config.ServiceMode.ValueString()
	}
	if !plan.Config.MC.IsNull() {
		updateRequest["mc"] = plan.Config.MC.ValueString()
	}
	if !plan.Config.OuterVID.IsNull() {
		updateRequest["outerVID"] = plan.Config.OuterVID.ValueString()
	}
	if !plan.Config.ImplicitTransportCapacity.IsNull() {
		updateRequest["implicitTransportCapacity"] = plan.Config.ImplicitTransportCapacity.ValueString()
	}

	if !plan.Config.Labels.IsNull() {
		labels := map[string]string{}
		diag := plan.Config.Labels.ElementsAs(ctx, &labels, true)
		if !diag.HasError() {
			updateRequest["labels"] = labels
		}
	}

	tflog.Debug(ctx, "NetworkConnectionResource: update ## ", map[string]interface{}{"id": plan.Id.ValueString(), "Update Request": updateRequest})

	if len(updateRequest) > 0 {
		// send Update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"NetworkConnectionResource: Update ##: Error Update LC",
				"Update: Could not Marshal NetworkConnectionResource, unexpected error: "+err.Error(),
			)
			return
		}
		tflog.Debug(ctx, "NetworkConnectionResource: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"cmd": "PUT /network-connections/" + plan.Id.ValueString()})
		body, err := r.client.ExecuteIPMHttpCommand("PUT", "/network-connections/"+plan.Id.ValueString(), rb)
		if err != nil {
			diags.AddError(
				"NetworkConnectionResource: Update ##: Error update NetworkConnectionResource",
				"Create:Could not Update NetworkConnectionResource, unexpected error: "+err.Error(),
			)
			return
		}
		tflog.Debug(ctx, "NetworkConnectionResource: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data = make(map[string]interface{})
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"NetworkConnectionResource: Update ##: Error Unmarshal NetworkConnectionResource",
				"Update:Could not Update NetworkConnectionResource, unexpected error: "+err.Error(),
			)
			return
		}
	}
	/*// get current endpoints in network
	body, err := r.client.ExecuteIPMHttpCommand("GET", "/network-connections/"+plan.Id.ValueString()+"?content=expanded", nil)
	if err != nil {
		diags.AddError(
			"NetworkConnectionResource: Get ##: Error Get NetworkConnectionResource",
			"Update:Could not Get NetworkConnectionResource, unexpected error: "+err.Error(),
		)
		return
	}
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NetworkConnectionResource: Update ##: Error Unmarshal response",
			"Update:Could not Update NetworkConnectionResource, unexpected error: "+err.Error(),
		)
		return
	}

	nwEndpoints := data["endpoints"].([]interface{})
	updatedEndpointIds := []string{}
	// Check for Create new endpoint or update existing endpoint
	for _, e := range plan.Endpoints {
		if e.Id.IsNull() {
			// add new endpoint
			e.Create(r.client, ctx, diags)
			if diags.HasError() {
				return
			}
		} else {
			// Update endpoint
			e.Update(r.client, ctx, diags)
			if diags.HasError() {
				return
			}
			updatedEndpointIds = append(updatedEndpointIds, e.Id.ValueString())
		}
	}

	for _, e := range nwEndpoints {
		endpoint := e.((map[string]interface{}))
		if !common.Contains(updatedEndpointIds, endpoint["id"].(string)) {
			//delete the endpoint
			_, err := r.client.ExecuteIPMHttpCommand("DELETE", "/network-connections/"+plan.Id.ValueString()+"/endpoints/"+endpoint["id"].(string), nil)
			if err != nil {
				diags.AddError(
					"NetworkConnectionResource: delete ##: Error Delete endpoint "+endpoint["id"].(string)+" for NetworkConnection "+plan.Id.ValueString(),
					"Delete:Could not Delete endpoint "+endpoint["id"].(string)+" for NetworkConnection "+plan.Id.ValueString()+", unexpected error: "+err.Error(),
				)
				return
			}
		}
	}*/

	// Check for Update existing endpoint
	//if plan.Config.CapacityMode.ValueString() != "portMode" {
	/*		for _, ep := range plan.Endpoints {
			ep.Update(r.client, ctx, diags)
			if diags.HasError() {
				return
			}
		}*/
	//}

	r.read(plan, ctx, diags, 3)

	tflog.Debug(ctx, "NetworkConnectionResource: update ## ", map[string]interface{}{"plan": plan})
}

func (r *NetworkConnectionResource) read(state *NetworkConnectionResourceData, ctx context.Context, diags *diag.Diagnostics, retryCount ...int) {

	numRetry := 1
	if len(retryCount) > 0 {
		numRetry = retryCount[0]
	}

	tflog.Debug(ctx, "NetworkConnectionResource: read ##", map[string]interface{}{"id": state.Id.ValueString()})
	queryString := "?content=expanded"
	if state.Id.IsNull() {
		queryString = queryString + "&q={\"$and\":[{\"config.serviceMode\":\""+ state.Config.ServiceMode.ValueString() +"\"},{\"endpoints\":{\"$elemMatch\":{\"config.capacity\":"+ strconv.Itoa((int)(state.Endpoints[0].Config.Capacity.ValueInt64())) + ",\"config.selector.moduleIfSelectorByModuleName.moduleName\":\"" + state.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString() +"\",\"config.selector.moduleIfSelectorByModuleName.moduleClientIfAid\":\"" + state.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString() + "\"}}},{\"endpoints\":{\"$elemMatch\":{\"config.capacity\":"+ strconv.Itoa((int)(state.Endpoints[0].Config.Capacity.ValueInt64())) + ",\"config.selector.moduleIfSelectorByModuleName.moduleName\":\"" + state.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString() +"\",\"config.selector.moduleIfSelectorByModuleName.moduleClientIfAid\":\"" + state.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString() + "\"}}}]}"
	} else {
		queryString = "/" + state.Id.ValueString() + queryString
	}
	var err error
	body := []byte{}
	for i := 1; i <= numRetry; i++ {
		body, err = r.client.ExecuteIPMHttpCommand("GET", "/network-connections"+queryString, nil)
		if err == nil {
			break
		} else if i < numRetry {
			time.Sleep(3 * time.Second)
		}
	}
	if err != nil {
		diags.AddError(
			"NetworkConnectionResource: Get ##: Error Get NetworkConnectionResource",
			"Update:Could not Get NetworkConnectionResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "NetworkConnectionResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"NetworkConnectionResource: read ##: Error Unmarshal response",
			"Update:Could not read NetworkConnectionResource, unexpected error: "+err.Error(),
		)
		return
	}
	// populate network state
	state.Populate(data, ctx, diags)
	tflog.Debug(ctx, "NetworkConnectionResource: read SUCCESS ")
}

func (r *NetworkConnectionResource) delete(plan *NetworkConnectionResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.Id.IsNull() {
		diags.AddError(
			"Error Delete NetworkConnectionResource",
			"Read: Could not delete. NC ID is not specified",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", "/network-connections/"+plan.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"NetworkConnectionResource: delete ##: Error Delete NetworkConnectionResource",
			"Update:Could not delete NetworkConnectionResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "NetworkConnectionResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (ncData *NetworkConnectionResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}
	tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## ", map[string]interface{}{"plan": data})
	//populate state
	if computeFlag {
		ncData.Id = types.StringValue(data["id"].(string))
	}
	ncData.Href = types.StringValue(data["href"].(string))

	// populate Config
	if data["config"] != nil {
		if ncData.Config == nil {
			ncData.Config = &NetworkConnectionConfig{}
		}
		ncConfig := data["config"].(map[string]interface{})
		for k, v := range ncConfig {
			switch k {
			case "name":
				if !ncData.Config.Name.IsNull() || computeFlag {
					ncData.Config.Name = types.StringValue(v.(string))
				}
			case "serviceMode":
				if !ncData.Config.ServiceMode.IsNull() || computeFlag {
					ncData.Config.ServiceMode = types.StringValue(v.(string))
				}
			case "mc":
				if !ncData.Config.MC.IsNull() || computeFlag {
					ncData.Config.MC = types.StringValue(v.(string))
				}
			case "outerVID":
				if !ncData.Config.OuterVID.IsNull() || computeFlag {
					ncData.Config.OuterVID = types.StringValue(v.(string))
				}
			case "implicitTransportCapacity":
				if !ncData.Config.ImplicitTransportCapacity.IsNull() || computeFlag {
					ncData.Config.ImplicitTransportCapacity = types.StringValue(v.(string))
				}
			case "labels":
				labels := types.MapNull(types.StringType)
				labelData := make(map[string]attr.Value)
				for k, label := range v.(map[string]interface{}) {
					labelData[k] = types.StringValue(label.(string))
				}
				labels = types.MapValueMust(types.StringType, labelData)
				if !ncData.Config.Labels.IsNull() || computeFlag {
					ncData.Config.Labels = labels
				}
			}
		}
	}

  // populate  endpoints
  tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## Endpoint ***  ")
	endpoints := []NCEndpointResourceData{}
	if data["endpoints"] != nil {
		for _, ep := range ncData.Endpoints {
			if ep.Config.Selector.ModuleIfSelectorByModuleName != nil {
				planModuleName :=  ep.Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString()
				ncEndpoints := data["endpoints"].([]interface{})
				tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## Endpoint Module Name  ", map[string]interface{}{"Plan Module Name": planModuleName})
				for _, v2 := range ncEndpoints {
					epData := v2.(map[string]interface{})
					config := epData["config"].(map[string]interface{})
					selector := config["selector"].(map[string]interface{})
					moduleIfSelectorByModuleName := selector["moduleIfSelectorByModuleName"].(map[string]interface{})
					moduleName := moduleIfSelectorByModuleName["moduleName"].(string)
					tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## Endpoint ID  ", map[string]interface{}{"module name": moduleName})
					if planModuleName == moduleName {
						endpoint := NCEndpointResourceData{}
						endpoint.Populate(epData, ctx, diags, true)
						endpoints = append(endpoints, endpoint)
					}
				}
			} else if ep.Config.Selector.ModuleIfSelectorByModuleId != nil {
				planModuleId :=  ep.Config.Selector.ModuleIfSelectorByModuleId.ModuleId.ValueString()
				ncEndpoints := data["endpoints"].([]interface{})
				tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## Endpoint Module Id  ", map[string]interface{}{"Plan Module Id": planModuleId})
				for _, v2 := range ncEndpoints {
					epData := v2.(map[string]interface{})
					config := epData["config"].(map[string]interface{})
					selector := config["selector"].(map[string]interface{})
					moduleIfSelectorByModuleId := selector["moduleIfSelectorByModuleId"].(map[string]interface{})
					moduleId := moduleIfSelectorByModuleId["moduleId"].(string)
					tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## Endpoint ID  ", map[string]interface{}{"module name": moduleId})
					if planModuleId == moduleId {
						endpoint := NCEndpointResourceData{}
						endpoint.Populate(epData, ctx, diags, true)
						endpoints = append(endpoints, endpoint)
					}
				}
			} else if ep.Config.Selector.ModuleIfSelectorByModuleMAC != nil {
				planModuleMAC :=  ep.Config.Selector.ModuleIfSelectorByModuleMAC.ModuleMAC.ValueString()
				ncEndpoints := data["endpoints"].([]interface{})
				tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## Endpoint Module MAC  ", map[string]interface{}{"Plan Module MAC": planModuleMAC})
				for _, v2 := range ncEndpoints {
					epData := v2.(map[string]interface{})
					config := epData["config"].(map[string]interface{})
					selector := config["selector"].(map[string]interface{})
					moduleIfSelectorByModuleMAC := selector["moduleIfSelectorByModuleMAC"].(map[string]interface{})
					moduleMAC := moduleIfSelectorByModuleMAC["moduleMAC"].(string)
					tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## Endpoint Module MAC  ", map[string]interface{}{"module MAC": moduleMAC})
					if planModuleMAC == moduleMAC {
						endpoint := NCEndpointResourceData{}
						endpoint.Populate(epData, ctx, diags, true)
						endpoints = append(endpoints, endpoint)
					}
				}
			} else if ep.Config.Selector.ModuleIfSelectorByModuleSerialNumber != nil {
				planModuleSerialNumber :=  ep.Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleSerialNumber.ValueString()
				ncEndpoints := data["endpoints"].([]interface{})
				tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## Endpoint Module SN  ", map[string]interface{}{"Plan Module SN": planModuleSerialNumber})
				for _, v2 := range ncEndpoints {
					epData := v2.(map[string]interface{})
					config := epData["config"].(map[string]interface{})
					selector := config["selector"].(map[string]interface{})
					moduleIfSelectorByModuleSerialNumber := selector["moduleIfSelectorByModuleSerialNumber"].(map[string]interface{})
					ModuleSerialNumber := moduleIfSelectorByModuleSerialNumber["moduleSerialNumber"].(string)
					tflog.Debug(ctx, "NetworkConnectionResourceData: populate ## Endpoint Module SN  ", map[string]interface{}{"module SN": ModuleSerialNumber})
					if planModuleSerialNumber == ModuleSerialNumber {
						endpoint := NCEndpointResourceData{}
						endpoint.Populate(epData, ctx, diags, true)
						endpoints = append(endpoints, endpoint)
					}
				}
			}
		}
	}
	ncData.Endpoints = endpoints

	// popuplate state
	if data["state"] != nil {
		ncData.State = types.ObjectValueMust(
			NetworkConnectionStateAttributeType(), NetworkConnectionStateAttributeValue(data["state"].(map[string]interface{})))
	}

	// populate LCss
	ncData.LCs = types.ListNull(LCObjectType())
	if data["lcs"] != nil {
		ncData.LCs = types.ListValueMust(LCObjectType(), LCObjectsValue(data["lcs"].([]interface{})))
	}

	tflog.Debug(ctx, "NetworkConnectionResourceData: read ## ", map[string]interface{}{"plan": ncData})
}

func NetworkConnectionSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Network Connection",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "Href of the Network Connection",
			Computed:    true,
		},
		//Config      NetworkConnectionConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Description: " NetworkConnection Config",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "name",
					Optional:    true,
				},
				"service_mode": schema.StringAttribute{
					Description: "service_mode",
					Optional:    true,
				},
				"mc": schema.StringAttribute{
					Description: "mc",
					Optional:    true,
				},
				"outer_vid": schema.StringAttribute{
					Description: "outer_vid",
					Optional:    true,
				},
				"implicit_transport_capacity": schema.StringAttribute{
					Description: "implicit_transport_capacity",
					Optional:    true,
				},
				"labels": schema.MapAttribute{
					Description: "labels",
					Optional:    true,
					ElementType: types.StringType,
				},
			},
		},
		//State      NetworkConnectionConfig `tfsdk:"state"`
		"state": schema.SingleNestedAttribute{
			Description: "Module",
			Computed:    true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "name",
					Computed:    true,
				},
				"service_mode": schema.StringAttribute{
					Description: "service_mode",
					Computed:    true,
				},
				"managed_by": schema.StringAttribute{
					Description: "managed_by",
					Computed:    true,
				},
				"lifecycle_state": schema.StringAttribute{
					Description: "lifecycle_state",
					Computed:    true,
				},
				"operational_status": schema.StringAttribute{
					Description: "OperationalStatus",
					Computed:    true,
				},
				"lifecycle_state_cause": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"action": schema.Int64Attribute{
							Description: "Numeric identifier of the Network.",
							Computed:    true,
						},
						"timestamp": schema.StringAttribute{
							Description: "timestamp",
							Computed:    true,
						},
						"trace_id": schema.StringAttribute{
							Description: "TraceId",
							Computed:    true,
						},
						"errors": schema.ListNestedAttribute{
							Description: "errors",
							Required:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"code": schema.StringAttribute{
										Description: "code",
										Computed:    true,
									},
									"message": schema.StringAttribute{
										Description: "Message",
										Computed:    true,
									},
								},
							},
						},
					},
				},
				"labels": schema.MapAttribute{
					Description: "capabilities",
					Optional:    true,
					ElementType: types.StringType,
				},
			},
		},
		"end_points": schema.ListNestedAttribute{
			Description: "List of NC Endpoints",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: NCEndpointSchemaAttributes(),
			},
		},
		"lcs": schema.ListAttribute{
			Computed:    true,
			ElementType: LCObjectType(),
		},
	}
}

func NetworkConnectionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: NetworkConnectionAttributeType(),
	}
}

func NetworkConnectionObjectsValue(data []interface{}) []attr.Value {
	endpoints := []attr.Value{}
	for _, v := range data {
		endpoint := v.(map[string]interface{})
		if endpoint != nil {
			endpoints = append(endpoints, types.ObjectValueMust(
				NetworkConnectionAttributeType(),
				NetworkConnectionAttributeValue(endpoint)))
		}
	}
	return endpoints
}

func NetworkConnectionAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":         types.StringType,
		"href":       types.StringType,
		"config":     types.ObjectType{AttrTypes: NetworkConnectionConfigAttributeType()},
		"state":      types.ObjectType{AttrTypes: NetworkConnectionStateAttributeType()},
		"end_points": types.ListType{ElemType: NCEndpointObjectType()},
		"lcs":        types.ListType{ElemType: LCObjectType()},
	}
}

func NetworkConnectionAttributeValue(networkConnection map[string]interface{}) map[string]attr.Value {
	id := types.StringNull()
	if networkConnection["id"] != nil {
		id = types.StringValue(networkConnection["id"].(string))
	}
	href := types.StringNull()
	if networkConnection["href"] != nil {
		href = types.StringValue(networkConnection["href"].(string))
	}
	config := types.ObjectNull(NetworkConnectionConfigAttributeType())
	if (networkConnection["config"]) != nil {
		config = types.ObjectValueMust(NetworkConnectionConfigAttributeType(), NetworkConnectionConfigAttributeValue(networkConnection["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(NetworkConnectionStateAttributeType())
	if (networkConnection["state"]) != nil {
		state = types.ObjectValueMust(NetworkConnectionStateAttributeType(), NetworkConnectionStateAttributeValue(networkConnection["state"].(map[string]interface{})))
	}
	endpoints := types.ListNull(LCObjectType())
	if (networkConnection["endpoints"]) != nil {
		endpoints = types.ListValueMust(NCEndpointObjectType(), NCEndpointObjectsValue(networkConnection["endpoints"].([]interface{})))
	}
	lcs := types.ListNull(LCObjectType())
	if (networkConnection["lcs"]) != nil {
		lcs = types.ListValueMust(LCObjectType(), LCObjectsValue(networkConnection["lcs"].([]interface{})))
	}

	return map[string]attr.Value{
		"id":         id,
		"href":       href,
		"config":     config,
		"state":      state,
		"end_points": endpoints,
		"lcs":        lcs,
	}
}

func NetworkConnectionStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                  types.StringType,
		"service_mode":          types.StringType,
		"managed_by":            types.StringType,
		"lifecycle_state":       types.StringType,
		"lifecycle_state_cause": types.ObjectType{AttrTypes: common.LifecycleStateCauseAttributeType()},
		"operational_status":    types.StringType,
		"labels":                types.MapType{ElemType: types.StringType},
	}
}

func NetworkConnectionStateAttributeValue(state map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	if state["name"] != nil {
		name = types.StringValue(state["name"].(string))
	}
	serviceMode := types.StringNull()
	if state["serviceMode"] != nil {
		serviceMode = types.StringValue(state["serviceMode"].(string))
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
		lifecycleStateCause = types.ObjectValueMust(common.LifecycleStateCauseAttributeType(), common.LifecycleStateCauseAttributeValue(state["lifecycleStateCause"].(map[string]interface{})))
	}
	operationalStatus := types.StringNull()
	if state["operationalStatus"] != nil {
		operationalStatus = types.StringValue(state["operationalStatus"].(string))
	}
	labels := types.MapNull(types.StringType)
	if state["labels"] != nil {
		data := make(map[string]attr.Value)
		for k, v := range state["labels"].(map[string]interface{}) {
			data[k] = types.StringValue(v.(string))
		}
		labels = types.MapValueMust(types.StringType, data)
	}
	return map[string]attr.Value{
		"name":                  name,
		"service_mode":          serviceMode,
		"managed_by":            managedBy,
		"lifecycle_state":       lifecycleState,
		"lifecycle_state_cause": lifecycleStateCause,
		"operational_status":    operationalStatus,
		"labels":                labels,
	}
}

func NetworkConnectionConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                        types.StringType,
		"service_mode":                types.StringType,
		"mc":                          types.StringType,
		"outer_vid":                   types.StringType,
		"implicit_transport_capacity": types.StringType,
		"labels":                      types.MapType{ElemType: types.StringType},
	}
}

func NetworkConnectionConfigAttributeValue(config map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	if config["name"] != nil {
		name = types.StringValue(config["name"].(string))
	}
	serviceMode := types.StringNull()
	if config["serviceMode"] != nil {
		serviceMode = types.StringValue(config["serviceMode"].(string))
	}
	mc := types.StringNull()
	if config["mc"] != nil {
		mc = types.StringValue(config["mc"].(string))
	}
	outerVID := types.StringNull()
	if config["outer_vid"] != nil {
		outerVID = types.StringValue(config["outerVID"].(string))
	}
	implicitTransportCapacity := types.StringNull()
	if config["implicitTransportCapacity"] != nil {
		implicitTransportCapacity = types.StringValue(config["implicitTransportCapacity"].(string))
	}
	labels := types.MapNull(types.StringType)
	if config["labels"] != nil {
		data := make(map[string]attr.Value)
		for k, v := range config["labels"].(map[string]interface{}) {
			data[k] = types.StringValue(v.(string))
		}
		labels = types.MapValueMust(types.StringType, data)
	}

	return map[string]attr.Value{
		"name":                        name,
		"service_mode":                serviceMode,
		"mc":                          mc,
		"outer_vid":                   outerVID,
		"implicit_transport_capacity": implicitTransportCapacity,
		"labels":                      labels,
	}
}
