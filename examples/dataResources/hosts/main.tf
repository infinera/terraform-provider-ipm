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

data "ipm_hosts" "hosts" {
  //id = "5ff66884-bf1b-4e77-8340-ec9d739c7ca8"
}

output "hosts" {
  value = data.ipm_hosts.hosts
}
