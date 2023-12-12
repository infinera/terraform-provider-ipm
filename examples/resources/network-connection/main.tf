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

// Constellation Network Resource supports CRUD functions
resource "ipm_network_connection" "network_connection" {
  config = {
    name                    = "Network4"
    service_mode            = "XR-L1"
    implicit_transport_capacity = "none"
    labels = { label1 = "test1", label2 = "test2", label3 = "test3"}
  }
  endpoints = [ // minimum 2
    {
      config = {
        selector = {
                      module_if_selector_by_module_name = {
                        module_client_if_aid = "XR-T4"
                        module_name = "A-LRN-Hub-Orlando"
                      }
                  }
        capacity = 100
      }
    },
    {
      config = {
        selector = {
                      module_if_selector_by_module_name = {
                        module_client_if_aid = "XR-T1"
                        module_name = "A-LRN-Leaf4-MI"
                      }
                  }
        capacity = 100
      }
    }
  ]
}


output "network_connection" {
  value = ipm_network_connection.network_connection
}



