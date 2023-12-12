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
  host     = "sv-xrarch-prd.infinera.com"
}

// Constellation Network Resource supports CRUD functions
resource "ipm_leaf_module" "leaf_module" {
  network_id = "f7fd3841-c649-4577-b786-b6d93e41e874"
  config = {
    selector = {
      module_selector_by_module_name = {
        module_name = "A-LRN-Leaf7"
      }
    }
    module = {
      traffic_mode : "L1Mode"
      //fiber_connection_mode : "dual"
    }
  }
}

output "leaf_module" {
  value = ipm_leaf_module.leaf_module
}



