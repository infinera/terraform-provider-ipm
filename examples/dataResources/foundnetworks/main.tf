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


data "ipm_found_networks" "networks" {
  hub_selector = {
        module_selector_by_module_name = {
          module_name = "Test_HUB1"
        }
      }
}

output "networks" {
  value = data.ipm_found_networks.networks.networks
}
