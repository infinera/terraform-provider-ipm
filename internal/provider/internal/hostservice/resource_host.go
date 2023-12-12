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
	_ resource.Resource                = &HostResource{}
	_ resource.ResourceWithConfigure   = &HostResource{}
	_ resource.ResourceWithImportState = &HostResource{}
)

// NewModuleResource is a helper function to simplify the provider implementation.
func NewHostResource() resource.Resource {
	return &HostResource{}
}

type HostResource struct {
	client *ipm_pf.Client
}

type HostLocation struct {
	Latitude types.Int64 `tfsdk:"latitude"`
	Longitude types.Int64 `tfsdk:"longitude"`
}

type HostSelector struct {
	ModuleSelectorByModuleId           *common.ModuleSelectorByModuleId           `tfsdk:"module_selector_by_module_id"`
	ModuleSelectorByModuleName         *common.ModuleSelectorByModuleName         `tfsdk:"module_selector_by_module_name"`
	ModuleSelectorByModuleMAC          *common.ModuleSelectorByModuleMAC          `tfsdk:"module_selector_by_module_mac"`
	ModuleSelectorByModuleSerialNumber *common.ModuleSelectorByModuleSerialNumber `tfsdk:"module_selector_by_module_serial_number"`
	HostSelectorByHostChassisId        *common.HostSelectorByHostChassisId   `tfsdk:"host_selector_by_host_chassis_id"`
}

type HostConfig struct {
	Name             types.String `tfsdk:"name"`
	ManagedBy        types.String `tfsdk:"managed_by"`
	Location         HostLocation `tfsdk:"location"`
	Selector         HostSelector `tfsdk:"selector"`
	Labels           types.Map    `tfsdk:"labels"`
}

type HostResourceData struct {
	Id        types.String `tfsdk:"id"`
	Href      types.String `tfsdk:"href"`
	Config    *HostConfig   `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
	Ports     types.List   `tfsdk:"ports"`
}

// Metadata returns the data source type name.
func (r *HostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

// Schema defines the schema for the data source.
func (r *HostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type HostResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an Host",
		Attributes:  HostSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *HostResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r HostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HostResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "NetworkResource: Create - ", map[string]interface{}{"NetworkResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)
	resp.State.Set(ctx, &data)

}

func (r HostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HostResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "HostResource: Read - ", map[string]interface{}{"HostResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r HostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HostResourceData

	tflog.Debug(ctx, "HostResource: Update 1", map[string]interface{}{"UpdateRequest": req.Plan})

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "HostResource: Update", map[string]interface{}{"HostResourceData": data})

	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r HostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HostResourceData

	diags := req.State.Get(ctx, &data)
	diags.AddError(
		"Error Delete Hub Module",
		"Read: Could not delete Hub Module. It is deleted together with its Network",
	)
	resp.Diagnostics.Append(diags...)
}

func (r *HostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *HostResource) create(plan *HostResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "HostResource: create ## ", map[string]interface{}{"plan": plan})

	var createRequest = make(map[string]interface{})

	// get Network config settings
	if !plan.Config.Name.IsNull() {
		createRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.ManagedBy.IsNull() {
		createRequest["managedBy"] = plan.Config.ManagedBy.ValueString()
	}
	if !plan.Config.Location.Latitude.IsNull() {
		location := make(map[string]interface{})
		location["latitude"] = plan.Config.Location.Latitude.ValueInt64()
		location["longitude"] = plan.Config.Location.Longitude.ValueInt64()
		createRequest["location"] = location
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
	} else if plan.Config.Selector.HostSelectorByHostChassisId != nil {
		aSelector["chassisId"] = plan.Config.Selector.HostSelectorByHostChassisId.ChassisId.ValueString()
		aSelector["chassisIdSubtype"] = plan.Config.Selector.HostSelectorByHostChassisId.ChassisIdSubtype.ValueString()
		selector["hostPortSelectorByChassisId"] = aSelector
	} else {
		diags.AddError(
			"Error Create HostResource",
			"Create: Could not create HostResource, No hub module selector specified",
		)
		return
	}
	createRequest["selector"] = selector

	tflog.Debug(ctx, "HostResource: create ## ", map[string]interface{}{"Create Request": createRequest})

	// send create request to server
	var request []map[string]interface{}
	request = append(request, createRequest)
	rb, err := json.Marshal(request)
	if err != nil {
		diags.AddError(
			"HostResource: create ##: Error Create AC",
			"Create: Could not Marshal HostResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/hosts", rb)
	if err != nil {
		if !strings.Contains(err.Error(), "status: 202") {
			diags.AddError(
				"HostResource: create ##: Error create HostResource",
				"Create:Could not create HostResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "HostResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HostResource: Create ##: Error Unmarshal response",
			"Update:Could not Create HostResource, unexpected error: "+err.Error(),
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
		tflog.Debug(ctx, "HostResource: create failed. Can't find the created network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "HostResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *HostResource) update(plan *HostResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "HostResource: update - plan", map[string]interface{}{"plan": plan})
	if plan.Id.IsNull() {
		diags.AddError(
			"HostResource: Error Update Host",
			"Update: Could not Update Host. Id is not specified",
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
	if !plan.Config.Location.Latitude.IsNull() {
		location := make(map[string]interface{})
		location["latitude"] = plan.Config.Location.Latitude.ValueInt64()
		location["longitude"] = plan.Config.Location.Longitude.ValueInt64()
		updateRequest["location"] = location
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
			"HostResource: Update ##: Error Update AC",
			"Update: Could not Marshal HostResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HostResource: update - rb", map[string]interface{}{"rb": rb})

	body, err := r.client.ExecuteIPMHttpCommand("PUT", "/hosts/"+plan.Id.ValueString(), rb)

	if err != nil {
		diags.AddError(
			"HostResource: Update ##: Error Update HostResource",
			"Update:Could not Update HostResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HostResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HostResource: Update ##: Error Unmarshal HostResource",
			"Update:Could not Update HostResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HostResource: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": data})

	r.read(plan, ctx, diags)

	tflog.Debug(ctx, "HostResource: update SUCCESS ", map[string]interface{}{"plan": plan})
}

func (r *HostResource) read(state *HostResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() {
		diags.AddError(
			"HostResource: Error Read Host",
			"Update: Could not Read, Host ID is not specified.",
		)
		return
	}
	tflog.Debug(ctx, "HostResource: read ", map[string]interface{}{"state": state})

	body, err := r.client.ExecuteIPMHttpCommand("GET", "/hosts/"+state.Id.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"HostResource: read ##: Error Update HostResource",
			"Update:Could not read HostResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "HostResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HostResource: read ##: Error Unmarshal response",
			"Update:Could not read HostResource, unexpected error: "+err.Error(),
		)
		return
	}

	// populate state
	hostData := data[0].(map[string]interface{})
	state.Populate(hostData, ctx, diags)

	tflog.Debug(ctx, "HostResource: read ## ", map[string]interface{}{"plan": state})
}

func (hData *HostResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}

	tflog.Debug(ctx, "HostResourceData: populate ## ", map[string]interface{}{"computeFlag": computeFlag, "data": data})
	if computeFlag {
		hData.Id = types.StringValue(data["id"].(string))
	}
	hData.Href = types.StringValue(data["href"].(string))

	tflog.Debug(ctx, "HostResourceData: populate Config## ")
	// populate Config
	if data["config"] != nil {
		if hData.Config == nil {
			hData.Config = &HostConfig{}
		}
		hostConfig := data["config"].(map[string]interface{})
		labels := types.MapNull(types.StringType)
		if hostConfig["labels"] != nil {
			data := make(map[string]attr.Value)
			for k, v := range hostConfig["labels"].(map[string]interface{}) {
				data[k] = types.StringValue(v.(string))
			}
			labels = types.MapValueMust(types.StringType, data)
		} 
		if !hData.Config.Labels.IsNull() || computeFlag {
			hData.Config.Labels = labels
		}
		for k, v := range hostConfig {
			switch k {
			case "name": 
				if !hData.Config.Name.IsNull() || computeFlag {
					hData.Config.Name = types.StringValue(v.(string))
				}
			case "managedBy": 
				if !hData.Config.ManagedBy.IsNull() || computeFlag {
					hData.Config.ManagedBy = types.StringValue(v.(string))
				}
			case "location": 
				if !hData.Config.Location.Latitude.IsNull() || computeFlag {
					location := v.(map[string]interface{})
					hData.Config.Location.Latitude = types.Int64Value(int64(location["latitude"].(float64)))
					hData.Config.Location.Longitude = types.Int64Value(int64(location["longitude"].(float64)))
				}
			case "selector":
				hostConfigSelector := v.(map[string]interface{})
				for k1, v1 := range hostConfigSelector {
					switch k1 {
					case "moduleSelectorByModuleId":
						if hData.Config.Selector.ModuleSelectorByModuleId != nil || computeFlag {
							moduleSelectorByModuleId := v1.(map[string]interface{})
							if hData.Config.Selector.ModuleSelectorByModuleId == nil {
								hData.Config.Selector.ModuleSelectorByModuleId = &common.ModuleSelectorByModuleId{}
							}
							hData.Config.Selector.ModuleSelectorByModuleId.ModuleId = types.StringValue(moduleSelectorByModuleId["ModuleId"].(string))
						}
					case "moduleSelectorByModuleName":
						if hData.Config.Selector.ModuleSelectorByModuleName != nil || computeFlag {
							moduleSelectorByModuleName := v1.(map[string]interface{})
							if hData.Config.Selector.ModuleSelectorByModuleName == nil {
								hData.Config.Selector.ModuleSelectorByModuleName = &common.ModuleSelectorByModuleName{}
							}
							hData.Config.Selector.ModuleSelectorByModuleName.ModuleName = types.StringValue(moduleSelectorByModuleName["moduleName"].(string))
						}
					case "moduleSelectorByModuleMAC":
						if hData.Config.Selector.ModuleSelectorByModuleMAC != nil || computeFlag {
							moduleSelectorByModuleMAC := v1.(map[string]interface{})
							if hData.Config.Selector.ModuleSelectorByModuleMAC == nil {
								hData.Config.Selector.ModuleSelectorByModuleMAC = &common.ModuleSelectorByModuleMAC{}
							}
							hData.Config.Selector.ModuleSelectorByModuleMAC.ModuleMAC = types.StringValue(moduleSelectorByModuleMAC["moduleMAC"].(string))
						}
					case "moduleSelectorByModuleSerialNumber":
						if hData.Config.Selector.ModuleSelectorByModuleSerialNumber != nil || computeFlag {
							moduleSelectorByModuleSerialNumber := v1.(map[string]interface{})
							if hData.Config.Selector.ModuleSelectorByModuleSerialNumber == nil {
								hData.Config.Selector.ModuleSelectorByModuleSerialNumber = &common.ModuleSelectorByModuleSerialNumber{}
							}
							hData.Config.Selector.ModuleSelectorByModuleSerialNumber.ModuleSerialNumber = types.StringValue(moduleSelectorByModuleSerialNumber["moduleSerialNumber"].(string))
						}
					case "hostPortSelectorByChassisId":
						if hData.Config.Selector.HostSelectorByHostChassisId != nil || computeFlag {
							hostPortSelectorByChassisId := v1.(map[string]interface{})
							if hData.Config.Selector.HostSelectorByHostChassisId == nil {
								hData.Config.Selector.HostSelectorByHostChassisId = &common.HostSelectorByHostChassisId{}
							}
							hData.Config.Selector.HostSelectorByHostChassisId.ChassisId = types.StringValue(hostPortSelectorByChassisId["chassisId"].(string))
							hData.Config.Selector.HostSelectorByHostChassisId.ChassisIdSubtype = types.StringValue(hostPortSelectorByChassisId["chassisIdSubtype"].(string))
						}
					}
				}
			}
		}
	}
	tflog.Debug(ctx, "HostResourceData: populate State## ")
	// populate state
	if data["state"] != nil {
		hData.State =types.ObjectValueMust(
			HostStateAttributeType(),HostStateAttributeValue(data["state"].(map[string]interface{})))
	}

	// populate ports
	hData.Ports = types.ListNull(HostPortObjectType())
	if data["ports"] != nil && len(data["ports"].([]interface{})) > 0 {
		hData.Ports = types.ListValueMust(HostPortObjectType(), HostPortObjectsValue(data["ports"].([]interface{})))
	}
	
	tflog.Debug(ctx, "HostResourceData: populate SUCCESS ")
}

func HostSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the network module",
			Computed:    true,
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
				"location": schema.SingleNestedAttribute{
					Description: "location",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"latitude": schema.Int64Attribute{
							Description: "latitude",
							Optional:    true,
						},
						"longitude": schema.Int64Attribute{
							Description: "longitude",
							Optional:    true,
						},
					},
				},
				"labels": schema.MapAttribute{
					Description: "labels",
					Optional:    true,
					ElementType: types.StringType,
				},
				"selector": schema.SingleNestedAttribute{
					Description: "selector",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"module_selector_by_module_id": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_id",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"module_id": schema.StringAttribute{
									Description: "module_id",
									Optional:    true,
								},
							},
						},
						"module_selector_by_module_name": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_name",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"module_name": schema.StringAttribute{
									Description: "module_name",
									Optional:    true,
								},
							},
						},
						"module_selector_by_module_mac": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_mac",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"module_mac": schema.StringAttribute{
									Description: "module_mac",
									Optional:    true,
								},
							},
						},
						"module_selector_by_module_serial_number": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_serial_number",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"module_serial_number": schema.StringAttribute{
									Description: "module_serial_number",
									Optional:    true,
								},
							},
						},
						"host_selector_by_host_chassis_id": schema.SingleNestedAttribute{
							Description: "host_selector_by_host_chassis_id",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"chassis_id": schema.StringAttribute{
									Description: "chassis_id",
									Optional:    true,
								},
								"chassis_id_subtype": schema.StringAttribute{
									Description: "chassis_id_subtype",
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: HostStateAttributeType(),
		},
		//Ports      types.List `tfsdk:"ports"`
		"ports":schema.ListAttribute{
			Computed: true,
			ElementType: HostPortObjectType(),
		},
	}
}

func HostObjectType() (types.ObjectType) {
	return types.ObjectType{	
					AttrTypes: HostAttributeType(),
				}	
}

func HostObjectsValue(data []interface{}) ([]attr.Value) {
	hosts := []attr.Value{}
	for _, v := range data {
		host := v.(map[string]interface{})
		if host != nil {
			hosts = append(hosts, types.ObjectValueMust(
														HostAttributeType(),
														HostAttributeValue(host)))
		}
	}
	return hosts
}

func HostAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"href":    types.StringType, 
		"id" : types.StringType,
		"config" : types.ObjectType{AttrTypes: HostConfigAttributeType()},
		"state" : types.ObjectType{AttrTypes: HostStateAttributeType()},
		"ports" : types.ListType{ElemType:  HostPortObjectType()},
	}
}

func HostAttributeValue(host map[string]interface{}) (map[string]attr.Value) {
	href := types.StringNull()
	if host["href"] != nil {
		href = types.StringValue(host["href"].(string))
	}
	id := types.StringNull()
	if host["id"] != nil {
		id = types.StringValue(host["id"].(string))
	}
	config := types.ObjectNull(HostConfigAttributeType())
	if host["config"] != nil {
		config =  types.ObjectValueMust( HostConfigAttributeType(), HostConfigAttributeValue(host["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(HostStateAttributeType())
	if host["state"] != nil {
		state =  types.ObjectValueMust( HostStateAttributeType(), HostStateAttributeValue(host["state"].(map[string]interface{})))
	}
	ports := types.ListNull(HostPortObjectType())
	if host["ports"] != nil {
		ports =  types.ListValueMust(HostPortObjectType(),HostPortObjectsValue(host["ports"].([]interface{})))
	}
	return map[string]attr.Value{
		"href": href,
		"id": id,
		"config": config,
		"state": state,
		"ports": ports,
	}
}

func HostStateAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"name":    types.StringType,
		"chassis_id_subtype" :types.StringType,
		"chassis_id"  : types.StringType,
		"sys_name" : types.StringType,
		"sys_descr": types.StringType,
		"lldp_state": types.StringType,
		"managed_by" : types.StringType,
		"location": types.ObjectType{AttrTypes: HostLocationAttributeType()},
		"labels": types.MapType {ElemType:  types.StringType},
	}
}

func HostStateAttributeValue(hostState map[string]interface{}) (map[string]attr.Value) {
	name := types.StringNull()
	if hostState["name"] != nil {
		name = types.StringValue(hostState["name"].(string))
	}
	chassisIdSubtype := types.StringNull()
	if hostState["chassisIdSubtype"] != nil {
		chassisIdSubtype = types.StringValue(hostState["chassisIdSubtype"].(string))
	}
	chassisId := types.StringNull()
	if hostState["chassisId"] != nil {
		chassisId = types.StringValue(hostState["chassisId"].(string))
	}
	sysName := types.StringNull()
	if hostState["sysName"] != nil {
		sysName = types.StringValue(hostState["sysName"].(string))
	}
	sysDescr := types.StringNull()
	if hostState["sysDescr"] != nil {
		sysDescr = types.StringValue(hostState["sysDescr"].(string))
	}
	lldpState := types.StringNull()
	if hostState["lldpState"] != nil {
		lldpState = types.StringValue(hostState["lldpState"].(string))
	}
	managedBy := types.StringNull()
	if hostState["managedBy"] != nil {
		managedBy = types.StringValue(hostState["managedBy"].(string))
	}
	location := types.ObjectNull(HostLocationAttributeType())
	if hostState["location"] != nil {
		location = types.ObjectValueMust( HostLocationAttributeType(),
		HostLocationAttributeValue(hostState["location"].(map[string]interface{})))
	}
	labels := types.MapNull(types.StringType)
	if hostState["labels"] != nil {
		data := make(map[string]attr.Value)
		for k, v := range hostState["labels"].(map[string]interface{}) {
			data[k] = types.StringValue(v.(string))
		}
		labels = types.MapValueMust(types.StringType, data)
	}
	return map[string]attr.Value{
		"name": name,
		"chassis_id_subtype": chassisIdSubtype,
		"chassis_id": chassisId,
		"sys_name": sysName,
		"sys_descr": sysDescr,
		"managed_by": managedBy,
		"lldp_state": lldpState,
		"location": location,
		"labels": labels,
	}
}

func HostConfigAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"name":    types.StringType,
		"managed_by" : types.StringType,
		"location": types.ObjectType{AttrTypes: HostLocationAttributeType()},
		"labels": types.MapType {ElemType:  types.StringType},
		"selector": types.ObjectType{AttrTypes: HostSelectorAttributeType()},
	}
}

func HostConfigAttributeValue(hostConfig map[string]interface{}) (map[string]attr.Value) {
	name := types.StringNull()
	if hostConfig["name"] != nil {
		name = types.StringValue(hostConfig["name"].(string))
	}
	managedBy := types.StringNull()
	if hostConfig["managedBy"] != nil {
		managedBy = types.StringValue(hostConfig["managedBy"].(string))
	}
	location := types.ObjectNull(HostLocationAttributeType())
	if hostConfig["location"] != nil {
		location = types.ObjectValueMust( HostLocationAttributeType(),
		HostLocationAttributeValue(hostConfig["location"].(map[string]interface{})))
	}
	labels := types.MapNull(types.StringType)
	if hostConfig["labels"] != nil {
		data := make(map[string]attr.Value)
		for k, v := range hostConfig["labels"].(map[string]interface{}) {
			data[k] = types.StringValue(v.(string))
		}
		labels, _ = types.MapValue(types.StringType, data)
	}
	selector := types.ObjectNull(HostSelectorAttributeType())
	if hostConfig["selector"] != nil {
		selector =  types.ObjectValueMust( HostSelectorAttributeType(),
		HostSelectorAttributeValue(hostConfig["selector"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"name": name,
		"managed_by": managedBy,
		"location": location,
		"labels": labels,
		"selector": selector,
	}
}

func HostLocationAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"latitude":    types.Int64Type, 
		"longitude" : types.Int64Type,
	}
}

func HostLocationAttributeValue(location map[string]interface{}) (map[string]attr.Value) {
	latitude := types.Int64Null()
	if location["latitude"] != nil {
		latitude = types.Int64Value(int64(location["latitude"].(float64)))
	}
	longitude := types.Int64Null()
	if location["longitude"] != nil {
		longitude = types.Int64Value(int64(location["longitude"].(float64)))
	}
	return map[string]attr.Value{
		"latitude": latitude,
		"longitude": longitude,
	}
}

func HostSelectorAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
								"host_selector_by_host_chassis_id": types.ObjectType{AttrTypes: HostSelectorByChassisIdAttributeType()},
								"module_selector_by_module_id": types.ObjectType{AttrTypes: ModuleSelectorByModuleIdAttributeType()},
								"module_selector_by_module_name": types.ObjectType{AttrTypes: ModuleSelectorByModuleNameAttributeType()},
								"module_selector_by_module_mac": types.ObjectType{AttrTypes: ModuleSelectorByModuleMACAttributeType()},
								"module_selector_by_module_serial_number": types.ObjectType{AttrTypes: ModuleSelectorByModuleSerialNumberAttributeType()},
							}
}

func HostSelectorAttributeValue(selector map[string]interface{}) (map[string]attr.Value) {
	hostSelectorByChassisId := types.ObjectNull(HostSelectorByChassisIdAttributeType())
	if selector["hostSelectorByChassisId"] != nil {
		hostSelectorByChassisId = types.ObjectValueMust( HostSelectorByChassisIdAttributeType(),
		HostSelectorByChassisIdAttributeValue(selector["hostSelectorByChassisId"].(map[string]interface{})))
	}
	moduleSelectorByModuleId := types.ObjectNull(ModuleSelectorByModuleIdAttributeType())
	if selector["moduleSelectorByModuleId"] != nil {
		moduleSelectorByModuleId = types.ObjectValueMust( ModuleSelectorByModuleIdAttributeType(),
		ModuleSelectorByModuleIdAttributeValue(selector["moduleSelectorByModuleId"].(map[string]interface{})))
	}
	moduleSelectorByModuleName := types.ObjectNull(ModuleSelectorByModuleNameAttributeType())
	if selector["moduleSelectorByModuleName"] != nil {
		moduleSelectorByModuleName = types.ObjectValueMust( ModuleSelectorByModuleNameAttributeType(),
		ModuleSelectorByModuleNamedAttributeValue(selector["moduleSelectorByModuleName"].(map[string]interface{})))
	}
	moduleSelectorByModuleMAC := types.ObjectNull(ModuleSelectorByModuleMACAttributeType())
	if selector["moduleSelectorByModuleMAC"] != nil {
		moduleSelectorByModuleMAC = types.ObjectValueMust( ModuleSelectorByModuleMACAttributeType(),
		ModuleSelectorByModuleMACAttributeValue(selector["moduleSelectorByModuleMAC"].(map[string]interface{})))
	}
	moduleSelectorByModuleSerialNumber := types.ObjectNull(ModuleSelectorByModuleSerialNumberAttributeType())
	if selector["moduleSelectorByModuleSerialNumber"] != nil {
		moduleSelectorByModuleSerialNumber = types.ObjectValueMust( ModuleSelectorByModuleSerialNumberAttributeType(),
		ModuleSelectorByModuleSerialNumberAttributeValue(selector["moduleSelectorByModuleSerialNumber"].(map[string]interface{})))
	}
	
	return map[string]attr.Value{
		"host_selector_by_host_chassis_id": hostSelectorByChassisId,
		"module_selector_by_module_id": moduleSelectorByModuleId,
		"module_selector_by_module_name": moduleSelectorByModuleName,
		"module_selector_by_module_mac": moduleSelectorByModuleMAC,
		"module_selector_by_module_serial_number": moduleSelectorByModuleSerialNumber,
	}
}

func HostSelectorByChassisIdAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"chassis_id_subtype":    types.StringType, 
		"chassis_id" : types.StringType,
	}
}

func HostSelectorByChassisIdAttributeValue(hostSelectorByChassisId map[string]interface{}) (map[string]attr.Value) {
	chassisIdSubtype := types.StringNull()
	if hostSelectorByChassisId["chassisIdSubtype"] != nil {
		chassisIdSubtype = types.StringValue(hostSelectorByChassisId["chassisIdSubtype"].(string))
	}
	chassisId := types.StringNull()
	if hostSelectorByChassisId["chassisId"] != nil {
		chassisId = types.StringValue(hostSelectorByChassisId["chassisId"].(string))
	}
	return map[string]attr.Value{
		"chassis_id_subtype": chassisIdSubtype,
		"chassis_id": chassisId,
	}
}


func ModuleSelectorByModuleIdAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"module_id":    types.StringType,
	}
}

func ModuleSelectorByModuleIdAttributeValue(moduleSelectorByModuleId map[string]interface{}) (map[string]attr.Value) {
	moduleId := types.StringNull()
	if moduleSelectorByModuleId["moduleId"] != nil {
		moduleId = types.StringValue(moduleSelectorByModuleId["moduleId"].(string))
	}
	return map[string]attr.Value{
		"module_id": moduleId,
	}
}

func ModuleSelectorByModuleNameAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"module_name":    types.StringType,
	}
}

func ModuleSelectorByModuleNamedAttributeValue(moduleSelectorByModuleName map[string]interface{}) (map[string]attr.Value) {
	moduleName := types.StringNull()
	if moduleSelectorByModuleName["moduleName"] != nil {
		moduleName = types.StringValue(moduleSelectorByModuleName["moduleName"].(string))
	}
	return map[string]attr.Value{
		"module_name": moduleName,
	}
}

func ModuleSelectorByModuleMACAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"module_mac":    types.StringType,
	}
}

func ModuleSelectorByModuleMACAttributeValue(moduleSelectorByModuleMAC map[string]interface{}) (map[string]attr.Value) {
	moduleMAC := types.StringNull()
	if moduleSelectorByModuleMAC["moduleName"] != nil {
		moduleMAC = types.StringValue(moduleSelectorByModuleMAC["moduleName"].(string))
	}
	return map[string]attr.Value{
		"module_mac": moduleMAC,
	}
}

func ModuleSelectorByModuleSerialNumberAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type {
		"module_serial_number":    types.StringType,
	}
}

func ModuleSelectorByModuleSerialNumberAttributeValue(moduleSelectorByModuleSerialNumber map[string]interface{}) (map[string]attr.Value) {
	moduleMAC := types.StringNull()
	if moduleSelectorByModuleSerialNumber["moduleName"] != nil {
		moduleMAC = types.StringValue(moduleSelectorByModuleSerialNumber["moduleName"].(string))
	}
	return map[string]attr.Value{
		"module_mac": moduleMAC,
	}
}