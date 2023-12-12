terraform {
  required_providers {
    ipm = {
      source = "infinera.com/poc/ipm"
    }
  }
}

/*provider "ipm" {
  username = "xr-user-1"
  password = "xr"
  host     = "https://pt-xrivk824-dv"
}*/

provider "ipm" {
  username = "xr-user-1"
  password = "infinera"
  host     = "ipm-eval4.westus3.cloudapp.azure.com"
}


data "ipm_found_transport_capacities" "tcs" {
  capacity_mode =  "portMode"
  a_selector = {
    module_if_selector_by_module_name = {
      module_client_if_aid = "XR-T1"
      module_name = "Test_HUB1"
    }
  }
  z_selector = {
    module_if_selector_by_module_name = {
      module_client_if_aid = "XR-T1"
      module_name = "Test_LEAF1"
    }
  }

}

output "tcs" {
  value = data.ipm_found_transport_capacities.tcs.transport_capacities
}
