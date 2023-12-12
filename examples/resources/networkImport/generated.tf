# __generated__ by Terraform
# Please review these resources and move them into your main configuration files.

# __generated__ by Terraform
resource "ipm_constellation_network" "constellation_networks" {
  config = {
    constellation_frequency = null
    managed_by              = null
    modulation              = null
    name                    = null
    tc_mode                 = null
    topology                = null
  }
  hub_module = {
    config = {
      managed_by = null
      module = {
        fec_iterations               = null
        fiber_connection_mode        = null
        max_dscs                     = null
        max_tx_dscs                  = null
        planned_capacity             = null
        requested_nominal_psd_offset = null
        traffic_mode                 = null
        tx_clp_target                = null
      }
      selector = {
        host_port_selector_by_name              = null
        host_port_selector_by_port_id           = null
        host_port_selector_by_port_source_mac   = null
        host_port_selector_by_sys_name          = null
        module_selector_by_module_id            = null
        module_selector_by_module_mac           = null
        module_selector_by_module_name          = null
        module_selector_by_module_serial_number = null
      }
    }
    network_id = null
  }
}
