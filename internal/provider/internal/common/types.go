package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeviceIdentifier struct {
	DeviceId types.String `tfsdk:"device_id"`
	DeviceName types.String `tfsdk:"device_name"`
	DeviceSerialNumber types.String `tfsdk:"device_serial_number"`
	DeviceMACAddress types.String `tfsdk:"device_mac_address"`
}

func DeviceIdentifierAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
						Description: "Module Identifier",
						Optional : true,
						Attributes: map[string]schema.Attribute{
							"device_name": schema.StringAttribute{
								Description: "device name",
								Optional : true,
							},
							"device_id": schema.StringAttribute{
								Description: "device id",
								Optional : true,
							},
							"device_mac_address": schema.StringAttribute{
								Description: "device mac_address",
								Optional : true,
							},
							"device_serial_number": schema.StringAttribute{
								Description: "device module_serial_number",
								Optional : true,
							},
						},
					}
}


type ResourceIdentifier struct {
	DeviceId types.String  `tfsdk:"device_id"`
	Href types.String `tfsdk:"href"`
	Id types.String `tfsdk:"id"`
	Aid types.String `tfsdk:"aid"`
	GrandParentColId types.String `tfsdk:"grand_parent_col_id"`
	ParentColId types.String `tfsdk:"parent_col_id"`
	ColId types.String `tfsdk:"col_id"`
}

func ResourceIdentifierAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute {
			Description: "Identifier",
			Optional : true,
			Attributes: map[string]schema.Attribute{
				"device_id": schema.StringAttribute{
					Description: "Device ID",
					Optional : true,
				},
				"href": schema.StringAttribute{
					Description: "Resource HREF",
					Optional : true,
				},
				"aid": schema.StringAttribute{
					Description: "Resource AID",
					Optional : true,
				},
				"id": schema.StringAttribute{
					Description: "Resource id",
					Optional : true,
				},
				"grand_parent_col_id": schema.StringAttribute{
					Description: "Resource Grand Parent col ID",
					Optional : true,
				},
				"parent_col_id": schema.StringAttribute{
					Description: "Resource Parent col ID",
					Optional : true,
				},
				"col_id": schema.StringAttribute{
					Description: "Resource col ID",
					Optional : true,
				},
			},
		}
}


func ModuleIfObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: ModuleIfAttributeType(),
	}
}

func ModuleIfAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_id":            types.StringType,
		"module_name":          types.StringType,
		"serial_number":        types.StringType,
		"current_role":         types.StringType,
		"client_if_col_id":      types.Int64Type,
		"client_if_aid":        types.StringType,
		"client_if_port_speed": types.Int64Type,
	}
}

func ModuleIfAttributeValue(moduleIf map[string]interface{}) map[string]attr.Value {
	moduleId := types.StringNull()
	if moduleIf["moduleId"] != nil {
		moduleId = types.StringValue(moduleIf["moduleId"].(string))
	}
	moduleName := types.StringNull()
	if moduleIf["moduleName"] != nil {
		moduleName = types.StringValue(moduleIf["moduleName"].(string))
	}
	serialNumber := types.StringNull()
	if moduleIf["serialNumber"] != nil {
		serialNumber = types.StringValue(moduleIf["serialNumber"].(string))
	}
	currentRole := types.StringNull()
	if moduleIf["currentRole"] != nil {
		currentRole = types.StringValue(moduleIf["currentRole"].(string))
	}
	clientIfColId := types.Int64Null()
	if moduleIf["clientIfColId"] != nil {
		clientIfColId = types.Int64Value(int64(moduleIf["clientIfColId"].(float64)))
	}
	clientIfPortSpeed := types.Int64Null()
	if moduleIf["clientIfPortSpeed"] != nil {
		clientIfPortSpeed = types.Int64Value(int64(moduleIf["clientIfPortSpeed"].(float64)))
	}
	clientIfAid := types.StringNull()
	if moduleIf["clientIfAid"] != nil {
		clientIfAid = types.StringValue(moduleIf["clientIfAid"].(string))
	}
	return map[string]attr.Value{
		"module_id":            moduleId,
		"module_name":          moduleName,
		"serial_number":        serialNumber,
		"current_role":         currentRole,
		"client_if_col_id":     clientIfColId,
		"client_if_aid":        clientIfAid,
		"client_if_port_speed": clientIfPortSpeed,
	}
}

func IfSelectorAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_if_selector_by_module_id":            types.ObjectType{AttrTypes: ModuleIfSelectorByModuleIdAttributeType()},
		"module_if_selector_by_module_name":          types.ObjectType{AttrTypes: ModuleIfSelectorByModuleNameAttributeType()},
		"module_if_selector_by_module_mac":           types.ObjectType{AttrTypes: ModuleIfSelectorByModuleMACAttributeType()},
		"module_if_selector_by_module_serial_number": types.ObjectType{AttrTypes: ModuleIfSelectorByModuleSerialNumberAttributeType()},
		"host_port_selector_by_name":                 types.ObjectType{AttrTypes: HostPortSelectorByNameAttributeType()},
		"host_port_selector_by_port_id":              types.ObjectType{AttrTypes: HostPortSelectorByPortIdAttributeType()},
		"host_port_selector_by_sys_name":             types.ObjectType{AttrTypes: HostPortSelectorBySysNameAttributeType()},
		"host_port_selector_by_port_source_mac":      types.ObjectType{AttrTypes: HostPortSelectorByPortSourceMACAttributeType()},
	}
}

func IfSelectorAttributeValue(selector map[string]interface{}) map[string]attr.Value {
	moduleIfSelectorByModuleId := types.ObjectNull(ModuleIfSelectorByModuleIdAttributeType())
	if selector["moduleIfSelectorByModuleId"] != nil {
		aSelector := selector["moduleIfSelectorByModuleId"].(map[string]interface{})
		moduleIfSelectorByModuleId = types.ObjectValueMust(ModuleIfSelectorByModuleIdAttributeType(), ModuleIfSelectorByModuleIdAttributeValue(aSelector))
	}
	moduleIfSelectorByModuleName := types.ObjectNull(ModuleIfSelectorByModuleNameAttributeType())
	if selector["moduleIfSelectorByModuleName"] != nil {
		aSelector := selector["moduleIfSelectorByModuleName"].(map[string]interface{})
		moduleIfSelectorByModuleName = types.ObjectValueMust(ModuleIfSelectorByModuleNameAttributeType(), ModuleIfSelectorByModuleNameAttributeValue(aSelector))
	}
	moduleIfSelectorByModuleMAC := types.ObjectNull(ModuleIfSelectorByModuleMACAttributeType())
	if selector["moduleIfSelectorByModuleMAC"] != nil {
		aSelector := selector["moduleIfSelectorByModuleMAC"].(map[string]interface{})
		moduleIfSelectorByModuleMAC = types.ObjectValueMust(ModuleIfSelectorByModuleMACAttributeType(), ModuleIfSelectorByModuleMACAttributeValue(aSelector))
	}
	moduleIfSelectorByModuleSerialNumber := types.ObjectNull(ModuleIfSelectorByModuleSerialNumberAttributeType())
	if selector["moduleIfSelectorByModuleSerialNumber"] != nil {
		aSelector := selector["moduleIfSelectorByModuleSerialNumber"].(map[string]interface{})
		moduleIfSelectorByModuleSerialNumber = types.ObjectValueMust(ModuleIfSelectorByModuleSerialNumberAttributeType(), ModuleIfSelectorByModuleSerialNumberAttributeValue(aSelector))
	}
	hostPortSelectorByName := types.ObjectNull(HostPortSelectorByNameAttributeType())
	if selector["hostPortSelectorByName"] != nil {
		aSelector := selector["hostPortSelectorByName"].(map[string]interface{})
		hostPortSelectorByName = types.ObjectValueMust(HostPortSelectorByNameAttributeType(), HostPortSelectorByNameAttributeValue(aSelector))
	}
	hostPortSelectorByPortSourceMAC := types.ObjectNull(HostPortSelectorByPortSourceMACAttributeType())
	if selector["hostPortSelectorByPortSourceMAC"] != nil {
		aSelector := selector["hostPortSelectorByPortSourceMAC"].(map[string]interface{})
		hostPortSelectorByPortSourceMAC = types.ObjectValueMust(HostPortSelectorByPortSourceMACAttributeType(), HostPortSelectorByPortSourceMACAttributeValue(aSelector))
	}
	hostPortSelectorByPortId := types.ObjectNull(HostPortSelectorByPortIdAttributeType())
	if selector["hostPortSelectorByPortId"] != nil {
		aSelector := selector["hostPortSelectorByPortId"].(map[string]interface{})
		hostPortSelectorByPortId = types.ObjectValueMust(HostPortSelectorByPortIdAttributeType(), HostPortSelectorByPortIdAttributeValue(aSelector))
	}
	hostPortSelectorBySysName := types.ObjectNull(HostPortSelectorBySysNameAttributeType())
	if selector["hostPortSelectorBySysName"] != nil {
		aSelector := selector["hostPortSelectorBySysName"].(map[string]interface{})
		hostPortSelectorBySysName = types.ObjectValueMust(HostPortSelectorBySysNameAttributeType(), HostPortSelectorBySysNameAttributeValue(aSelector))
	}

	return map[string]attr.Value{
		"module_if_selector_by_module_id":            moduleIfSelectorByModuleId,
		"module_if_selector_by_module_name":          moduleIfSelectorByModuleName,
		"module_if_selector_by_module_mac":           moduleIfSelectorByModuleMAC,
		"module_if_selector_by_module_serial_number": moduleIfSelectorByModuleSerialNumber,
		"host_port_selector_by_name":              hostPortSelectorByName,
		"host_port_selector_by_port_id":           hostPortSelectorByPortId,
		"host_port_selector_by_sys_name":          hostPortSelectorBySysName,
		"host_port_selector_by_port_source_mac":   hostPortSelectorByPortSourceMAC,
	}
}

func ModuleIfSelectorByModuleIdAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_id":            types.StringType,
		"module_client_if_aid": types.StringType,
	}
}
func ModuleIfSelectorByModuleIdAttributeValue(moduleIfSelectorByModuleId map[string]interface{}) map[string]attr.Value {
	moduleId := types.StringNull()
	if moduleIfSelectorByModuleId != nil && moduleIfSelectorByModuleId["moduleId"] != nil {
		moduleId = types.StringValue(moduleIfSelectorByModuleId["moduleId"].(string))
	}
	moduleClientIfAid := types.StringNull()
	if moduleIfSelectorByModuleId != nil && moduleIfSelectorByModuleId["moduleClientIfAid"] != nil {
		moduleClientIfAid = types.StringValue(moduleIfSelectorByModuleId["moduleClientIfAid"].(string))
	}
	return map[string]attr.Value{
		"module_id":            moduleId,
		"module_client_if_aid": moduleClientIfAid,
	}
}
func ModuleIfSelectorByModuleSerialNumberAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_serial_number":           types.StringType,
		"module_client_if_aid": types.StringType,
	}
}
func ModuleIfSelectorByModuleSerialNumberAttributeValue(moduleIfSelectorByModuleSerialNumber map[string]interface{}) map[string]attr.Value {
	moduleSerialNumber := types.StringNull()
	if moduleIfSelectorByModuleSerialNumber != nil && moduleIfSelectorByModuleSerialNumber["moduleSerialNumber"] != nil {
		moduleSerialNumber = types.StringValue(moduleIfSelectorByModuleSerialNumber["moduleSerialNumber"].(string))
	}
	moduleClientIfAid := types.StringNull()
	if moduleIfSelectorByModuleSerialNumber != nil && moduleIfSelectorByModuleSerialNumber["moduleClientIfAid"] != nil {
		moduleClientIfAid = types.StringValue(moduleIfSelectorByModuleSerialNumber["moduleClientIfAid"].(string))
	}
	return map[string]attr.Value{
		"module_serial_number":            moduleSerialNumber,
		"module_client_if_aid": moduleClientIfAid,
	}
}

func ModuleIfSelectorByModuleMACAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_mac":           types.StringType,
		"module_client_if_aid": types.StringType,
	}
}
func ModuleIfSelectorByModuleMACAttributeValue(moduleIfSelectorByModuleMAC map[string]interface{}) map[string]attr.Value {
	moduleMAC := types.StringNull()
	if moduleIfSelectorByModuleMAC != nil && moduleIfSelectorByModuleMAC["moduleMAC"] != nil {
		moduleMAC = types.StringValue(moduleIfSelectorByModuleMAC["moduleMAC"].(string))
	}
	moduleClientIfAid := types.StringNull()
	if moduleIfSelectorByModuleMAC != nil && moduleIfSelectorByModuleMAC["moduleClientIfAid"] != nil {
		moduleClientIfAid = types.StringValue(moduleIfSelectorByModuleMAC["moduleClientIfAid"].(string))
	}
	return map[string]attr.Value{
		"module_mac":            moduleMAC,
		"module_client_if_aid": moduleClientIfAid,
	}
}

func ModuleIfSelectorByModuleNameAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_name":          types.StringType,
		"module_client_if_aid": types.StringType,
	}
}
func ModuleIfSelectorByModuleNameAttributeValue(moduleIfSelectorByModuleName map[string]interface{}) map[string]attr.Value {
	moduleName := types.StringNull()
	if moduleIfSelectorByModuleName != nil && moduleIfSelectorByModuleName["moduleName"] != nil {
		moduleName = types.StringValue(moduleIfSelectorByModuleName["moduleName"].(string))
	}
	moduleClientIfAid := types.StringNull()
	if moduleIfSelectorByModuleName != nil && moduleIfSelectorByModuleName["moduleClientIfAid"] != nil {
		moduleClientIfAid = types.StringValue(moduleIfSelectorByModuleName["moduleClientIfAid"].(string))
	}
	return map[string]attr.Value{
		"module_name":            moduleName,
		"module_client_if_aid": moduleClientIfAid,
	}
}

func HostPortSelectorByPortSourceMACAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"port_source_mac": types.StringType,
	}
}
func HostPortSelectorByPortSourceMACAttributeValue(hostPortSelectorByPortSourceMAC map[string]interface{}) map[string]attr.Value {
	portSourceMAC := types.StringNull()
	if hostPortSelectorByPortSourceMAC != nil && hostPortSelectorByPortSourceMAC["portSourceMAC"] != nil {
		portSourceMAC = types.StringValue(hostPortSelectorByPortSourceMAC["portSourceMAC"].(string))
	}
	return map[string]attr.Value{
		"port_source_mac": portSourceMAC,
	}
}

func HostPortSelectorByNameAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"host_name":      types.StringType,
		"host_port_name": types.StringType,
	}
}
func HostPortSelectorByNameAttributeValue(hostPortSelectorByName map[string]interface{}) map[string]attr.Value {
	hostName := types.StringNull()
	if hostPortSelectorByName != nil && hostPortSelectorByName["hostName"] != nil {
		hostName = types.StringValue(hostPortSelectorByName["hostName"].(string))
	}
	hostPortName := types.StringNull()
	if hostPortSelectorByName != nil && hostPortSelectorByName["hostPortName"] != nil {
		hostPortName = types.StringValue(hostPortSelectorByName["hostPortName"].(string))
	}
	return map[string]attr.Value{
		"host_name":      hostName,
		"host_port_name": hostPortName,
	}
}

func HostPortSelectorByPortIdAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"chassis_id_subtype": types.StringType,
		"chassis_id":         types.StringType,
		"port_id_subtype":    types.StringType,
		"port_id":            types.StringType,
	}
}
func HostPortSelectorByPortIdAttributeValue(hostPortSelectorByPortId map[string]interface{}) map[string]attr.Value {
	chassisIdSubtype := types.StringNull()
	if hostPortSelectorByPortId != nil && hostPortSelectorByPortId["chassisIdSubtype"] != nil {
		chassisIdSubtype = types.StringValue(hostPortSelectorByPortId["chassisIdSubtype"].(string))
	}
	chassisId := types.StringNull()
	if hostPortSelectorByPortId != nil && hostPortSelectorByPortId["chassisId"] != nil {
		chassisId = types.StringValue(hostPortSelectorByPortId["chassisId"].(string))
	}
	portIdSubtype := types.StringNull()
	if hostPortSelectorByPortId != nil && hostPortSelectorByPortId["portIdSubtype"] != nil {
		portIdSubtype = types.StringValue(hostPortSelectorByPortId["portIdSubtype"].(string))
	}
	portId := types.StringNull()
	if hostPortSelectorByPortId != nil && hostPortSelectorByPortId["portId"] != nil {
		portId = types.StringValue(hostPortSelectorByPortId["portId"].(string))
	}
	return map[string]attr.Value{
		"chassis_id_subtype": chassisIdSubtype,
		"chassis_id":         chassisId,
		"port_id_subtype":    portIdSubtype,
		"port_id":            portId,
	}
}

func HostPortSelectorBySysNameAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"sys_name":        types.StringType,
		"port_id_subtype": types.StringType,
		"port_id":         types.StringType,
	}
}
func HostPortSelectorBySysNameAttributeValue(hostPortSelectorBySysName map[string]interface{}) map[string]attr.Value {
	sysName := types.StringNull()
	if hostPortSelectorBySysName != nil && hostPortSelectorBySysName["sysName"] != nil {
		sysName = types.StringValue(hostPortSelectorBySysName["sysName"].(string))
	}
	portIdSubtype := types.StringNull()
	if hostPortSelectorBySysName != nil && hostPortSelectorBySysName["portIdSubtype"] != nil {
		portIdSubtype = types.StringValue(hostPortSelectorBySysName["portIdSubtype"].(string))
	}
	portId := types.StringNull()
	if hostPortSelectorBySysName != nil && hostPortSelectorBySysName["portId"] != nil {
		portId = types.StringValue(hostPortSelectorBySysName["portId"].(string))
	}
	return map[string]attr.Value{
		"sys_name":        sysName,
		"port_id_subtype": portIdSubtype,
		"port_id":         portId,
	}
}

func ModuleSelectorAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"module_selector_by_module_id":            types.ObjectType{AttrTypes: map[string]attr.Type{"module_id": types.StringType}},
		"module_selector_by_module_name":          types.ObjectType{AttrTypes: map[string]attr.Type{"module_name": types.StringType}},
		"module_selector_by_module_mac":           types.ObjectType{AttrTypes: map[string]attr.Type{"module_mac": types.StringType}},
		"module_selector_by_module_serial_number": types.ObjectType{AttrTypes: map[string]attr.Type{"module_serial_number": types.StringType}},
		"host_port_selector_by_name":              types.ObjectType{AttrTypes: map[string]attr.Type{"host_name": types.StringType, "host_port_name": types.StringType}},
		"host_port_selector_by_port_id":           types.ObjectType{AttrTypes: map[string]attr.Type{"chassis_id_subtype": types.StringType, "chassis_id": types.StringType, "port_id_subtype": types.StringType, "port_id": types.StringType}},
		"host_port_selector_by_sys_name":          types.ObjectType{AttrTypes: map[string]attr.Type{"sys_name": types.StringType, "port_id_subtype": types.StringType, "port_id": types.StringType}},
		"host_port_selector_by_port_source_mac":   types.ObjectType{AttrTypes: map[string]attr.Type{"port_source_mac": types.StringType}},
	}
}

func ModuleSelectorAttributeValue(selector map[string]interface{}) map[string]attr.Value {
	moduleSelectorByModuleId := types.ObjectNull(map[string]attr.Type{"module_id": types.StringType})
	if selector["moduleSelectorByModuleId"] != nil {
		aSelector := selector["moduleSelectorByModuleId"].(map[string]interface{})
		moduleSelectorByModuleId = types.ObjectValueMust(map[string]attr.Type{"module_id": types.StringType}, map[string]attr.Value{
			"module_id": types.StringValue(aSelector["moduleId"].(string))})
	}
	moduleSelectorByModuleName := types.ObjectNull(map[string]attr.Type{"module_name": types.StringType})
	if selector["moduleSelectorByModuleName"] != nil {
		aSelector := selector["moduleSelectorByModuleName"].(map[string]interface{})
		moduleSelectorByModuleName = types.ObjectValueMust(map[string]attr.Type{"module_name": types.StringType}, map[string]attr.Value{
			"module_name": types.StringValue(aSelector["moduleName"].(string))})
	}
	moduleSelectorByModuleMAC := types.ObjectNull(map[string]attr.Type{"module_mac": types.StringType})
	if selector["moduleSelectorByModuleMAC"] != nil {
		aSelector := selector["moduleSelectorByModuleMAC"].(map[string]interface{})
		moduleSelectorByModuleMAC = types.ObjectValueMust(map[string]attr.Type{"module_mac": types.StringType}, map[string]attr.Value{
			"module_mac": types.StringValue(aSelector["moduleMAC"].(string))})
	}
	moduleSelectorByModuleSerialNumber := types.ObjectNull(map[string]attr.Type{"module_serial_number": types.StringType})
	if selector["moduleSelectorByModuleSerialNumber"] != nil {
		aSelector := selector["moduleSelectorByModuleSerialNumber"].(map[string]interface{})
		moduleSelectorByModuleSerialNumber = types.ObjectValueMust(map[string]attr.Type{"module_serial_number": types.StringType}, map[string]attr.Value{
			"module_serial_number": types.StringValue(aSelector["moduleSerialNumber"].(string))})
	}
	hostPortSelectorByName := types.ObjectNull(map[string]attr.Type{"host_name": types.StringType, "host_port_name": types.StringType})
	if selector["hostPortSelectorByName"] != nil {
		aSelector := selector["hostPortSelectorByName"].(map[string]interface{})
		hostPortSelectorByName = types.ObjectValueMust(map[string]attr.Type{"host_name": types.StringType, "host_port_name": types.StringType}, map[string]attr.Value{
			"hostName": types.StringValue(aSelector["hostPortName"].(string))})
	}
	hostPortSelectorByPortSourceMAC := types.ObjectNull(map[string]attr.Type{"port_source_mac": types.StringType})
	if selector["hostPortSelectorByPortSourceMAC"] != nil {
		aSelector := selector["hostPortSelectorByPortSourceMAC"].(map[string]interface{})
		hostPortSelectorByPortSourceMAC = types.ObjectValueMust(map[string]attr.Type{"port_source_mac": types.StringType}, map[string]attr.Value{"portSourceMAC": types.StringValue(aSelector["portSourceMAC"].(string))})
	}
	hostPortSelectorByPortId := types.ObjectNull(map[string]attr.Type{"chassis_id": types.StringType, "chassis_id_subtype": types.StringType, "port_id": types.StringType, "port_id_subtype": types.StringType})
	if selector["hostPortSelectorByPortId"] != nil {
		aSelector := selector["hostPortSelectorByPortId"].(map[string]interface{})
		hostPortSelectorByPortId = types.ObjectValueMust(map[string]attr.Type{"chassis_id": types.StringType, "chassis_id_subtype": types.StringType, "port_id": types.StringType, "port_id_subtype": types.StringType}, map[string]attr.Value{"chassis_id": types.StringValue(aSelector["chassisId"].(string)), "chassis_id_subtype": types.StringValue(aSelector["chassisIdSubtype"].(string)), "port_id": types.StringValue(aSelector["portId"].(string)), "port_id_subtype": types.StringValue(aSelector["portIdSubtype"].(string))})
	}
	hostPortSelectorBySysName := types.ObjectNull(map[string]attr.Type{"sys_name": types.StringType, "port_id": types.StringType, "port_id_subtype": types.StringType})
	if selector["hostPortSelectorBySysName"] != nil {
		aSelector := selector["hostPortSelectorBySysName"].(map[string]interface{})
		hostPortSelectorBySysName = types.ObjectValueMust(map[string]attr.Type{"sys_name": types.StringType, "port_id": types.StringType, "port_id_subtype": types.StringType}, map[string]attr.Value{"sys_name": types.StringValue(aSelector["sysName"].(string)), "port_id": types.StringValue(aSelector["portId"].(string)), "port_id_subtype": types.StringValue(aSelector["portIdSubtype"].(string))})
	}

	return map[string]attr.Value{
		"module_selector_by_module_id":            moduleSelectorByModuleId,
		"module_selector_by_module_name":          moduleSelectorByModuleName,
		"module_selector_by_module_mac":           moduleSelectorByModuleMAC,
		"module_selector_by_module_serial_number": moduleSelectorByModuleSerialNumber,
		"host_port_selector_by_name":              hostPortSelectorByName,
		"host_port_selector_by_port_id":           hostPortSelectorByPortId,
		"host_port_selector_by_sys_name":          hostPortSelectorBySysName,
		"host_port_selector_by_port_source_mac":   hostPortSelectorByPortSourceMAC,
	}
}

func LifecycleStateCauseAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"action":    types.Int64Type,
		"timestamp": types.StringType,
		"trace_id":  types.StringType,
		"errors": types.ListType{
			ElemType: LifecycleStateCauseErrorObjectType(),
		},
	}
}

func LifecycleStateCauseAttributeValue(lifecycleStateCause map[string]interface{}) map[string]attr.Value {
	actionValue := types.Int64Null()
	if lifecycleStateCause != nil && lifecycleStateCause["action"] != nil {
		actionValue = types.Int64Value(int64(lifecycleStateCause["action"].(float64)))
	}
	timestampValue := types.StringNull()
	if lifecycleStateCause != nil && lifecycleStateCause["timestamp"] != nil {
		timestampValue = types.StringValue(lifecycleStateCause["timestamp"].(string))
	}
	trace_id := types.StringNull()
	if lifecycleStateCause != nil && lifecycleStateCause["trace_id"] != nil {
		trace_id = types.StringValue(lifecycleStateCause["timestamp"].(string))
	}
	errors := []attr.Value{}
	if lifecycleStateCause != nil && lifecycleStateCause["errors"] != nil {
		errors = LifecycleStateCauseErrorsValue(lifecycleStateCause["errors"].([]interface{}))
	}

	return map[string]attr.Value{
		"action":    actionValue,
		"timestamp": timestampValue,
		"trace_id":  trace_id,
		"errors": types.ListValueMust(
			LifecycleStateCauseErrorObjectType(),
			errors,
		),
	}
}

func LifecycleStateCauseErrorsValue(errors []interface{}) []attr.Value {
	values := []attr.Value{}
	for _, v := range errors {
		error := v.(map[string]interface{})
		values = append(values, types.ObjectValueMust(
			LifecycleStateCauseErrorAttributeType(),
			map[string]attr.Value{
				"code":    types.StringValue(error["code"].(string)),
				"message": types.StringValue(error["message"].(string)),
			}))
	}
	return values
}

func LifecycleStateCauseErrorObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: LifecycleStateCauseErrorAttributeType(),
	}
}

func LifecycleStateCauseErrorAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"code":    types.StringType,
		"message": types.StringType,
	}
}

func EndpointHostPortObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: EndpointHostPortAttributeType(),
	}
}

func EndpointHostPortAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"host_name":          types.StringType,
		"chassis_id_subtype": types.StringType,
		"chassis_id":         types.StringType,
		"port_id_subtype":    types.StringType,
		"port_id":            types.StringType,
		"port_descr":         types.StringType,
		"sys_name":            types.StringType,
		"port_source_mac":    types.StringType,
	}
}

func EndpointHostPortAttributeValue(endpointHostPort map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	hostName := types.StringNull()
	chassisIdSubtype := types.StringNull()
	chassisId := types.StringNull()
	sysName := types.StringNull()
	portId  := types.StringNull()
	portDescr := types.StringNull()
	portSourceMAC := types.StringNull()
	portIdSubtype := types.StringNull()

	for k, v := range endpointHostPort {
		switch k {
			case "name":
				name = types.StringValue(v.(string))
			case "hostName":
				hostName = types.StringValue(v.(string))
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
		}
	}

	return map[string]attr.Value{
		"name":               name,
		"host_name":          hostName,
		"chassis_id_subtype": chassisIdSubtype,
		"chassis_id":         chassisId,
		"sys_name":           sysName,
		"port_id":            portId,
		"port_descr":         portDescr,
		"port_source_mac":    portSourceMAC,
		"port_id_subtype":    portIdSubtype,
	}
}


func InventoryObjectType() (types.ObjectType) {
	return types.ObjectType{	
						AttrTypes: InventoryAttributeType(),
				}
}

func InventoryObjectsValue(data []interface{}) []attr.Value {
	fans := []attr.Value{}
	for _, v := range data {
		fan := v.(map[string]interface{})
		if fan != nil {
			fans = append(fans, types.ObjectValueMust(
				InventoryAttributeType(),
				InventoryAttributeValue(fan)))
		}
	}
	return fans
}

func InventoryAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"hardware_version":  types.StringType,
		"actual_type":     types.StringType,
		"actual_subtype":     types.StringType,
		"part_number":  types.StringType,
		"serial_number":     types.StringType,
		"clei":     types.StringType,
		"vendor":     types.StringType,
		"manufacture_date":     types.StringType,
	}
}

func InventoryAttributeValue(inventory map[string]interface{}) map[string]attr.Value {
	hardwareVersion := types.StringNull()
	actualType := types.StringNull()
	actualSubType := types.StringNull()
	partNumber := types.StringNull()
	serialNumber := types.StringNull()
	clei := types.StringNull()
	vendor := types.StringNull()
	manufactureDate := types.StringNull()

	for k, v := range inventory {
		switch k {
			case "hardwareVersion":
				hardwareVersion = types.StringValue(v.(string))
			case "actualType":
				actualType = types.StringValue(v.(string))
			case "actualSubType":
				actualSubType = types.StringValue(v.(string))
			case "partNumber":
				partNumber = types.StringValue(v.(string))
			case "serialNumber":
				serialNumber = types.StringValue(v.(string))
			case "clei":
				clei = types.StringValue(v.(string))
			case "vendor":
				vendor = types.StringValue(v.(string))
			case "manufactureDate":
				manufactureDate = types.StringValue(v.(string))
		}
	}

	return map[string]attr.Value{
		"hardware_version": hardwareVersion,
		"actual_type": actualType,
		"actual_subtype": actualSubType,
		"part_number": partNumber,
		"serial_number": serialNumber,
		"clei": clei,
		"vendor": vendor,
		"manufacture_date": manufactureDate,
	}
}
