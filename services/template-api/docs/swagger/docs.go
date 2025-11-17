// Package swagger provides API documentation
package swagger

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}{
	Version:     "1.0",
	Host:        "",
	BasePath:    "/",
	Schemes:     []string{},
	Title:       "Template API",
	Description: "Template API Service",
}

// Placeholder for swagger documentation
// Run 'swag init' to generate complete API documentation
