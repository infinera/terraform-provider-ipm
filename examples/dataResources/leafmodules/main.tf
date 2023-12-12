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

data "ipm_leaf_modules" "leafModules" {
  network_id = "b2f5711d-cd81-4901-91a8-809d926ba923"
}

output "leafModules" {
  value = data.ipm_leaf_modules.leafModules
  }
