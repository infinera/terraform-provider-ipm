terraform {
  required_providers {
    ipm = {
      source = "infinera.com/poc/ipm"
    }
  }
}

/*provider "ipm" {
  username = "xr-user-1"
  password = "xr"
  host     = "https://pt-xrivk824-dv"
}*/

provider "ipm" {
  username = "xr-user-1"
  password = "infinera"
  host     = "ipm-eval4.westus3.cloudapp.azure.com"
}


data "ipm_transport_capacities" "tcs" {
  //id = var.id
}

output "hosts" {
  value = data.ipm_transport_capacities.tcs
}
