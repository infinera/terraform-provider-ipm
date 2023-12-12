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
resource "ipm_constellation_network" "constellation_networks" {
  config = {
    name                    = "Network1"
    constellation_frequency = 193300000
    //modulation              = "QPSK"
    topology: "auto"
  }
  hub_module = {
    config = {
      selector = {
        module_selector_by_module_name = {
          module_name = "Test_HUB1"
        }
      }
      module = {
        traffic_mode : "L1Mode"
        //fiber_connection_mode : "dual"
      }
    }
  }
  /*leaf_modules = [
    { 
      config = {
        selector = {
          module_selector_by_module_name = {
          module_name = "A-LRN-Leaf5"
          }
        }
        module = {
          trafficMode: "L1Mode"
          fiberConnectionMode: "dual"
          requestedNominalPsdOffset: "0dB"
        }
      }
    }
  ]*/
}

output "constellation_networks" {
  value = ipm_constellation_network.constellation_networks
}



