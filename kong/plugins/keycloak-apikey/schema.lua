local typedefs = require "kong.db.schema.typedefs"

return {
  name = "keycloak-apikey",
  fields = {
    { config = {
        type = "record",
        fields = {
          { validation_url = {
              type = "string",
              required = true,
              default = "http://llm-api:8080/auth/validate-api-key",
              description = "URL of the API key validation endpoint"
          }},
          { validation_timeout = {
              type = "number",
              required = true,
              default = 5000,
              description = "Timeout for validation request in milliseconds"
          }},
          { hide_credentials = {
              type = "boolean",
              required = true,
              default = true,
              description = "Hide API key from downstream services"
          }},
          { run_on_preflight = {
              type = "boolean",
              required = true,
              default = false,
              description = "Run on CORS preflight requests"
          }},
        }
    }},
  },
}
