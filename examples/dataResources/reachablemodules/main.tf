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

data "ipm_reachable_modules" "reachableModules" {
  network_id = "27be828b-ea2b-48c5-ac6d-4cf705beb708"
}

output "reachableModules" {
  value = data.ipm_reachable_modules.reachableModules.modules
}
