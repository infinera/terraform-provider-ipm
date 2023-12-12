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
resource "ipm_network" "constellation_network" {
  id = "b0009783-c9b8-4a43-99c4-fe7038f82f6e"
  config = {
    name                    = "XR Network 1"
  }
  hub_module = {
    config = {
      selector = {
        module_selector_by_module_name = {
          module_name = "p2mp_HUB1"
        }
      }
      module = {
        traffic_mode : "L1Mode"
        fiber_connection_mode : "dual"
      }
    }
  }
}

output "network_output" {
  value = ipm_network.constellation_network
}



