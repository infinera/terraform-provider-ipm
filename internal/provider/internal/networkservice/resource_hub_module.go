package network

import (
	"context"
	"encoding/json"

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
	_ resource.Resource                = &HubModuleResource{}
	_ resource.ResourceWithConfigure   = &HubModuleResource{}
	_ resource.ResourceWithImportState = &HubModuleResource{}
)

// NewModuleResource is a helper function to simplify the provider implementation.
func NewHubModuleResource() resource.Resource {
	return &HubModuleResource{}
}

type HubModuleResource struct {
	client *ipm_pf.Client
}

type ModuleSelector struct {
	ModuleSelectorByModuleId           *common.ModuleSelectorByModuleId           `tfsdk:"module_selector_by_module_id"`
	ModuleSelectorByModuleName         *common.ModuleSelectorByModuleName         `tfsdk:"module_selector_by_module_name"`
	ModuleSelectorByModuleMAC          *common.ModuleSelectorByModuleMAC          `tfsdk:"module_selector_by_module_mac"`
	ModuleSelectorByModuleSerialNumber *common.ModuleSelectorByModuleSerialNumber `tfsdk:"module_selector_by_module_serial_number"`
	HostPortSelectorByName             *common.HostPortSelectorByName             `tfsdk:"host_port_selector_by_name"`
	HostPortSelectorByPortId           *common.HostPortSelectorByPortId           `tfsdk:"host_port_selector_by_port_id"`
	HostPortSelectorBySysName          *common.HostPortSelectorBySysName          `tfsdk:"host_port_selector_by_sys_name"`
	HostPortSelectorByPortSourceMAC    *common.HostPortSelectorByPortSourceMAC    `tfsdk:"host_port_selector_by_port_source_mac"`
}

type ConfigModule struct {
	PlannedCapacity           types.String `tfsdk:"planned_capacity"`
	TrafficMode               types.String `tfsdk:"traffic_mode"`
	FiberConnectionMode       types.String `tfsdk:"fiber_connection_mode"`
	FecIterations             types.String `tfsdk:"fec_iterations"`
	RequestedNominalPsdOffset types.String `tfsdk:"requested_nominal_psd_offset"`
	MaxDSCs                   types.Int64  `tfsdk:"max_dscs"`
	MaxTxDSCs                 types.Int64  `tfsdk:"max_tx_dscs"`
	TxCLPtarget               types.Int64  `tfsdk:"tx_clp_target"`
	/*AllowedTxCDSCs            types.List   `tfsdk:"allowed_tx_cdscs"`
	AllowedRxCDSCs            types.List   `tfsdk:"allowed_rx_cdscs"`*/
}

type NodeConfig struct {
	Selector  ModuleSelector `tfsdk:"selector"`
	Module    ConfigModule   `tfsdk:"module"`
	ManagedBy types.String   `tfsdk:"managed_by"`
}

type ModuleResourceData struct {
	NetworkId types.String `tfsdk:"network_id"`
	Id        types.String `tfsdk:"id"`
	Href      types.String `tfsdk:"href"`
	Config    *NodeConfig   `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *HubModuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hub_module"
}

// Schema defines the schema for the data source.
func (r *HubModuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type ModuleResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an Module",
		Attributes:  HubModuleSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *HubModuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r HubModuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ModuleResourceData

	diags := req.Config.Get(ctx, &data)
	tflog.Debug(ctx, "HubModuleResource: Create - ", map[string]interface{}{"ModuleResourceData": data})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r HubModuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ModuleResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "HubModuleResource: Read - ", map[string]interface{}{"ModuleResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r HubModuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ModuleResourceData

	tflog.Debug(ctx, "HubModuleResource: Update", map[string]interface{}{"UpdateRequest": req.Plan})

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "HubModuleResource: Update", map[string]interface{}{"ModuleResourceData": data})

	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r HubModuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ModuleResourceData

	diags := req.State.Get(ctx, &data)
	diags.AddError(
		"HubModuleResource: Error Delete Hub Module",
		"Delete: Could not delete Hub Module. It is deleted together with its Network",
	)
	resp.Diagnostics.Append(diags...)
}

func (r *HubModuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *HubModuleResource) update(plan *ModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "HubModuleResource: update - plan", map[string]interface{}{"plan": plan})
	if plan.NetworkId.IsNull() {
		diags.AddError(
			"HubModuleResource: Error Update Module",
			"Update: Could not Update  NetworkId is not specified. Or leaf module can not be updated",
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
			"HubModuleResource: Update ##: Error Update AC",
			"Update: Could not Marshal HubModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HubModuleResource: update - rb", map[string]interface{}{"rb": rb})

	body, err := r.client.ExecuteIPMHttpCommand("PUT", "/xr-networks/"+plan.NetworkId.ValueString()+"/hubModule", rb)

	if err != nil {
		diags.AddError(
			"HubModuleResource: Update ##: Error Update HubModuleResource",
			"Update:Could not Update HubModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HubModuleResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HubModuleResource: Update ##: Error Unmarshal HubModuleResource",
			"Update:Could not Update HubModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "HubModuleResource: Update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": data})

	r.read(plan, ctx, diags)

	tflog.Debug(ctx, "HubModuleResource: update SUCCESS ", map[string]interface{}{"plan": plan})
}

func (r *HubModuleResource) read(state *ModuleResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.NetworkId.IsNull() {
		diags.AddError(
			"HubModuleResource: Error Read Module",
			"Update: Could not read.  NetworkId is not specified.",
		)
		return
	}
	tflog.Debug(ctx, "HubModuleResource: read ", map[string]interface{}{"state": state})

	body, err := r.client.ExecuteIPMHttpCommand("GET", "/xr-networks/"+state.NetworkId.ValueString()+"/hubModule?content=expanded", nil)
	if err != nil {
		diags.AddError(
			"HubModuleResource: read ##: Error Update HubModuleResource",
			"Update:Could not read HubModuleResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "HubModuleResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HubModuleResource: read ##: Error Unmarshal response",
			"Update:Could not read HubModuleResource, unexpected error: "+err.Error(),
		)
		return
	}

	// populate state
	moduleData := data[0].(map[string]interface{})
	state.Populate(moduleData, ctx, diags)

	tflog.Debug(ctx, "HubModuleResource: read ## ", map[string]interface{}{"plan": state})
}

func (mData *ModuleResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}

	tflog.Debug(ctx, "ModuleResourceData: populate ## ", map[string]interface{}{"computeFlag": computeFlag, "data": data})
	if computeFlag {
		mData.NetworkId = types.StringValue(data["parentId"].(string))
	}
	mData.Href = types.StringValue(data["href"].(string))
	mData.Id = types.StringValue(data["id"].(string))

	tflog.Debug(ctx, "ModuleResourceData: populate Config## ")
	// populate Config
	if data["config"] != nil {
		if mData.Config == nil {
			mData.Config = &NodeConfig{}
		}
		moduleConfig := data["config"].(map[string]interface{})
		if !mData.Config.ManagedBy.IsNull() || computeFlag {
			mData.Config.ManagedBy = types.StringValue(moduleConfig["managedBy"].(string))
		}
		for k, v := range moduleConfig {
			switch k {
			case "selector":
				moduleConfigSelector := v.(map[string]interface{})
				for k1, v1 := range moduleConfigSelector {
					switch k1 {
					case "moduleSelectorByModuleId":
						if mData.Config.Selector.ModuleSelectorByModuleId != nil || computeFlag {
							moduleSelectorByModuleId := v1.(map[string]interface{})
							if mData.Config.Selector.ModuleSelectorByModuleId == nil {
								mData.Config.Selector.ModuleSelectorByModuleId = &common.ModuleSelectorByModuleId{}
							}
							mData.Config.Selector.ModuleSelectorByModuleId.ModuleId = types.StringValue(moduleSelectorByModuleId["ModuleId"].(string))
						}
					case "moduleSelectorByModuleName":
						if mData.Config.Selector.ModuleSelectorByModuleName != nil || computeFlag {
							moduleSelectorByModuleName := v1.(map[string]interface{})
							if mData.Config.Selector.ModuleSelectorByModuleName == nil {
								mData.Config.Selector.ModuleSelectorByModuleName = &common.ModuleSelectorByModuleName{}
							}
							mData.Config.Selector.ModuleSelectorByModuleName.ModuleName = types.StringValue(moduleSelectorByModuleName["moduleName"].(string))
						}
					case "moduleSelectorByModuleMAC":
						if mData.Config.Selector.ModuleSelectorByModuleMAC != nil || computeFlag {
							moduleSelectorByModuleMAC := v1.(map[string]interface{})
							if mData.Config.Selector.ModuleSelectorByModuleMAC == nil {
								mData.Config.Selector.ModuleSelectorByModuleMAC = &common.ModuleSelectorByModuleMAC{}
							}
							mData.Config.Selector.ModuleSelectorByModuleMAC.ModuleMAC = types.StringValue(moduleSelectorByModuleMAC["moduleMAC"].(string))
						}
					case "moduleSelectorByModuleSerialNumber":
						if mData.Config.Selector.ModuleSelectorByModuleSerialNumber != nil || computeFlag {
							moduleSelectorByModuleSerialNumber := v1.(map[string]interface{})
							if mData.Config.Selector.ModuleSelectorByModuleSerialNumber == nil {
								mData.Config.Selector.ModuleSelectorByModuleSerialNumber = &common.ModuleSelectorByModuleSerialNumber{}
							}
							mData.Config.Selector.ModuleSelectorByModuleSerialNumber.ModuleSerialNumber = types.StringValue(moduleSelectorByModuleSerialNumber["moduleSerialNumber"].(string))
						}
					case "hostPortSelectorByName":
						if mData.Config.Selector.HostPortSelectorByName != nil || computeFlag {
							hostPortSelectorByName := v1.(map[string]interface{})
							if mData.Config.Selector.HostPortSelectorByName == nil {
								mData.Config.Selector.HostPortSelectorByName = &common.HostPortSelectorByName{}
							}
							mData.Config.Selector.HostPortSelectorByName.HostName = types.StringValue(hostPortSelectorByName["hostName"].(string))
							mData.Config.Selector.HostPortSelectorByName.HostPortName = types.StringValue(hostPortSelectorByName["hostPortName"].(string))
						}
					case "hostPortSelectorByPortId":
						if mData.Config.Selector.HostPortSelectorByPortId != nil || computeFlag {
							hostPortSelectorByPortId := v1.(map[string]interface{})
							if mData.Config.Selector.HostPortSelectorByPortId == nil {
								mData.Config.Selector.HostPortSelectorByPortId = &common.HostPortSelectorByPortId{}
							}
							mData.Config.Selector.HostPortSelectorByPortId.ChassisId = types.StringValue(hostPortSelectorByPortId["chassisId"].(string))
							mData.Config.Selector.HostPortSelectorByPortId.ChassisIdSubtype = types.StringValue(hostPortSelectorByPortId["chassisIdSubtype"].(string))
							mData.Config.Selector.HostPortSelectorByPortId.PortId = types.StringValue(hostPortSelectorByPortId["portId"].(string))
							mData.Config.Selector.HostPortSelectorByPortId.PortIdSubtype = types.StringValue(hostPortSelectorByPortId["portIdSubtype"].(string))
						}
					case "hostPortSelectorBySysName":
						if mData.Config.Selector.HostPortSelectorBySysName != nil || computeFlag {
							hostPortSelectorBySysName := v1.(map[string]interface{})
							if mData.Config.Selector.HostPortSelectorBySysName == nil {
								mData.Config.Selector.HostPortSelectorBySysName = &common.HostPortSelectorBySysName{}
							}
							mData.Config.Selector.HostPortSelectorBySysName.SysName = types.StringValue(hostPortSelectorBySysName["sysName"].(string))
							mData.Config.Selector.HostPortSelectorBySysName.PortId = types.StringValue(hostPortSelectorBySysName["portId"].(string))
							mData.Config.Selector.HostPortSelectorBySysName.PortIdSubtype = types.StringValue(hostPortSelectorBySysName["portIdSubtype"].(string))
						}
					case "hostPortSelectorByPortSourceMAC":
						if mData.Config.Selector.HostPortSelectorByPortSourceMAC != nil || computeFlag {
							hostPortSelectorByPortSourceMAC := moduleConfigSelector["hostPortSelectorByPortSourceMAC"].(map[string]interface{})
							if mData.Config.Selector.HostPortSelectorByPortSourceMAC == nil {
								mData.Config.Selector.HostPortSelectorByPortSourceMAC = &common.HostPortSelectorByPortSourceMAC{}
							}
							mData.Config.Selector.HostPortSelectorByPortSourceMAC.PortSourceMAC = types.StringValue(hostPortSelectorByPortSourceMAC["portSourceMAC"].(string))
						}
					}
				}
			case "module":
				configModule := v.(map[string]interface{})
				for k1, v1 := range configModule {
					switch k1 {
					case "plannedCapacity":
						if !mData.Config.Module.PlannedCapacity.IsNull() || computeFlag {
							mData.Config.Module.PlannedCapacity = types.StringValue(v1.(string))
						}
					case "trafficMode":
						if !mData.Config.Module.TrafficMode.IsNull() || computeFlag {
							mData.Config.Module.TrafficMode = types.StringValue(v1.(string))
						}
					case "fecIterations":
						if !mData.Config.Module.FecIterations.IsNull() || computeFlag {
							mData.Config.Module.FecIterations = types.StringValue(v1.(string))
						}
					case "fiberConnectionMode":
						if !mData.Config.Module.FiberConnectionMode.IsNull() || computeFlag {
							mData.Config.Module.FiberConnectionMode = types.StringValue(v1.(string))
						}
					case "requestedNominalPsdOffset":
						if !mData.Config.Module.RequestedNominalPsdOffset.IsNull() || computeFlag {
							mData.Config.Module.RequestedNominalPsdOffset = types.StringValue(v1.(string))
						}
					case "maxDSCs":
						if !mData.Config.Module.MaxDSCs.IsNull() || computeFlag {
							mData.Config.Module.MaxDSCs = types.Int64Value(int64(v1.(float64)))
						}
					case "maxTxDSCs":
						if !mData.Config.Module.MaxTxDSCs.IsNull() || computeFlag {
							mData.Config.Module.MaxTxDSCs = types.Int64Value(int64(v1.(float64)))
						}
					case "txCLPtarget":
						if !mData.Config.Module.TxCLPtarget.IsNull() || computeFlag {
							mData.Config.Module.TxCLPtarget = types.Int64Value(int64(v1.(float64)))
						}
					/*case "allowedRxCDSCs":
						if !mData.Config.Module.AllowedRxCDSCs.IsNull() || computeFlag {
							allowedRxCDSCs := types.ListNull(types.Int64Type)
							if v1 != nil {
								allowedRxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(v1.([]interface{})))
							} 
							mData.Config.Module.AllowedRxCDSCs = allowedRxCDSCs
						} else {
							mData.Config.Module.AllowedRxCDSCs = types.ListNull(types.Int64Type)
						}
					case "allowedTxCDSCs":
						if !mData.Config.Module.AllowedTxCDSCs.IsNull() || computeFlag {
							allowedTxCDSCs := types.ListNull(types.Int64Type)
							if v1 != nil {
								allowedTxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(v1.([]interface{})))
							}
							mData.Config.Module.AllowedTxCDSCs = allowedTxCDSCs
						} else {
							mData.Config.Module.AllowedTxCDSCs = types.ListNull(types.Int64Type)
						}*/
					}
				}
			}
		}
	}
	tflog.Debug(ctx, "ModuleResourceData: populate State## ")
	// populate state
	if data["state"] != nil {
		mData.State = types.ObjectValueMust(
			NWModuleStateAttributeType(), NWModuleStateAttributeValue(data["state"].(map[string]interface{})))
	}
	tflog.Debug(ctx, "ModuleResourceData: populate SUCCESS ")
}

func HubModuleSchemaAttributes() map[string]schema.Attribute {
	return ModuleSchemaAttributes(false)
}

func LeafModuleSchemaAttributes() map[string]schema.Attribute {
	return ModuleSchemaAttributes(false)
}

func ComputedOnlyModuleSchemaAttributes() map[string]schema.Attribute {
	return ModuleSchemaAttributes(true)
}

func ModuleSchemaAttributes(computed bool) map[string]schema.Attribute {
	optionalFlag := !computed
	return map[string]schema.Attribute{
		"network_id": schema.StringAttribute{
			Description: "Numeric identifier of the Constellation Network.",
			Computed:    computed,
			Optional:    optionalFlag,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
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
			Computed: computed,
			Optional: optionalFlag,
			Attributes: map[string]schema.Attribute{
				"selector": schema.SingleNestedAttribute{
					Description: "selector",
					Computed:    computed,
					Optional:    optionalFlag,
					Attributes: map[string]schema.Attribute{
						"module_selector_by_module_id": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_id",
							Computed:    computed,
							Optional:    optionalFlag,
							Attributes: map[string]schema.Attribute{
								"module_id": schema.StringAttribute{
									Description: "module_id",
									Computed:    computed,
									Optional:    optionalFlag,
								},
							},
						},
						"module_selector_by_module_name": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_name",
							Computed:    computed,
							Optional:    optionalFlag,
							Attributes: map[string]schema.Attribute{
								"module_name": schema.StringAttribute{
									Description: "module_name",
									Computed:    computed,
									Optional:    optionalFlag,
								},
							},
						},
						"module_selector_by_module_mac": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_mac",
							Computed:    computed,
							Optional:    optionalFlag,
							Attributes: map[string]schema.Attribute{
								"module_mac": schema.StringAttribute{
									Description: "module_mac",
									Computed:    computed,
									Optional:    optionalFlag,
								},
							},
						},
						"module_selector_by_module_serial_number": schema.SingleNestedAttribute{
							Description: "module_selector_by_module_serial_number",
							Computed:    computed,
							Optional:    optionalFlag,
							Attributes: map[string]schema.Attribute{
								"module_serial_number": schema.StringAttribute{
									Description: "module_serial_number",
									Computed:    computed,
									Optional:    optionalFlag,
								},
							},
						},
						"host_port_selector_by_name": schema.SingleNestedAttribute{
							Description: "host_port_selector_by_name",
							Computed:    computed,
							Optional:    optionalFlag,
							Attributes: map[string]schema.Attribute{
								"host_name": schema.StringAttribute{
									Description: "host_name",
									Computed:    computed,
									Optional:    optionalFlag,
								},
								"host_port_name": schema.StringAttribute{
									Description: "host_port_name",
									Computed:    computed,
									Optional:    optionalFlag,
								},
							},
						},
						"host_port_selector_by_port_id": schema.SingleNestedAttribute{
							Description: "host_port_selector_by_port_id",
							Computed:    computed,
							Optional:    optionalFlag,
							Attributes: map[string]schema.Attribute{
								"chassis_id_subtype": schema.StringAttribute{
									Description: "chassis_id_subtype",
									Computed:    computed,
									Optional:    optionalFlag,
								},
								"chassis_id": schema.StringAttribute{
									Description: "chassis_id",
									Computed:    computed,
									Optional:    optionalFlag,
								},
								"port_id_subtype": schema.StringAttribute{
									Description: "port_id_subtype",
									Computed:    computed,
									Optional:    optionalFlag,
								},
								"port_id": schema.StringAttribute{
									Description: "port_id",
									Computed:    computed,
									Optional:    optionalFlag,
								},
							},
						},
						"host_port_selector_by_sys_name": schema.SingleNestedAttribute{
							Description: "host_port_selector_by_sys_name",
							Computed:    computed,
							Optional:    optionalFlag,
							Attributes: map[string]schema.Attribute{
								"sysname": schema.StringAttribute{
									Description: "sysname",
									Computed:    computed,
									Optional:    optionalFlag,
								},
								"port_id_subtype": schema.StringAttribute{
									Description: "port_id_subtype",
									Computed:    computed,
									Optional:    optionalFlag,
								},
								"port_id": schema.StringAttribute{
									Description: "port_id",
									Computed:    computed,
									Optional:    optionalFlag,
								},
							},
						},
						"host_port_selector_by_port_source_mac": schema.SingleNestedAttribute{
							Description: "host_port_selector_by_port_source_mac",
							Computed:    computed,
							Optional:    optionalFlag,
							Attributes: map[string]schema.Attribute{
								"port_source_mac": schema.StringAttribute{
									Description: "port_source_mac",
									Computed:    computed,
									Optional:    optionalFlag,
								},
							},
						},
					},
				},
				"module": schema.SingleNestedAttribute{
					Description: "module",
					Computed:    computed,
					Optional:    optionalFlag,
					Attributes: map[string]schema.Attribute{
						"planned_capacity": schema.StringAttribute{
							Description: "plannedCapacity",
							Computed:    computed,
							Optional:    optionalFlag,
						},
						"traffic_mode": schema.StringAttribute{
							Description: "traffic_mode",
							Computed:    computed,
							Optional:    optionalFlag,
						},
						"fiber_connection_mode": schema.StringAttribute{
							Description: "fiber_connection_mode",
							Computed:    computed,
							Optional:    optionalFlag,
						},
						"fec_iterations": schema.StringAttribute{
							Description: "fec_iterations",
							Computed:    computed,
							Optional:    optionalFlag,
						},
						"requested_nominal_psd_offset": schema.StringAttribute{
							Description: "requested_nominal_psd_offset",
							Computed:    computed,
							Optional:    optionalFlag,
						},
						"tx_clp_target": schema.Int64Attribute{
							Description: "tx_clp_target",
							Computed:    computed,
							Optional:    optionalFlag,
						},
						"max_dscs": schema.Int64Attribute{
							Description: "maxDSCs",
							Computed:    computed,
							Optional:    optionalFlag,
						},
						"max_tx_dscs": schema.Int64Attribute{
							Description: "maxTxDSCs",
							Computed:    computed,
							Optional:    optionalFlag,
						},
						/*"allowed_rx_cdscs": schema.ListAttribute{
							Description: "allowed_rx_cdscs",
							Computed:    computed,
							Optional:    optionalFlag,
							ElementType: types.Int64Type,
						},
						"allowed_tx_cdscs": schema.ListAttribute{
							Description: "allowed_tx_cdscs",
							Computed:    computed,
							Optional:    optionalFlag,
							ElementType: types.Int64Type,
						},*/
					},
				},
				"managed_by": schema.StringAttribute{
					Description: "managed_by",
					Computed:    computed,
					Optional:    optionalFlag,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: NWModuleStateAttributeType(),
		},
	}
}

func NWModuleObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: NWModuleAttributeType(),
	}
}

func NWModuleAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":     types.StringType,
		"href":   types.StringType,
		"config": types.ObjectType{AttrTypes: NWModuleConfigAttributeType()},
		"state":  types.ObjectType{AttrTypes: NWModuleStateAttributeType()},
	}
}

func NWModulesValue(data []interface{}) []attr.Value {
	modules := []attr.Value{}
	for _, v := range data {
		module := v.(map[string]interface{})
		if module != nil {
			modules = append(modules, types.ObjectValueMust(
				NWModuleAttributeType(),
				NWModuleAttributeValue(module)))
		}
	}
	return modules
}

func NWModuleStateAttributeValue(state map[string]interface{}) map[string]attr.Value {

	lifecycleState := types.StringNull()
	if state["lifecycleState"] != nil {
		lifecycleState = types.StringValue(state["lifecycleState"].(string))
	}
	managedBy := types.StringNull()
	if state["managedBy"] != nil {
		managedBy = types.StringValue(state["managedBy"].(string))
	}
	endpoints := types.ListNull(NWModuleEndpointObjectType())
	if state["endpoints"] != nil {
		endpoints = types.ListValueMust(NWModuleEndpointObjectType(), NWModuleEndpointsValue(state["endpoints"].([]interface{})))
	}
	lifecycleStateCause := types.ObjectNull(common.LifecycleStateCauseAttributeType())
	if state["lifecycleStateCause"] != nil {
		lifecycleStateCause = types.ObjectValueMust(common.LifecycleStateCauseAttributeType(),
			common.LifecycleStateCauseAttributeValue(state["lifecycleStateCause"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"lifecycle_state":       lifecycleState,
		"lifecycle_state_cause": lifecycleStateCause,
		"module":                types.ObjectValueMust(NWModuleStateModuleAttributeType(), NWModuleStateModuleAttributeValue(state["module"].(map[string]interface{}))),
		"endpoints":             endpoints,
		"managed_by":            managedBy,
	}
}

func NWModuleAttributeValue(module map[string]interface{}) map[string]attr.Value {
	id := types.StringNull()
	if module["id"] != nil {
		id = types.StringValue(module["id"].(string))
	}
	href := types.StringNull()
	if module["href"] != nil {
		href = types.StringValue(module["href"].(string))
	}
	config := types.ObjectNull(NWModuleConfigAttributeType())
	if (module["config"]) != nil {
		config = types.ObjectValueMust(NWModuleConfigAttributeType(), NWModuleConfigValue(module["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(NWModuleStateAttributeType())
	if (module["state"]) != nil {
		state = types.ObjectValueMust(NWModuleStateAttributeType(), NWModuleStateAttributeValue(module["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"id":     id,
		"href":   href,
		"config": config,
		"state":  state,
	}
}

func NWModuleConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module":     types.ObjectType{AttrTypes: NWModuleConfigModuleAttributeType()},
		"selector":   types.ObjectType{AttrTypes: common.ModuleSelectorAttributeType()},
		"managed_by": types.StringType,
	}
}

func NWModuleConfigValue(config map[string]interface{}) map[string]attr.Value {
	selector := types.ObjectNull(common.ModuleSelectorAttributeType())
	if config["selector"] != nil {
		selector = types.ObjectValueMust(common.ModuleSelectorAttributeType(), common.ModuleSelectorAttributeValue(config["selector"].(map[string]interface{})))
	}
	module := types.ObjectNull(NWModuleConfigModuleAttributeType())
	if config["module"] != nil {
		module = types.ObjectValueMust(NWModuleConfigModuleAttributeType(), NWModuleConfigModuleValue(config["module"].(map[string]interface{})))
	}
	managedBy := types.StringNull()
	if config["managedBy"] != nil {
		managedBy = types.StringValue(config["managedBy"].(string))
	}
	return map[string]attr.Value{
		"selector":   selector,
		"module":     module,
		"managed_by": managedBy,
	}

}

func NWModuleConfigModuleAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"planned_capacity":             types.StringType,
		"traffic_mode":                 types.StringType,
		"fiber_connection_mode":        types.StringType,
		"fec_iterations":               types.StringType,
		"requested_nominal_psd_offset": types.StringType,
		"tx_clp_target":                types.Int64Type,
		"max_dscs":                     types.Int64Type,
		"max_tx_dscs":                  types.Int64Type,
		/*"allowed_tx_cdscs":             types.ListType{ElemType: types.Int64Type},
		"allowed_rx_cdscs":             types.ListType{ElemType: types.Int64Type},*/
	}
}

func NWModuleConfigModuleValue(config map[string]interface{}) map[string]attr.Value {
	plannedCapacity := types.StringNull()
	if config["plannedCapacity"] != nil {
		plannedCapacity = types.StringValue(config["plannedCapacity"].(string))
	}
	trafficMode := types.StringNull()
	if config["trafficMode"] != nil {
		trafficMode = types.StringValue(config["trafficMode"].(string))
	}
	fiberConnectionMode := types.StringNull()
	if config["fiberConnectionMode"] != nil {
		fiberConnectionMode = types.StringValue(config["fiberConnectionMode"].(string))
	}
	fecIterations := types.StringNull()
	if config["fecIterations"] != nil {
		fecIterations = types.StringValue(config["fecIterations"].(string))
	}
	requestedNominalPsdOffset := types.StringNull()
	if config["requestedNominalPsdOffset"] != nil {
		requestedNominalPsdOffset = types.StringValue(config["requestedNominalPsdOffset"].(string))
	}
	maxDSCs := types.Int64Null()
	if config["maxDSCs"] != nil {
		maxDSCs = types.Int64Value(int64(config["maxDSCs"].(float64)))
	}
	txCLPtarget := types.Int64Null()
	if config["txCLPtarget"] != nil {
		txCLPtarget = types.Int64Value(int64(config["txCLPtarget"].(float64)))
	}
	maxTxDSCs := types.Int64Null()
	if config["maxTxDSCs"] != nil {
		maxTxDSCs = types.Int64Value(int64(config["maxTxDSCs"].(float64)))
	}
	/*allowedTxCDSCs := types.ListNull(types.Int64Type)
	if config["allowedTxCDSCs"] != nil {
		allowedTxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(config["allowedTxCDSCs"].([]interface{})))
	}
	allowedRxCDSCs := types.ListNull(types.Int64Type)
	if config["allowedRxCDSCs"] != nil {
		allowedRxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(config["allowedRxCDSCs"].([]interface{})))
	}*/
	return map[string]attr.Value{
		"planned_capacity":             plannedCapacity,
		"traffic_mode":                 trafficMode,
		"fiber_connection_mode":        fiberConnectionMode,
		"fec_iterations":               fecIterations,
		"requested_nominal_psd_offset": requestedNominalPsdOffset,
		"tx_clp_target":                txCLPtarget,
		"max_dscs":                     maxDSCs,
		"max_tx_dscs":                  maxTxDSCs,
		/*"allowed_tx_cdscs":             allowedTxCDSCs,
		"allowed_rx_cdscs":             allowedRxCDSCs,*/
	}
}

func NWModuleStateObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: NWModuleStateAttributeType(),
	}
}

func NWModuleStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"lifecycle_state":       types.StringType,
		"lifecycle_state_cause": types.ObjectType{AttrTypes: common.LifecycleStateCauseAttributeType()},
		"module":                types.ObjectType{AttrTypes: NWModuleStateModuleAttributeType()},
		"endpoints":             types.ListType{ElemType: NWModuleEndpointObjectType()},
		"managed_by":            types.StringType,
	}
}

func NWModuleStateModuleAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_id":                      types.StringType,
		"module_name":                    types.StringType,
		"mac_address":                    types.StringType,
		"serial_number":                  types.StringType,
		"hid":                            types.StringType,
		"hport_id":                       types.StringType,
		"configured_role":                types.StringType,
		"current_role":                   types.StringType,
		"role_status":                    types.StringType,
		"traffic_mode":                   types.StringType,
		"topology":                       types.StringType,
		"config_state":                   types.StringType,
		"tc_mode":                        types.BoolType,
		"constellation_frequency":        types.Int64Type,
		"host_frequency":                 types.Int64Type,
		"actual_constellation_frequency": types.Int64Type,
		"operating_frequency":            types.Int64Type,
		"modulation":                     types.StringType,
		"host_modulation":                types.StringType,
		"operating_modulation":           types.StringType,
		"fec_iterations":                 types.StringType,
		"host_fec_iterations":            types.StringType,
		"operating_fec_iterations":       types.StringType,
		"spectral_bandwidth":             types.Int64Type,
		"client_port_mode":               types.StringType,
		"baud_rate":                      types.Int64Type,
		"tx_clp_target":                  types.Int64Type,
		"host_tx_clp_target":             types.Int64Type,
		"actual_tx_clp_target":           types.Int64Type,
		"adv_line_ctrl":                  types.StringType,
		"max_dscs":                       types.Int64Type,
		"host_max_dscs":                  types.Int64Type,
		"operating_max_dscs":             types.Int64Type,
		/*"allowed_rx_cdscs":               types.ListType{ElemType: types.Int64Type},
		"allowed_tx_cdscs":               types.ListType{ElemType: types.Int64Type},*/
		"host_allowed_tx_cdscs":          types.ListType{ElemType: types.Int64Type},
		"actual_allowed_tx_cdscs":        types.ListType{ElemType: types.Int64Type},
		"host_allowed_rx_cdscs":          types.ListType{ElemType: types.Int64Type},
		"actual_allowed_rx_cdscs":        types.ListType{ElemType: types.Int64Type},
		"capabilities":                   types.MapType{ElemType: types.StringType},
	}
}

func NWModuleStateModuleAttributeValue(module map[string]interface{}) map[string]attr.Value {
	moduleId := types.StringNull()
	if module["moduleId"] != nil {
		moduleId = types.StringValue(module["moduleId"].(string))
	}
	moduleName := types.StringNull()
	if module["moduleName"] != nil {
		moduleName = types.StringValue(module["moduleName"].(string))
	}
	macAddress := types.StringNull()
	if module["macAddress"] != nil {
		macAddress = types.StringValue(module["macAddress"].(string))
	}
	serialNumber := types.StringNull()
	if module["serialNumber"] != nil {
		serialNumber = types.StringValue(module["serialNumber"].(string))
	}
	hId := types.StringNull()
	if module["hId"] != nil {
		hId = types.StringValue(module["hId"].(string))
	}
	hPortId := types.StringNull()
	if module["hPortId"] != nil {
		hPortId = types.StringValue(module["hPortId"].(string))
	}
	configuredRole := types.StringNull()
	if module["configuredRole"] != nil {
		configuredRole = types.StringValue(module["configuredRole"].(string))
	}
	currentRole := types.StringNull()
	if module["currentRole"] != nil {
		configuredRole = types.StringValue(module["currentRole"].(string))
	}
	roleStatus := types.StringNull()
	if module["roleStatus"] != nil {
		roleStatus = types.StringValue(module["roleStatus"].(string))
	}
	trafficMode := types.StringNull()
	if module["trafficMode"] != nil {
		trafficMode = types.StringValue(module["trafficMode"].(string))
	}
	topology := types.StringNull()
	if module["topology"] != nil {
		topology = types.StringValue(module["topology"].(string))
	}
	configState := types.StringNull()
	if module["configState"] != nil {
		configState = types.StringValue(module["configState"].(string))
	}
	tcMode := types.BoolNull()
	if module["tcMode"] != nil {
		tcMode = types.BoolValue(module["tcMode"].(bool))
	}
	constellationFrequency := types.Int64Null()
	if module["constellationFrequency"] != nil {
		constellationFrequency = types.Int64Value(int64(module["constellationFrequency"].(float64)))
	}
	hostFrequency := types.Int64Null()
	if module["hostFrequency"] != nil {
		hostFrequency = types.Int64Value(int64(module["hostFrequency"].(float64)))
	}
	actualConstellationFrequency := types.Int64Null()
	if module["actualConstellationFrequency"] != nil {
		actualConstellationFrequency = types.Int64Value(int64(module["actualConstellationFrequency"].(float64)))
	}
	operatingFrequency := types.Int64Null()
	if module["operatingFrequency"] != nil {
		operatingFrequency = types.Int64Value(int64(module["operatingFrequency"].(float64)))
	}
	modulation := types.StringNull()
	if module["modulation"] != nil {
		modulation = types.StringValue(module["modulation"].(string))
	}
	hostModulation := types.StringNull()
	if module["hostModulation"] != nil {
		hostModulation = types.StringValue(module["hostModulation"].(string))
	}
	operatingModulation := types.StringNull()
	if module["operatingModulation"] != nil {
		operatingModulation = types.StringValue(module["operatingModulation"].(string))
	}
	fecIterations := types.StringNull()
	if module["fecIterations"] != nil {
		fecIterations = types.StringValue(module["fecIterations"].(string))
	}
	hostFecIterations := types.StringNull()
	if module["hostFecIterations"] != nil {
		hostFecIterations = types.StringValue(module["hostFecIterations"].(string))
	}
	operatingFecIterations := types.StringNull()
	if module["operatingFecIterations"] != nil {
		operatingFecIterations = types.StringValue(module["operatingFecIterations"].(string))
	}
	spectralBandwidth := types.Int64Null()
	if module["spectralBandwidth"] != nil {
		spectralBandwidth = types.Int64Value(int64(module["spectralBandwidth"].(float64)))
	}
	clientPortMode := types.StringNull()
	if module["clientPortMode"] != nil {
		clientPortMode = types.StringValue(module["clientPortMode"].(string))
	}
	baudRate := types.Int64Null()
	if module["baudRate"] != nil {
		baudRate = types.Int64Value(int64(module["baudRate"].(float64)))
	}
	txCLPtarget := types.Int64Null()
	if module["txClpTarget"] != nil {
		txCLPtarget = types.Int64Value(int64(module["txCLPtarget"].(float64)))
	}
	hostTxCLPtarget := types.Int64Null()
	if module["hostTxCLPtarget"] != nil {
		hostTxCLPtarget = types.Int64Value(int64(module["hostTxCLPtarget"].(float64)))
	}
	actualTxCLPtarget := types.Int64Null()
	if module["actualTxClpTarget"] != nil {
		actualTxCLPtarget = types.Int64Value(int64(module["actualTxCLPtarget"].(float64)))
	}
	advLineCtrl := types.StringNull()
	if module["advLineCtrl"] != nil {
		advLineCtrl = types.StringValue(module["advLineCtrl"].(string))
	}
	maxDSCs := types.Int64Null()
	if module["maxDSCs"] != nil {
		maxDSCs = types.Int64Value(int64(module["maxDSCs"].(float64)))
	}
	hostMaxDSCs := types.Int64Null()
	if module["hostMaxDSCs"] != nil {
		hostMaxDSCs = types.Int64Value(int64(module["hostMaxDSCs"].(float64)))
	}
	operatingMaxDSCs := types.Int64Null()
	if module["operatingMaxDSCs"] != nil {
		operatingMaxDSCs = types.Int64Value(int64(module["operatingMaxDSCs"].(float64)))
	}
	/*allowedTxCDSCs := types.ListNull(types.Int64Type)
	if module["allowedTxCDSCs"] != nil {
		allowedTxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(module["allowedTxCDSCs"].([]interface{})))
	}
	allowedRxCDSCs := types.ListNull(types.Int64Type)
	if module["allowedRxCDSCs"] != nil {
		allowedRxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(module["allowedRxCDSCs"].([]interface{})))
	}*/
	hostAllowedTxCDSCs := types.ListNull(types.Int64Type)
	if module["hostAllowedTxCDSCs"] != nil {
		hostAllowedTxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(module["hostAllowedTxCDSCs"].([]interface{})))
	}
	actualAllowedTxCDSCs := types.ListNull(types.Int64Type)
	if module["actualAllowedTxCDSCs"] != nil {
		actualAllowedTxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(module["actualAllowedTxCDSCs"].([]interface{})))
	}
	hostAllowedRxCDSCs := types.ListNull(types.Int64Type)
	if module["hostAllowedRxCDSCs"] != nil {
		hostAllowedRxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(module["hostAllowedRxCDSCs"].([]interface{})))
	}
	actualAllowedRxCDSCs := types.ListNull(types.Int64Type)
	if module["actualAllowedRxCDSCs"] != nil {
		actualAllowedRxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(module["actualAllowedRxCDSCs"].([]interface{})))
	}
	capabilities := types.MapNull(types.StringType)
	if module["capabilities"] != nil {
		capabilities = types.MapValueMust(types.StringType, common.MapAttributeValue(module["capabilities"].(map[string]interface{})))
	}
	return map[string]attr.Value{
		"module_id":                      moduleId,
		"module_name":                    moduleName,
		"mac_address":                    macAddress,
		"serial_number":                  serialNumber,
		"hid":                            hId,
		"hport_id":                       hPortId,
		"configured_role":                configuredRole,
		"current_role":                   currentRole,
		"role_status":                    roleStatus,
		"traffic_mode":                   trafficMode,
		"topology":                       topology,
		"config_state":                   configState,
		"tc_mode":                        tcMode,
		"constellation_frequency":        constellationFrequency,
		"host_frequency":                 hostFrequency,
		"actual_constellation_frequency": actualConstellationFrequency,
		"operating_frequency":            operatingFrequency,
		"modulation":                     modulation,
		"host_modulation":                hostModulation,
		"operating_modulation":           operatingModulation,
		"fec_iterations":                 fecIterations,
		"host_fec_iterations":            hostFecIterations,
		"operating_fec_iterations":       operatingFecIterations,
		"spectral_bandwidth":             spectralBandwidth,
		"client_port_mode":               clientPortMode,
		"baud_rate":                      baudRate,
		"tx_clp_target":                  txCLPtarget,
		"host_tx_clp_target":             hostTxCLPtarget,
		"actual_tx_clp_target":           actualTxCLPtarget,
		"adv_line_ctrl":                  advLineCtrl,
		"max_dscs":                       maxDSCs,
		"host_max_dscs":                  hostMaxDSCs,
		"operating_max_dscs":             operatingMaxDSCs,
		/*"allowed_tx_cdscs":               allowedTxCDSCs,
		"allowed_rx_cdscs":               allowedRxCDSCs,*/
		"host_allowed_tx_cdscs":          hostAllowedTxCDSCs,
		"actual_allowed_tx_cdscs":        actualAllowedTxCDSCs,
		"host_allowed_rx_cdscs":          hostAllowedRxCDSCs,
		"actual_allowed_rx_cdscs":        actualAllowedRxCDSCs,
		"capabilities":                   capabilities,
	}
}

func NWModuleEndpointObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: NWModuleEndpointAttributeType(),
	}
}

func NWModuleEndpointAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"host_port": types.ObjectType{
			AttrTypes: common.EndpointHostPortAttributeType(),
		},
		"module_if": types.ObjectType{
			AttrTypes: NWModuleEndpointModuleIfAttributeType(),
		},
	}
}
func NWModuleEndpointModuleIfAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"client_if_colid":      types.Int64Type,
		"client_if_aid":        types.StringType,
		"client_if_port_speed": types.Int64Type,
	}
}


func NWModuleEndpointsValue(data []interface{}) []attr.Value {
	endpoints := []attr.Value{}
	for _, v := range data {
		endpoint := v.(map[string]interface{})
		hostPort := types.ObjectNull(common.EndpointHostPortAttributeType())
		if (endpoint["hostPort"]) != nil {
			hp := endpoint["hostPort"].(map[string]interface{})
			name := types.StringNull()
			if hp["name"] != nil {
				name = types.StringValue(hp["name"].(string))
			}
			hostName := types.StringNull()
			if hp["hostName"] != nil {
				hostName = types.StringValue(hp["hostName"].(string))
			}
			chassisId := types.StringNull()
			if hp["chassisId"] != nil {
				chassisId = types.StringValue(hp["chassisId"].(string))
			}
			chassisIdSubtype := types.StringNull()
			if hp["chassisIdSubtype"] != nil {
				chassisIdSubtype = types.StringValue(hp["chassisIdSubtype"].(string))
			}
			portId := types.StringNull()
			if hp["portId"] != nil {
				portId = types.StringValue(hp["portId"].(string))
			}
			portDescr := types.StringNull()
			if hp["portDescr"] != nil {
				portDescr = types.StringValue(hp["portDescr"].(string))
			}
			portSourceMAC := types.StringNull()
			if hp["portSourceMAC"] != nil {
				portSourceMAC = types.StringValue(hp["portSourceMAC"].(string))
			}
			portIdSubtype := types.StringNull()
			if hp["portIdSubtype"] != nil {
				portIdSubtype = types.StringValue(hp["portIdSubtype"].(string))
			}
			sysName := types.StringNull()
			if hp["sysName"] != nil {
				sysName = types.StringValue(hp["sysName"].(string))
			}
			hostPort = types.ObjectValueMust(common.EndpointHostPortAttributeType(), map[string]attr.Value{
				"name":               name,
				"host_name":          hostName,
				"chassis_id_subtype": chassisIdSubtype,
				"chassis_id":         chassisId,
				"port_id":            portId,
				"port_descr":         portDescr,
				"sys_name":           sysName,
				"port_source_mac":    portSourceMAC,
				"port_id_subtype":    portIdSubtype,
			})
		}
		moduleIf := types.ObjectNull(NWModuleEndpointModuleIfAttributeType())
		if (endpoint["moduleIf"]) != nil {
			mIf := endpoint["moduleIf"].(map[string]interface{})
			clientIfColId := types.Int64Null()
			if mIf["clientIfColId"] != nil {
				clientIfColId = types.Int64Value(int64(mIf["clientIfColId"].(float64)))
			}
			clientIfPortSpeed := types.Int64Null()
			if mIf["clientIfPortSpeed"] != nil {
				clientIfPortSpeed = types.Int64Value(int64(mIf["clientIfPortSpeed"].(float64)))
			}
			clientIfAid := types.StringNull()
			if mIf["clientIfAid"] != nil {
				clientIfAid = types.StringValue(mIf["clientIfAid"].(string))
			}
			moduleIf = types.ObjectValueMust(NWModuleEndpointModuleIfAttributeType(), map[string]attr.Value{
				"client_if_colid":      clientIfColId,
				"client_if_aid":        clientIfAid,
				"client_if_port_speed": clientIfPortSpeed,
			})
		}
		endpoints = append(endpoints, types.ObjectValueMust(
			NWModuleEndpointAttributeType(),
			map[string]attr.Value{
				"host_port": hostPort,
				"module_if": moduleIf,
			}))
	}
	return endpoints
}
