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
  host     = "ipm-eval4.westus3.cloudapp.azure.com"
}


data "ipm_networks" "networks" {
  //network_id = "c41dc6fb-b2a4-46b2-9371-71d9b26c7831"
}

output "networks" {
  value = data.ipm_networks.networks
}
