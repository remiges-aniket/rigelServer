package rigel

import (
	"testing"

	"github.com/remiges-aniket/types"
)

func TestGetSchemaFieldsPath(t *testing.T) {
	tests := []struct {
		appName      string
		moduleName   string
		version      int
		expectedPath string
	}{
		{"testApp", "testModule", 1, "/remiges/rigel/testApp/testModule/1/fields"},
	}

	for _, tt := range tests {
		path := getSchemaFieldsPath(tt.appName, tt.moduleName, tt.version)
		if path != tt.expectedPath {
			t.Errorf("Expected %s but got %s", tt.expectedPath, path)
		}
	}
}

func TestGetConfPath(t *testing.T) {
	tests := []struct {
		appName      string
		moduleName   string
		version      int
		namedConfig  string
		expectedPath string
	}{
		{"testApp", "testModule", 1, "testConf", "/remiges/rigel/testApp/testModule/1/config/testConf"},
	}

	for _, tt := range tests {
		path := getConfPath(tt.appName, tt.moduleName, tt.version, tt.namedConfig)
		if path != tt.expectedPath {
			t.Errorf("Expected %s but got %s", tt.expectedPath, path)
		}
	}
}

func TestGetConfKeyPath(t *testing.T) {
	tests := []struct {
		appName      string
		moduleName   string
		version      int
		namedConfig  string
		confKey      string
		expectedPath string
	}{
		{"testApp", "testModule", 1, "testConf", "testKey", "/remiges/rigel/testApp/testModule/1/config/testConf/testKey"},
	}

	for _, tt := range tests {
		path := getConfKeyPath(tt.appName, tt.moduleName, tt.version, tt.namedConfig, tt.confKey)
		if path != tt.expectedPath {
			t.Errorf("Expected %s but got %s", tt.expectedPath, path)
		}
	}
}

func TestGetSchemaPath(t *testing.T) {
	tests := []struct {
		appName      string
		moduleName   string
		version      int
		expectedPath string
	}{
		{"testApp", "testModule", 1, "/remiges/rigel/testApp/testModule/1/"},
	}

	for _, tt := range tests {
		path := getSchemaPath(tt.appName, tt.moduleName, tt.version)
		if path != tt.expectedPath {
			t.Errorf("Expected %s but got %s", tt.expectedPath, path)
		}
	}
}

func TestValidateValueAgainstConstraints(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		field    types.Field
		expected bool
	}{
		{
			name:  "int within range",
			value: "5",
			field: types.Field{
				Type: "int",
				Constraints: &types.Constraints{
					Min: new(int),
					Max: new(int),
				},
			},
			expected: true,
		},
		{
			name:  "int out of range",
			value: "15",
			field: types.Field{
				Type: "int",
				Constraints: &types.Constraints{
					Min: new(int),
					Max: new(int),
				},
			},
			expected: false,
		},
		{
			name:  "float within range",
			value: "3.5",
			field: types.Field{
				Type: "float",
				Constraints: &types.Constraints{
					Min: new(int),
					Max: new(int),
				},
			},
			expected: true,
		},
		{
			name:  "float out of range",
			value: "5.6",
			field: types.Field{
				Type: "float",
				Constraints: &types.Constraints{
					Min: new(int),
					Max: new(int),
				},
			},
			expected: false,
		},
		{
			name:  "string within length",
			value: "abc",
			field: types.Field{
				Type: "string",
				Constraints: &types.Constraints{
					Min: new(int),
					Max: new(int),
				},
			},
			expected: true,
		},
		{
			name:  "string out of length",
			value: "abcdef",
			field: types.Field{
				Type: "string",
				Constraints: &types.Constraints{
					Min: new(int),
					Max: new(int),
				},
			},
			expected: false,
		},
		{
			name:  "enum within range",
			value: "option1",
			field: types.Field{
				Type: "string",
				Constraints: &types.Constraints{
					Enum: []string{"option1", "option2", "option3"},
				},
			},
			expected: true,
		},
		{
			name:  "enum out of range",
			value: "option4",
			field: types.Field{
				Type: "string",
				Constraints: &types.Constraints{
					Enum: []string{"option1", "option2", "option3"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.Constraints.Min != nil && tt.field.Constraints.Max != nil {
				*tt.field.Constraints.Min = 1
				*tt.field.Constraints.Max = 5
			}
			if got := validateValueAgainstConstraints(tt.value, &tt.field); got != tt.expected {
				t.Errorf("validateValueAgainstConstraints() = %v, want %v", got, tt.expected)
			}
		})
	}
}
