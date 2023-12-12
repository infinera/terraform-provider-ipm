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
resource "ipm_hubmodule" "hub_module" {
  network_id = "61eb703e-0cbb-4a98-96cd-1be4e2506120"
  config = {
    selector = {
      module_selector_by_module_name = {
        module_name = "PORT_MODE_HUB"
      }
    }
    module = {
      traffic_mode : "L1Mode"
      fiber_connection_mode : "dual"
    }
  }
}

output "hub_module" {
  value = ipm_hubmodule.hub_module
}



