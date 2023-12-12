package common

import (
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func IfSelectorSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "selector",
		Optional:    true,
		Attributes: IfSelectorAttributes(),
	}
}

func IfSelectorAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"module_if_selector_by_module_id": schema.SingleNestedAttribute{
			Description: "module_if_selector_by_module_id",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"module_id": schema.StringAttribute{
					Description: "module_id",
					Optional:    true,
				},
				"module_client_if_aid": schema.StringAttribute{
					Description: "module_client_if_aid",
					Optional:    true,
				},
			},
		},
		"module_if_selector_by_module_name": schema.SingleNestedAttribute{
			Description: "module_if_selector_by_module_name",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"module_name": schema.StringAttribute{
					Description: "module_name",
					Optional:    true,
				},
				"module_client_if_aid": schema.StringAttribute{
					Description: "module_client_if_aid",
					Optional:    true,
				},
			},
		},
		"module_if_selector_by_module_mac": schema.SingleNestedAttribute{
			Description: "module_if_selector_by_module_mac",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"module_mac": schema.StringAttribute{
					Description: "module_mac",
					Optional:    true,
				},
				"module_client_if_aid": schema.StringAttribute{
					Description: "module_client_if_aid",
					Optional:    true,
				},
			},
		},
		"module_if_selector_by_module_serial_number": schema.SingleNestedAttribute{
			Description: "module_if_selector_by_module_serial_number",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"module_serial_number": schema.StringAttribute{
					Description: "module_serial_number",
					Optional:    true,
				},
				"module_client_if_aid": schema.StringAttribute{
					Description: "module_client_if_aid",
					Optional:    true,
				},
			},
		},
		"host_port_selector_by_name": schema.SingleNestedAttribute{
			Description: "HostPort_port_selector_by_chassis_id",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"host_name": schema.StringAttribute{
					Description: "HostPort_name",
					Optional:    true,
				},
				"host_port_name": schema.StringAttribute{
					Description: "hostPortName",
					Optional:    true,
				},
			},
		},
		"host_port_selector_by_port_id": schema.SingleNestedAttribute{
			Description: "HostPort_port_selector_by_chassis_id",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"chassis_id_subtype": schema.StringAttribute{
					Description: "chassis_id_subtype",
					Optional:    true,
				},
				"chassis_id": schema.StringAttribute{
					Description: "chassis_id",
					Optional:    true,
				},
				"port_id_subtype": schema.StringAttribute{
					Description: "port_id_subtype",
					Optional:    true,
				},
				"port_id": schema.StringAttribute{
					Description: "port_id",
					Optional:    true,
				},
			},
		},
		"host_port_selector_by_sys_name": schema.SingleNestedAttribute{
			Description: "host_port_selector_by_sys_name",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"sys_name": schema.StringAttribute{
					Description: "sys_name",
					Optional:    true,
				},
				"port_id_subtype": schema.StringAttribute{
					Description: "port_id_subtype",
					Optional:    true,
				},
				"port_id": schema.StringAttribute{
					Description: "port_id",
					Optional:    true,
				},
			},
		},
		"host_port_selector_by_port_source_mac": schema.SingleNestedAttribute{
			Description: "host_port_selector_by_port_source_mac",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"port_source_mac": schema.StringAttribute{
					Description: "port_source_mac",
					Optional:    true,
				},
			},
		},
	}
}

func IfSelectorPopulate(selectorData map[string]interface{}, selector *IfSelector, computeOnly ...bool) {
	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}
	for k1, v1 := range selectorData {
		switch k1 {
		case "moduleIfSelectorByModuleId":
			if selector.ModuleIfSelectorByModuleId != nil || computeFlag {
				moduleIfSelectorByModuleId := v1.(map[string]interface{})
				if selector.ModuleIfSelectorByModuleId == nil {
					selector.ModuleIfSelectorByModuleId = &ModuleIfSelectorByModuleId{}
				}
				selector.ModuleIfSelectorByModuleId.ModuleId = types.StringValue(moduleIfSelectorByModuleId["ModuleId"].(string))
				selector.ModuleIfSelectorByModuleId.ModuleClientIfAid = types.StringValue(moduleIfSelectorByModuleId["moduleClientIfAid"].(string))
			}
		case "moduleIfSelectorByModuleName":
			if selector.ModuleIfSelectorByModuleName != nil || computeFlag {
				moduleIfSelectorByModuleName := v1.(map[string]interface{})
				if selector.ModuleIfSelectorByModuleName == nil {
					selector.ModuleIfSelectorByModuleName = &ModuleIfSelectorByModuleName{}
				}
				selector.ModuleIfSelectorByModuleName.ModuleName = types.StringValue(moduleIfSelectorByModuleName["moduleName"].(string))
				selector.ModuleIfSelectorByModuleName.ModuleClientIfAid = types.StringValue(moduleIfSelectorByModuleName["moduleClientIfAid"].(string))
			}
		case "moduleIfSelectorByModuleMAC":
			if selector.ModuleIfSelectorByModuleMAC != nil || computeFlag {
				moduleIfSelectorByModuleMAC := v1.(map[string]interface{})
				if selector.ModuleIfSelectorByModuleMAC == nil {
					selector.ModuleIfSelectorByModuleMAC = &ModuleIfSelectorByModuleMAC{}
				}
				selector.ModuleIfSelectorByModuleMAC.ModuleMAC = types.StringValue(moduleIfSelectorByModuleMAC["moduleMAC"].(string))
				selector.ModuleIfSelectorByModuleMAC.ModuleClientIfAid = types.StringValue(moduleIfSelectorByModuleMAC["moduleClientIfAid"].(string))
			}
		case "moduleIfSelectorByModuleSerialNumber":
			if selector.ModuleIfSelectorByModuleSerialNumber != nil || computeFlag {
				moduleIfSelectorByModuleSerialNumber := v1.(map[string]interface{})
				if selector.ModuleIfSelectorByModuleSerialNumber == nil {
					selector.ModuleIfSelectorByModuleSerialNumber = &ModuleIfSelectorByModuleSerialNumber{}
				}
				selector.ModuleIfSelectorByModuleSerialNumber.ModuleSerialNumber = types.StringValue(moduleIfSelectorByModuleSerialNumber["moduleSerialNumber"].(string))
				selector.ModuleIfSelectorByModuleSerialNumber.ModuleClientIfAid = types.StringValue(moduleIfSelectorByModuleSerialNumber["moduleClientIfAid"].(string))
			}
		case "hostPortSelectorByName":
			if selector.HostPortSelectorByName != nil || computeFlag {
				hostPortSelectorByName := v1.(map[string]interface{})
				if selector.HostPortSelectorByName == nil {
					selector.HostPortSelectorByName = &HostPortSelectorByName{}
				}
				selector.HostPortSelectorByName.HostName = types.StringValue(hostPortSelectorByName["hostName"].(string))
				selector.HostPortSelectorByName.HostPortName = types.StringValue(hostPortSelectorByName["hostPortName"].(string))
			}
		case "hostPortSelectorByPortId":
			if selector.HostPortSelectorByPortId != nil || computeFlag {
				hostPortSelectorByPortId := v1.(map[string]interface{})
				if selector.HostPortSelectorByPortId == nil {
					selector.HostPortSelectorByPortId = &HostPortSelectorByPortId{}
				}
				selector.HostPortSelectorByPortId.ChassisIdSubtype = types.StringValue(hostPortSelectorByPortId["chassisIdSubtype"].(string))
				selector.HostPortSelectorByPortId.ChassisId = types.StringValue(hostPortSelectorByPortId["chassisId"].(string))
				selector.HostPortSelectorByPortId.PortIdSubtype = types.StringValue(hostPortSelectorByPortId["portIdSubtype"].(string))
				selector.HostPortSelectorByPortId.PortId = types.StringValue(hostPortSelectorByPortId["portId"].(string))
			}
		case "hostPortSelectorBySysName":
			if selector.HostPortSelectorBySysName != nil || computeFlag {
				hostPortSelectorBySysName := v1.(map[string]interface{})
				if selector.HostPortSelectorBySysName == nil {
					selector.HostPortSelectorBySysName = &HostPortSelectorBySysName{}
				}
				selector.HostPortSelectorBySysName.SysName = types.StringValue(hostPortSelectorBySysName["sysName"].(string))
				selector.HostPortSelectorBySysName.PortIdSubtype = types.StringValue(hostPortSelectorBySysName["portIdSubtype"].(string))
				selector.HostPortSelectorBySysName.PortId = types.StringValue(hostPortSelectorBySysName["portId"].(string))
			}
		case "hostPortSelectorByPortSourceMAC":
			if selector.HostPortSelectorByPortSourceMAC != nil || computeFlag {
				hostPortSelectorByPortSourceMAC := v1.(map[string]interface{})
				if selector.HostPortSelectorByPortSourceMAC == nil {
					selector.HostPortSelectorByPortSourceMAC = &HostPortSelectorByPortSourceMAC{}
				}
				selector.HostPortSelectorByPortSourceMAC.PortSourceMAC = types.StringValue(hostPortSelectorByPortSourceMAC["portSourceMAC"].(string))
			}
		}
	}
}


type IfSelector struct {
	ModuleIfSelectorByModuleId           *ModuleIfSelectorByModuleId           `tfsdk:"module_if_selector_by_module_id"`
	ModuleIfSelectorByModuleName         *ModuleIfSelectorByModuleName         `tfsdk:"module_if_selector_by_module_name"`
	ModuleIfSelectorByModuleMAC          *ModuleIfSelectorByModuleMAC          `tfsdk:"module_if_selector_by_module_mac"`
	ModuleIfSelectorByModuleSerialNumber *ModuleIfSelectorByModuleSerialNumber `tfsdk:"module_if_selector_by_module_serial_number"`
	HostPortSelectorByName               *HostPortSelectorByName               `tfsdk:"host_port_selector_by_name"`
	HostPortSelectorByPortId             *HostPortSelectorByPortId             `tfsdk:"host_port_selector_by_port_id"`
	HostPortSelectorBySysName            *HostPortSelectorBySysName            `tfsdk:"host_port_selector_by_sys_name"`
	HostPortSelectorByPortSourceMAC      *HostPortSelectorByPortSourceMAC      `tfsdk:"host_port_selector_by_port_source_mac"`
}

type ModuleIfSelectorByModuleId struct {
	ModuleId types.String `tfsdk:"module_id"`
	ModuleClientIfAid types.String `tfsdk:"module_client_if_aid"`
}

type ModuleIfSelectorByModuleName struct {
	ModuleName types.String `tfsdk:"module_name"`
	ModuleClientIfAid types.String `tfsdk:"module_client_if_aid"`
}

type ModuleIfSelectorByModuleMAC struct {
	ModuleMAC types.String `tfsdk:"module_mac"`
	ModuleClientIfAid types.String `tfsdk:"module_client_if_aid"`
}

type ModuleIfSelectorByModuleSerialNumber struct {
	ModuleSerialNumber types.String `tfsdk:"module_serial_number"`
	ModuleClientIfAid types.String `tfsdk:"module_client_if_aid"`
}

type HostPortSelectorByName struct {
	HostName     types.String `tfsdk:"host_name"`
	HostPortName types.String `tfsdk:"host_port_name"`
}

type HostPortSelectorByPortId struct {
	ChassisIdSubtype types.String `tfsdk:"chassis_id_subtype"`
	ChassisId        types.String `tfsdk:"chassis_id"`
	PortIdSubtype    types.String `tfsdk:"port_id_subtype"`
	PortId           types.String `tfsdk:"port_id"`
}

type HostPortSelectorByChassisId struct {
	ChassisIdSubtype types.String `tfsdk:"chassis_id_subtype"`
	ChassisId        types.String `tfsdk:"chassis_id"`
}

type HostPortSelectorBySysName struct {
	SysName       types.String `tfsdk:"sys_name"`
	PortIdSubtype types.String `tfsdk:"port_id_subtype"`
	PortId        types.String `tfsdk:"port_id"`
}

type HostPortSelectorByPortSourceMAC struct {
	PortSourceMAC types.String `tfsdk:"port_source_mac"`
}

type ModuleSelectorByModuleId struct {
	ModuleId types.String `tfsdk:"module_id"`
}

type ModuleSelectorByModuleName struct {
	ModuleName types.String `tfsdk:"module_name"`
}

type ModuleSelectorByModuleMAC struct {
	ModuleMAC types.String `tfsdk:"module_mac"`
}

type ModuleSelectorByModuleSerialNumber struct {
	ModuleSerialNumber types.String `tfsdk:"module_serial_number"`
}

type HostSelectorByHostChassisId struct {
	ChassisIdSubtype types.String `tfsdk:"chassis_id_subtype"`
	ChassisId        types.String `tfsdk:"chassis_id"`
}
