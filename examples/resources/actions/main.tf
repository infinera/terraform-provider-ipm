terraform {
  required_providers {
    ipm = {
      source = "infinera.com/poc/ipm"
    }
  }
}

provider "ipm" {
  username = "xr-user-1"
  password = "Infinera#1"
  host     = "sv-osgams2-dt"
}

// Constellation Network Resource supports CRUD functions
resource "ipm_actions" "actions" {
  resource_actions = [
    {
      identifier = {
        module_id = "617a5eeb-675e-4704-8642-e770cc9c3023"
      }
      type = "module"
      action = "coldStart"
    },
    {
      identifier = {
        module_id = "617a5eeb-675e-4704-8642-e770cc9c3023"
      }
      type = "module"
      action = "warmStart"
    }
  ]
}

output "actions" {
  value = ipm_actions.actions
}



