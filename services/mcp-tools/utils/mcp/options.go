package mcp

import (
	"reflect"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// ReflectToMCPOptions converts a struct definition into MCP tool options using
// reflection metadata. It parses json and jsonschema tags to construct the
// appropriate argument definitions for the mark3labs MCP server SDK.
func ReflectToMCPOptions(description string, v interface{}) []mcpgo.ToolOption {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	opts := []mcpgo.ToolOption{
		mcpgo.WithDescription(description),
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		jsonTag := f.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		name := strings.Split(jsonTag, ",")[0]

		jsSchema := f.Tag.Get("jsonschema")
		required := strings.Contains(jsSchema, "required")
		desc := extractDescription(jsSchema)

		baseType := f.Type
		if baseType.Kind() == reflect.Ptr {
			baseType = baseType.Elem()
		}

		var arg mcpgo.ToolOption
		switch baseType.Kind() {
		case reflect.String:
			if required {
				arg = mcpgo.WithString(name, mcpgo.Required(), mcpgo.Description(desc))
			} else {
				arg = mcpgo.WithString(name, mcpgo.Description(desc))
			}
		case reflect.Int:
			if required {
				arg = mcpgo.WithNumber(name, mcpgo.Required(), mcpgo.Description(desc))
			} else {
				arg = mcpgo.WithNumber(name, mcpgo.Description(desc))
			}
		case reflect.Bool:
			if required {
				arg = mcpgo.WithBoolean(name, mcpgo.Required(), mcpgo.Description(desc))
			} else {
				arg = mcpgo.WithBoolean(name, mcpgo.Description(desc))
			}
		case reflect.Slice:
			if baseType.Elem().Kind() == reflect.String {
				propertyOpts := []mcpgo.PropertyOption{mcpgo.WithStringItems()}
				if desc != "" {
					propertyOpts = append(propertyOpts, mcpgo.Description(desc))
				}
				if required {
					propertyOpts = append(propertyOpts, mcpgo.Required())
				}
				arg = mcpgo.WithArray(name, propertyOpts...)
			}
		default:
			continue
		}

		opts = append(opts, arg)
	}

	return opts
}

func extractDescription(tag string) string {
	for _, part := range strings.Split(tag, ",") {
		if strings.HasPrefix(part, "description=") {
			return strings.TrimPrefix(part, "description=")
		}
	}
	return ""
}
