variable network_profile_specication {
  type = map(object({network_profile = map(object({network_config_profile = optional(string), hub_config_profile= optional(string), leaf_config_profile= optional(string)})), 
                    network_config_profile = map(object({constellation_frequency= optional(number), modulation = optional(string), managed_by=optional(string), tc_mode=optional(string)})), 
                    module_config_profile = map(object({traffic_mode= optional(string),fiber_connection_mode= optional(string), managed_by= optional(string), planned_capacity= optional(string), requested_nominal_psd_offset= optional(string), fec_iterations= optional(string), tx_clp_target= optional(string)}))}))
  description = "Network profiles"
}

variable network_connection_profile_specification {
  type = map(object({ service_mode = optional(string), mc = optional(string), outer_vid = optional(string),
                      implicit_transport_capacity = optional(string), labels = optional(map(string)), 
                      endpoint_capacity : optional(string) }))
   description = "NC Config profiles"
}

variable "transport_capacity_profile_specification" {
  type = map(object({ capacity_mode = optional(string), labels = optional(map(string)), endpoint_capacity : optional(string)}))
  description = "Map of tc profiles"
}