terraform {
  required_providers {
    ipm = {
      source = "infinera.com/poc/ipm"
    }
  }
}

provider "ipm" {
  username = "xr-user-1"
  password = "xr"
  host     = "https://pt-xrivk824-dv"
}

data "ipm_hub_module" "hubModule" {
  network_id = "8b31a576-3ad3-4e47-a5c8-764211f90165"
}

output "hubModule" {
  value = data.ipm_hub_module.hubModule
}
