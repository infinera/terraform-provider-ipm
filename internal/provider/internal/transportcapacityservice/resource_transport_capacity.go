package transportcapacity

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"terraform-provider-ipm/internal/ipm_pf"
	"terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &TransportCapacityResource{}
	_ resource.ResourceWithConfigure   = &TransportCapacityResource{}
	_ resource.ResourceWithImportState = &TransportCapacityResource{}
)

// NewTransportCapacityResource is a helper function to simplify the provider implementation.
func NewTransportCapacityResource() resource.Resource {
	return &TransportCapacityResource{}
}

type TransportCapacityResource struct {
	client *ipm_pf.Client
}


type TCConfig struct {
	Name            types.String  `tfsdk:"name"`
	CapacityMode    types.String  `tfsdk:"capacity_mode"`
  Labels          types.Map    `tfsdk:"labels"`
}

type TransportCapacityResourceData struct {
	Id               types.String                  `tfsdk:"id"`
	Href             types.String                  `tfsdk:"href"`
	Config           *TCConfig                      `tfsdk:"config"`
	State            types.Object                  `tfsdk:"state"`
	Endpoints        []TCEndpointResourceData      `tfsdk:"end_points"`
	CapacityLinks    types.List                    `tfsdk:"capacity_links"`
}

// Metadata returns the data source type name.
func (r *TransportCapacityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transport_capacity"
}

// Schema defines the schema for the data source.
func (r *TransportCapacityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type TransportCapacityResourceData struct 
	resp.Schema = schema.Schema{
		Description: "Manages an TransportCapacity",
		Attributes: TransportCapacitySchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *TransportCapacityResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r TransportCapacityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TransportCapacityResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "TransportCapacityResource: Create - ", map[string]interface{}{"TransportCapacityResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, &data)

	resp.Diagnostics.Append(diags...)
}

func (r TransportCapacityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TransportCapacityResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "TransportCapacityResource: Read - ", map[string]interface{}{"TransportCapacityResourceData": data})

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

func (r TransportCapacityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TransportCapacityResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "TransportCapacityResource: Update 222", map[string]interface{}{"id": data.Id.ValueString(), "TransportCapacityResourceData": data})

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

func (r TransportCapacityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TransportCapacityResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"TransportCapacityResourceData": data})

	resp.Diagnostics.Append(diags...)

		// DELAY TO MAKE SURE CONNECTION IS DELETED
		time.Sleep(1 * time.Second)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *TransportCapacityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TransportCapacityResource) create(plan *TransportCapacityResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TransportCapacityResource: create ## ", map[string]interface{}{"plan": plan})

	if len(plan.Endpoints) != 2  {
		diags.AddError(
			"TransportCapacityResource: Error create TransportCapacity",
			"Create: Could not Create TransportCapacity. Must have two and only two endpoints",
		)
		return
	}

	var createRequest = make(map[string]interface{})
	var config = make(map[string]interface{})

	// get TC config settings
	if !plan.Config.Name.IsNull() {
		config["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.CapacityMode.IsNull() {
		config["capacityMode"] = plan.Config.CapacityMode.ValueString()
	}
	if !plan.Config.Labels.IsNull() {
		labels := map[string]string{}
		diag := plan.Config.Labels.ElementsAs(ctx, &labels, true)
		if !diag.HasError() {
			config["labels"] = labels
		}
	}
	createRequest["config"] = config

	// get TC Endpoints
	var endpoints []map[string]interface{}
	var queryStrings []string
	var endpointIds []string
	for _, v := range plan.Endpoints {
		endpoint := make(map[string]interface{})
		queryString :=  "/xr-networks?content=expanded"
		id := ""
		endpoint["capacity"] = v.Config.Capacity.ValueInt64()
		selector := make(map[string]interface{})
		aSelector := make(map[string]interface{})
		if v.Config.Selector.ModuleIfSelectorByModuleId != nil {
			aSelector["moduleId"] = v.Config.Selector.ModuleIfSelectorByModuleId.ModuleId.ValueString()
			aSelector["moduleClientIfAid"] = v.Config.Selector.ModuleIfSelectorByModuleId.ModuleClientIfAid.ValueString()
			selector["moduleIfSelectorByModuleId"] = aSelector
			id = aSelector["moduleId"].(string)
			queryString = queryString + "&q={\"hubModule.state.module.moduleId\":\"" + id + "\"}"
		} else if v.Config.Selector.ModuleIfSelectorByModuleName != nil {
			aSelector["moduleName"] = v.Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString()
			aSelector["moduleClientIfAid"] = v.Config.Selector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString()
			tflog.Debug(ctx, "TransportCapacityResource: create ## moduleName", map[string]interface{}{"ModuleClientIfAid": aSelector["moduleClientIfAid"]})
			selector["moduleIfSelectorByModuleName"] = aSelector
			id = aSelector["moduleName"].(string)
			queryString = queryString + "&q={\"hubModule.state.module.moduleName\":\"" + id + "\"}"
		} else if v.Config.Selector.ModuleIfSelectorByModuleMAC != nil {
			aSelector["moduleMAC"] = v.Config.Selector.ModuleIfSelectorByModuleMAC.ModuleMAC.ValueString()
			aSelector["moduleClientIfAid"] = v.Config.Selector.ModuleIfSelectorByModuleMAC.ModuleClientIfAid.ValueString()
			selector["moduleIfSelectorByModuleMAC"] = aSelector
			id = aSelector["moduleMAC"].(string)
			queryString = queryString + "&q={\"hubModule.state.module.macAddress\":\"" + id + "\"}"
		} else if v.Config.Selector.ModuleIfSelectorByModuleSerialNumber != nil {
			aSelector["moduleSerialNumber"] = v.Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleSerialNumber.ValueString()
			aSelector["moduleClientIfAid"] = v.Config.Selector.ModuleIfSelectorByModuleSerialNumber.ModuleClientIfAid.ValueString()
			selector["moduleIfSelectorByModuleSerialNumber"] = aSelector
			id = aSelector["moduleSerialNumber"].(string)
			queryString = queryString + "&q={\"hubModule.state.module.serialNumber\":\"" + id + "\"}"
		} else if v.Config.Selector.HostPortSelectorByName != nil {
			aSelector["hostName"] = v.Config.Selector.HostPortSelectorByName.HostName.ValueString()
			aSelector["hostPortName"] = v.Config.Selector.HostPortSelectorByName.HostPortName.ValueString()
			selector["hostPortSelectorByName"] = aSelector
			id = aSelector["hostName"].(string)
			queryString = queryString + "&q={\"$and\":[{\"hubModule.config.selector.hostPortSelectorByName.hostName\":\"" + id + "\"}, {\"hubModule.config.selector.hostPortSelectorByName.hostPortName\":\"" + aSelector["hostPortName"].(string) + "\"}]}"
			id = id + ":" + aSelector["hostPortName"].(string)
		}  else if v.Config.Selector.HostPortSelectorByPortId != nil {
			aSelector["chassisIdSubtype"] = v.Config.Selector.HostPortSelectorByPortId.ChassisIdSubtype.ValueString()
			aSelector["chassisId"] = v.Config.Selector.HostPortSelectorByPortId.ChassisId.ValueString()
			aSelector["portIdSubtype"] = v.Config.Selector.HostPortSelectorByPortId.PortIdSubtype.ValueString()
			aSelector["portId"] = v.Config.Selector.HostPortSelectorByPortId.PortId.ValueString()
			selector["hostPortSelectorByPortId"] = aSelector
			id = aSelector["chassisId"].(string)
			queryString = queryString + "&q={\"$and\":[{\"hubModule.config.selector.hostPortSelectorByPortId.chassisIdSubtype\":\"" + aSelector["chassisIdSubtype"].(string) + "\"}, {\"hubModule.config.selector.hostPortSelectorByPortId.chassisId\":\"" + id + "\"}, {\"hubModule.config.selector.hostPortSelectorByPortId.portId\":\"" + aSelector["portId"].(string) + "\"}, {\"hubModule.config.selector.hostPortSelectorByPortId.portIdSubtype\":\"" + aSelector["portIdSubtype"].(string) + "\"}]}"
			id = id + ":" + aSelector["portId"].(string)
		} else if v.Config.Selector.HostPortSelectorBySysName != nil {
			aSelector["sysName"] = v.Config.Selector.HostPortSelectorBySysName.SysName.ValueString()
			aSelector["portIdSubtype"] = v.Config.Selector.HostPortSelectorBySysName.PortIdSubtype.ValueString()
			aSelector["portId"] = v.Config.Selector.HostPortSelectorBySysName.PortId.ValueString()
			selector["hostPortSelectorByPortId"] = aSelector
			id = aSelector["sysName"].(string)
			queryString = queryString + "&q={\"$and\":[{\"hubModule.config.selector.hostPortSelectorBySysName.sysName\":\"" + id + "\"}, {\"hubModule.config.selector.hostPortSelectorBySysName.portIdSubtype\":\"" + aSelector["portIdSubtype"].(string) + "\"}, {\"hubModule.config.selector.hostPortSelectorBySysName.portId\":\"" + aSelector["portId"].(string) + "\"}]"
			id = id + ":" + aSelector["portId"].(string)
		} else if v.Config.Selector.HostPortSelectorByPortSourceMAC != nil {
			aSelector["portSourceMAC"] = v.Config.Selector.HostPortSelectorByPortSourceMAC.PortSourceMAC.ValueString()
			selector["hostPortSelectorByName"] = aSelector
			id = aSelector["portSourceMAC"].(string)
			queryString = queryString + "&q={\"hubModule.config.selector.hostPortSelectorByPortSourceMAC.portSourceMAC\":\"" + id + "\"}"
		} else {
			diags.AddError(
				"TransportCapacityResource: Error Create TC. No selector specify for Endpoint",
				"Create: Could not create TransportCapacityResource, No selector specify for Endpoint",
			)
			return
		}
		tflog.Debug(ctx, "TransportCapacityResource: create 1## ", map[string]interface{}{"Create Request selector": selector})
		endpoint["selector"] = selector
		endpoints = append(endpoints, endpoint)
		queryStrings = append(queryStrings, queryString)
		endpointIds =  append(endpointIds, id)
	}

	tflog.Debug(ctx, "TransportCapacityResource: Network QueryString ## ", map[string]interface{}{"QueryString1": queryStrings[0], "QueryString2": queryStrings[1]})

	_, err := common.CheckNetworkState(ctx, r.client, queryStrings, endpointIds, 5) 
	if err != nil {
		diags.AddError(
			"Error Creating TC",
			"Create: Could not create TC:  CheckNetworkState :" + queryStrings[0] + ":" + queryStrings[1] + err.Error(),
		)
		return
	}

	createRequest["endpoints"] = endpoints

	tflog.Debug(ctx, "TransportCapacityResource: create 2## ", map[string]interface{}{"Create Request": createRequest})

	// send create request to server
	var request []map[string]interface{}
	request = append(request, createRequest)
	rb, err := json.Marshal(request)
	if err != nil {
		diags.AddError(
			"TransportCapacityResource: create ##: Error Create AC",
			"Create: Could not Marshal TransportCapacityResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/transport-capacities", rb)
	if err != nil {
		if !strings.Contains(err.Error(), "status: 202") {
			diags.AddError(
				"TransportCapacityResource: create ##: Error create TransportCapacityResource",
				"Create:Could not create TransportCapacityResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "TransportCapacityResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TransportCapacityResource: Create ##: Error Unmarshal response",
			"create:Could not Create TransportCapacityResource, unexpected error: "+err.Error(),
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
		tflog.Debug(ctx, "TransportCapacityResource: create failed. Can't find the created network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "TransportCapacityResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *TransportCapacityResource) update(plan *TransportCapacityResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TransportCapacityResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Id.IsNull(){
		diags.AddError(
			"TransportCapacityResource: Error Update TransportCapacity",
			"Update: Could not Update TransportCapacity. TC Id is not specified",
		)
		return
	}
	if len(plan.Endpoints) != 2  {
		diags.AddError(
			"TransportCapacityResource: Error Update TransportCapacity",
			"Update: Could not Update TransportCapacity. Must have two and only two endpoints",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get TC config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.Labels.IsNull() {
		labels := map[string]string{}
		diag := plan.Config.Labels.ElementsAs(ctx, &labels, true)
		if !diag.HasError() {
			updateRequest["labels"] = labels
		}
	}

	tflog.Debug(ctx, "TransportCapacityResource: update ## ", map[string]interface{}{"id": plan.Id.ValueString(),"Update Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"TransportCapacityResource: update ##: Error Create TransportCapacity",
				"Create: Could not Marshal TransportCapacityResource, unexpected error: "+err.Error(),
			)
			return
		}
		body, err := r.client.ExecuteIPMHttpCommand("PUT", "/transport-capacities/"+ plan.Id.ValueString(), rb)
		if err != nil {
			if !strings.Contains(err.Error(), "status: 202") {
				diags.AddError(
					"TransportCapacityResource: update ##: Error update TransportCapacityResource",
					"Create:Could not update TransportCapacityResource, unexpected error: "+err.Error(),
				)
				return
			}
		}

		tflog.Debug(ctx, "TransportCapacityResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"TransportCapacityResource: update ##: Error Unmarshal response",
				"Update:Could not update TransportCapacityResource, unexpected error: "+err.Error(),
			)
			return
		}
	}
	// Check for Update existing endpoint
	if plan.Config.CapacityMode.ValueString() != "portMode" {
		for _, ep := range plan.Endpoints {
			ep.Update(r.client, ctx, diags)
			if diags.HasError() {
				return
			}
		}
	}

	r.read(plan, ctx, diags, 2)
	if diags.HasError() {
		tflog.Debug(ctx, "TransportCapacityResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "TransportCapacityResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *TransportCapacityResource) read(state *TransportCapacityResourceData, ctx context.Context, diags *diag.Diagnostics, retryCount ...int) {

	numRetry := 1
	if len(retryCount) > 0 {
		numRetry = retryCount[0]
	}

	tflog.Debug(ctx, "TransportCapacityResource: read ## ", map[string]interface{}{"plan": state})
	queryString := "?content=expanded"
	if state.Id.IsNull() {
		queryString = queryString + "&q={\"$and\":[{\"+ config.capacityMode\":\"" + state.Config.CapacityMode.ValueString() + "\"},{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleName.moduleName\":\"" + state.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleName.moduleClientIfAid\":" + state.Endpoints[0].Config.Selector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString() + "\"}}},{\"endpoints\":{\"$elemMatch\":{\"config.selector.moduleIfSelectorByModuleName.moduleName\":\"" + state.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleName.ModuleName.ValueString() + "\",\"config.selector.moduleIfSelectorByModuleName.moduleClientIfAid\":" + state.Endpoints[1].Config.Selector.ModuleIfSelectorByModuleName.ModuleClientIfAid.ValueString() + "\"}}}]}"
	} else {
		queryString = "/" + state.Id.ValueString() + queryString
	}
	var err error
	body := []byte{}
	for i := 1; i <= numRetry; i++ {
		body, err = r.client.ExecuteIPMHttpCommand("GET", "/transport-capacities"+queryString, nil)
		if err == nil {
			break
		} else if i < numRetry {
			time.Sleep(2 * time.Second)
		}
	} 

	if err != nil {
		diags.AddError(
			"TransportCapacityResource: read ##: Error Read TransportCapacity",
			"Read:Could not get Network, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "TransportCapacityResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TransportCapacityResource: read ##: Error Read TransportCapacity",
			"Read:Could not get Network, unexpected error: "+err.Error(),
		)
		return
	}
	// populate network state
	state.Populate(data, ctx, diags)
	tflog.Debug(ctx, "TransportCapacityResource: read SUCCESS ")
}

func (r *TransportCapacityResource) delete(plan *TransportCapacityResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "TransportCapacityResource: delete ## ", map[string]interface{}{"plan": plan})
	if plan.Id.IsNull() {
		diags.AddError(
			"TransportCapacityResource: Error Delete TransportCapacity",
			"Delete: Could not delete. TransportCapacity Id is not specified",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", "/transport-capacities/"+ plan.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"TransportCapacityResource: delete ##: Error Delete TransportCapacity",
			"Update:Could not delete TransportCapacity, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "NetworkResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (tcData *TransportCapacityResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}

	tflog.Debug(ctx, "TransportCapacityResourceData: populate ## ", map[string]interface{}{"computeFlag": computeFlag, "data": data})
	if computeFlag {
		tcData.Id = types.StringValue(data["id"].(string))
	}
	tcData.Href = types.StringValue(data["href"].(string))

	tflog.Debug(ctx, "TransportCapacityResourceData: populate Config## ")
	// populate Config
	if data["config"] != nil {
		if tcData.Config == nil {
			tcData.Config = &TCConfig{}
		}
		tcConfig := data["config"].(map[string]interface{})
		labels := types.MapNull(types.StringType)
		if tcConfig["labels"] != nil {
			data := make(map[string]attr.Value)
			for k, v := range tcConfig["labels"].(map[string]interface{}) {
				data[k] = types.StringValue(v.(string))
			}
			labels = types.MapValueMust(types.StringType, data)
		} 
		if !tcData.Config.Labels.IsNull() || computeFlag {
			tcData.Config.Labels = labels
		}
		for k, v := range tcConfig {
			switch k {
			case "name": 
				if !tcData.Config.Name.IsNull() || computeFlag {
					tcData.Config.Name = types.StringValue(v.(string))
				}
			case "capacityMode": 
				if !tcData.Config.CapacityMode.IsNull() || computeFlag {
					tcData.Config.CapacityMode = types.StringValue(v.(string))
				}	
			}
		}
	}
	tflog.Debug(ctx, "TransportCapacityResourceData: populate State## ")
	// populate state
	if data["state"] != nil {
		tcData.State =types.ObjectValueMust(
			TCStateAttributeType(),TCStateAttributeValue(data["state"].(map[string]interface{})))
	}

	// populate Endpoints
	if data["endpoints"] != nil {
		tcData.Endpoints = []TCEndpointResourceData{}
		for _, v := range data["endpoints"].([]interface{}) {
			endpoint := TCEndpointResourceData{}
			endpointData := v.(map[string]interface{})
			endpoint.Populate(endpointData,ctx, diags, true)
			tcData.Endpoints = append(tcData.Endpoints, endpoint)
		}
	}

	// populate CapacityLinks
	if data["capacityLinks"] != nil {
		tcData.CapacityLinks = types.ListValueMust(
			TCCapacityLinkObjectType(),TCCapacityLinksAttributeValue(data["capacityLinks"].([]interface{})))
	}
	tflog.Debug(ctx, "TransportCapacityResourceData: populate SUCCESS ")
}


func TransportCapacitySchemaAttributes() map[string]schema.Attribute {
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
					Description: "TC Name",
					Optional:    true,
				},
				"capacity_mode": schema.StringAttribute{
					Description: "capacity_mode",
					Optional:    true,
				},
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
			AttributeTypes: TCStateAttributeType(),
		},
		//LeafModules      types.List `tfsdk:"leaf_modules"`
		"end_points":schema.ListNestedAttribute{
			Optional:     true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"tc_id": schema.StringAttribute{
						Description: "Numeric identifier of the TC.",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
					},
					"id": schema.StringAttribute{
						Description: "Numeric identifier of the Endpoint.",
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
							"capacity": schema.Int64Attribute{
								Description: "capacity Name",
								Optional:    true,
							},
							"selector": common.IfSelectorSchema(),
						},
					},
					//State     types.Object   `tfsdk:"state"`
					"state": schema.ObjectAttribute{
						Computed: true,
						AttributeTypes: TCEndpointStateAttributeType(),
					},
				},
			},
		},
		"capacity_links":schema.ListAttribute{
			Computed: true,
			ElementType: TCCapacityLinkObjectType(),
		},
	}
}

func TCStateAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"name":    types.StringType,
		"managed_by":   types.StringType,
		"capacity_mode" : types.StringType,
		"life_cycle_state"  : types.StringType,
		"life_cycle_state_cause" : types.ObjectType{AttrTypes: common.LifecycleStateCauseAttributeType()},
		"labels": types.MapType {ElemType:  types.StringType},
	}
}

func TCStateAttributeValue(TCState map[string]interface{}) (map[string]attr.Value) {
	name := types.StringNull()
	if TCState["name"] != nil {
		name = types.StringValue(TCState["name"].(string))
	}
	managedBy := types.StringNull()
	if TCState["managedBy"] != nil {
		managedBy = types.StringValue(TCState["managedBy"].(string))
	}
	capacityMode := types.StringNull()
	if TCState["capacityMode"] != nil {
		capacityMode = types.StringValue(TCState["capacityMode"].(string))
	}
	lifecycleState := types.StringNull()
	if TCState["lifecycleState"] != nil {
		lifecycleState = types.StringValue(TCState["lifecycleState"].(string))
	}
	lifecycleStateCause := types.ObjectNull(common.LifecycleStateCauseAttributeType())
	if TCState["lifecycleStateCause"] != nil {
		lifecycleStateCause = types.ObjectValueMust(common.LifecycleStateCauseAttributeType(), common.LifecycleStateCauseAttributeValue(TCState["lifecycleStateCause"].(map[string]interface{})))
	}
	labels := types.MapNull(types.StringType)
	if TCState["labels"] != nil {
		data := make(map[string]attr.Value)
		for k, v := range TCState["labels"].(map[string]interface{}) {
			data[k] = types.StringValue(v.(string))
		}
		labels, _ = types.MapValue(types.StringType, data)
	}
	return map[string]attr.Value {
		"name":  name,
		"managed_by":  managedBy,
		"capacity_mode":  capacityMode,
		"life_cycle_state": lifecycleState,
		"life_cycle_state_cause" : lifecycleStateCause,
		"labels": labels,
	}
}
