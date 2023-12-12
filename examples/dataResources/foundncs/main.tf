terraform {
  required_providers {
    ipm = {
      source = "infinera.com/poc/ipm"
    }
  }
}

provider "ipm" {
  username = "xr-user-1"
  password = "infinera"
  host     = "ipm-eval4.westus3.cloudapp.azure.com"
}

data "ipm_found_network_connections" "ncs" {
  service_mode = "XR-L1"
  capacity = 100
  endpoint_selectors = [ { module_if_selector_by_module_name = {
                            module_client_if_aid = "XR-T1"
                            module_name = "Test_HUB1"
                          }}, { module_if_selector_by_module_name = {
                            module_client_if_aid = "XR-T1"
                            module_name = "Test_LEAF1"
                          }}
                       ]
}

output "ipm_ncs" {
  value = data.ipm_found_network_connections.ncs.ncs
}
