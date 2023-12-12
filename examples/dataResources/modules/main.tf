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
  host     = "ipm-eval4.westus3.cloudapp.azure.com"
}*/

provider "ipm" {
  username = "xr-user-1"
  password = "Infinera#1"
  host     = "sv-osgams2-dt"
}

data "ipm_modules" "modules" {
  for_each = { for server in var.mqttServers : server.identifier.module_name != null ? server.identifier.module_name : server.identifier.module_id != null ? server.identifier.module_id : "" => server }
    name = each.value.identifier.module_name
}

output "modules" {
  value = data.ipm_modules.modules
}
