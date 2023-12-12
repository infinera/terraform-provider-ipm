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

data "ipm_host_ports" "host_ports" {
  //id = "5ff66884-bf1b-4e77-8340-ec9d739c7ca8"
}

output "host_ports" {
  value = data.ipm_host_ports.host_ports
}
