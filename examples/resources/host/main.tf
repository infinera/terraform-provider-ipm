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

// Constellation Network Resource supports CRUD functions
resource "ipm_host" "host" {
  config = {
    name = "Berlin2"
    managed_by = "Host"
    location = { latitue = 45, longitude = 100}
    selector = {host_selector_by_host_chassis_id = { chassisId =  "192.148.10.43", chassis_id_subtype= "networkAddress"}}
    labels = {label1 : "host_label"}
  }
}

output "host" {
  value = ipm_host.host
}



