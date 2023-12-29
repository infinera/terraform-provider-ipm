terraform {
  required_providers {
    ipm = {
      source = "infinera/ipm"
    }
  }
}

/*provider "ipm" {
  username = "xr-user-1"
  password = "infinera"
  host     = "ipm-eval2.westus3.cloudapp.azure.com"
}*/

/*provider "ipm" {
  username = "xr-user-1"
  password = "infinera2"
  host     = "sv-xrdemo4-lt"
}*/

provider "ipm" {
  username = "xr-user-1"
  password = "infinera"
  host     = "ipm-5"
}

resource "ipm_module" "module" {
  identifier = {
    device_id = "ed4c1d56-07a4-4552-4b60-34575ee04c13"
  }
  config = { 
      debug_port_access = "disabled"
  }
}

output "module" {
  value = ipm_module.module
}
