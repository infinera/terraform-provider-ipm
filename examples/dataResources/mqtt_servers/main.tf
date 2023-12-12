terraform {
  required_providers {
    ipm = {
      source = "infinera.com/poc/ipm"
    }
  }
  
}

provider "ipm" {
  username = "xr-user-1"
  password = "Infinera#1"
  host     = "sv-osgams2-dt"
  // server 1
}

locals {
   servers = { for server in var.mqtt_servers : server.identifier.module_name != null ? server.identifier.module_name : server.identifier.module_id != null ? server.identifier.module_id : server.identifier.serial_number != null ? server.identifier.serial_number : server.identifier.mac_address != null ? server.identifier.mac_address : "" => server }
}

data "ipm_modules" "modules" {
  for_each = local.servers
    name = each.value.identifier.module_name != null ? each.value.identifier.module_name : null
    id = each.value.identifier.module_id != null ? each.value.identifier.module_id : null
    serial_number = each.value.identifier.module_serial_number != null ? each.value.identifier.module_serial_number : null
    mac_address = each.value.identifier.module_mac_address != null ? each.value.identifier.module_mac_address : null
}

data "ipm_mqtt_server" "mqtt_server" {
  for_each = local.servers
    //device_id =  data.ipm_modules.modules[each.key].id
    device_id = data.ipm_modules.modules[each.key].modules[0].id
    server_id =  each.value.identifier.server_id
}

output "mqtt_servers" {
  value = data.ipm_mqtt_server.mqtt_server
}




