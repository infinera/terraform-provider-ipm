terraform {
  required_providers {
    ipm = {
      source = "infinera.com/poc/ipm"
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
    device_id = "6bfbcdb7-a7ef-4ea2-6281-129cbac6d25d"
  }
  config = { 
      debug_port_access = "disabled"
  }
}

output "module" {
  value = ipm_module.module
}
