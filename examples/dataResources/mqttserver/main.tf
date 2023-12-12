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
  password = "Infinera#1"
  host     = "sv-osgams2-dt"
  // server 1
}

data "ipm_mqtt_server" "ipm_mqtt_server" {
  device_id = "7e574b35-cef9-4e53-5bb5-bc3a327e665c"
  server_id = 1
}

output "mqtt_server" {
  value = data.ipm_mqtt_server.ipm_mqtt_server
}
