variable "transport_capacities" {
  type = list(object({ profile = optional(string),
    name = string, capacity_mode = optional(string), labels = optional(map(string)),
    end_points = list(object({ identifier = object({ module_name = optional(string), module_id = optional(string), serial_number = optional(string),
      mac_address     = optional(string), host_port_name = optional(string), host_name = optional(string),
      host_chassis_id = optional(string), host_chassis_id_subtype = optional(string),
      host_port_id    = optional(string), host_port_id_subtype = optional(string),
      host_sys_name   = optional(string), host_port_source_mac = optional(string),
      module_client_if_aid = optional(string) }),
    capacity = optional(number) }))
  }))
  description = "List of Transport Capacities"
  default = [{ name = "tc1", profile = "system_tc_profile1",
    end_points = [{ identifier = { host_chassis_id = "192.168.101.1", host_chassis_id_subtype = "ipAddress", host_port_id = "192.168.101.1",
      host_port_id_subtype = "ipAddress" } }, { identifier = { host_chassis_id = "cb3b.783c.38db", host_chassis_id_subtype = "chassisComponent",
    host_port_id = "bc3b.783c.38bd", host_port_id_subtype = "portComponent" } }]
  }]
}
