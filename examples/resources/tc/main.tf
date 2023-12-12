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

// Constellation Network Resource supports CRUD functions
resource "ipm_transport_capacity" "tc" {
  config = {
    name         = "TC4"
    capacity_mode = "portMode"
    labels       = { label1="label1" }
  }
  end_points = [
    { 
      config = {  capacity = 100 ,
                  selector = {
                      module_if_selector_by_module_name = {
                        module_client_if_aid = "XR-T1"
                        module_name = "Test_HUB1"
                      }
                  }
      }
    },
    {
      config = {  capacity = 100 ,
                  selector = {
                      module_if_selector_by_module_name = {
                        module_client_if_aid = "XR-T1"
                        module_name = "Test_LEAF1"
                      }
                  }
      }
    }
  ]
}


output "transport_capacity" {
  value = ipm_transport_capacity.tc
}