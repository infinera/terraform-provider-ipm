variable "mqttServers" {
  type = list(object({
                       identifier = object({module_name = optional(string), module_id = optional(string), server_id = optional(string)}),
                      server = optional(string), port = optional(number), kai = optional(number),
                      type = optional(string), sub_type = optional(number), enabled = optional(bool),
                      log_level = optional(number), region = optional(string)
                      }))
  description = "List of mqtt Servers"
  default = [
              { identifier = {module_name = "MA122232A001", module_id = "0cc1b779-80d4-460a-412f-c7bc24f4887e", server_id = "1"},
                server = "10.100.69.138", port = 31883, kai = 5,
                      type = "fault", sub_type = 65535, enabled = true,
                      log_level = 4, region = "1"
              },
              { identifier = {module_name = "MA222309A010", module_id = "06092e9a-9fda-47cb-6812-68709b576470", server_id = "1"},
                server = "10.100.69.138", port = 31883, kai = 5,
                      type = "fault", sub_type = 65535, enabled = true,
                      log_level = 4, region = "1"
              },
              { identifier = {module_name = "MA222324A004", module_id = "7e574b35-cef9-4e53-5bb5-bc3a327e665c", server_id = "1"},
                server = "10.100.69.138", port = 31883, kai = 5,
                      type = "fault", sub_type = 65535, enabled = true,
                      log_level = 4, region = "1"
              },
            ]
}


