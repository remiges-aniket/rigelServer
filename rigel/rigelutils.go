package rigel

import (
	"fmt"

	"github.com/remiges-aniket/types"
)

// getSchemaFieldsPath constructs the path for a schema based on the provided appName, moduleName and version.
func getSchemaFieldsPath(appName string, moduleName string, version int) string {
	return fmt.Sprintf("%s/%s/%s/%d/fields", rigelPrefix, appName, moduleName, version)
}

// getConfPath constructs the path for a configuration based on the provided appName, moduleName and version.
func getConfPath(appName string, moduleName string, version int, namedConfig string) string {
	return fmt.Sprintf("%s/%s/%s/%d/config/%s", rigelPrefix, appName, moduleName, version, namedConfig)
}

// getConfKeyPath constructs the path for a configuration based on the provided appName, moduleName, version, namedConfig, and confKey.
func getConfKeyPath(appName string, moduleName string, version int, namedConfig string, confKey string) string {
	return fmt.Sprintf("%s/%s/%s/%d/config/%s/%s", rigelPrefix, appName, moduleName, version, namedConfig, confKey)
}

func GetConfKeyPath(appName string, moduleName string, version int, namedConfig string, confKey string) string {
	return fmt.Sprintf("%s/%s/%s/%d/%s/%s", rigelPrefix, appName, moduleName, version, namedConfig, confKey)
}

// getSchemaDescriptionPath constructs the path for a schema based on the provided appName, moduleName and version.
func GetSchemaDescriptionPath(appName string, moduleName string, version int) string {
	return fmt.Sprintf("%s/%s/%s/%d/description", rigelPrefix, appName, moduleName, version)
}

// getSchemaPath constructs the base key for a schema in etcd based on the provided appName, moduleName and version.
func getSchemaPath(appName string, moduleName string, version int) string {
	return fmt.Sprintf("%s/%s/%s/%d/", rigelPrefix, appName, moduleName, version)
}

func validateValueAgainstConstraints(value string, field *types.Field) bool {
	// Convert the value to the correct type
	val, err := convertToType(value, field.Type)
	if err != nil {
		return false
	}

	// Check the constraints
	if field.Constraints != nil {
		if field.Constraints.Min != nil {
			switch field.Type {
			case "int":
				if val.(int) < *field.Constraints.Min {
					return false
				}
			case "string":
				if len(val.(string)) < *field.Constraints.Min {
					return false
				}
			case "float":
				if val.(float64) < float64(*field.Constraints.Min) {
					return false
				}
			}
		}
		if field.Constraints.Max != nil {
			switch field.Type {
			case "int":
				if val.(int) > *field.Constraints.Max {
					return false
				}
			case "string":
				if len(val.(string)) > *field.Constraints.Max {
					return false
				}
			case "float":
				if val.(float64) > float64(*field.Constraints.Max) {
					return false
				}
			}
		}
		if field.Constraints.Enum != nil {
			if field.Type == "string" {
				found := false
				for _, v := range field.Constraints.Enum {
					if v == val.(string) {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}

	return true
}
