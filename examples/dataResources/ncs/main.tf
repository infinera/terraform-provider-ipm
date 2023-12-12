terraform {
  required_providers {
    ipm = {
      source = "infinera.com/poc/ipm"
    }
  }
}

provider "ipm" {
  username = "xr-user-1"
  password = "infinera"
  host     = "ipm-eval2.westus3.cloudapp.azure.com"
}

data "ipm_network_connections" "ncs" {
  //network_id = "d0399ce9-74e6-4e3a-a362-8a8b52100315"
}

output "ipm_ncs" {
  value = data.ipm_network_connections.ncs
}
