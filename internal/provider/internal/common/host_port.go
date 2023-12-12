package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HostPort struct {
	Name             types.String `tfsdk:"name"`
	HostName         types.String `tfsdk:"host_name"`
	ChassisIdSubtype types.String `tfsdk:"chassis_id_subtype"`
	ChassisId        types.String `tfsdk:"chassis_id"`
	PortIdSubtype    types.String `tfsdk:"port_id_subtype"`
	PortId           types.String `tfsdk:"port_id"`
	SysName          types.String `tfsdk:"sys_name"`
	PortSourceMAC    types.String `tfsdk:"port_source_mac"`
	PortDescr        types.String `tfsdk:"port_descr"`
}
