package rigel

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/remiges-aniket/etcd"
	"github.com/remiges-aniket/types"
)

const (
	rigelPrefix          = "/remiges/rigel"
	schemaDescriptionKey = "description"
	schemaNameKey        = "name"
	schemaVersionKey     = "version"
	schemaFieldsKey      = "fields"
	defaultEtcdEndpoints = "localhost:2379"
)

// Rigel represents a client for Rigel configuration manager server.
type Rigel struct {
	Storage types.Storage
	Cache   types.Cache
	App     string
	Module  string
	Version int
	Config  string
}

// New creates a new instance of Rigel with the provided Storage interface.
// The Storage interface is used by Rigel to interact with the underlying storage system.
// Currently, only etcd is supported as a storage system.
func New(storage types.Storage, app string, module string, version int, config string) *Rigel {
	return &Rigel{
		Storage: storage,
		Cache:   NewInMemoryCache(),
		App:     app,
		Module:  module,
		Version: version,
		Config:  config,
	}
}

// NewWithStorage creates a new instance of Rigel with the provided Storage interface.
// This function is useful when you want to create a Rigel object with a specific storage system,
// but you don't want to set the other parameters (app, module, version, config) at the time of creation.
// This is typically used in admin tasks like schema creation where version field is not known while
// adding a new schema definition. Once Rigel object is constructed using NewWithStorage other
// required params for admin tasks are supposed to be added using the with-prefixed functions like
// WithApp, WithModule, etc.
func NewWithStorage(storage types.Storage) *Rigel {
	return &Rigel{
		Storage: storage,
		Cache:   NewInMemoryCache(),
	}
}

// WithApp sets the App field of the Rigel struct and returns the modified Rigel object.
// This method is typically used for method chaining during Rigel object creation.
func (r *Rigel) WithApp(app string) *Rigel {
	r.App = app
	return r
}

// WithModule sets the Module field of the Rigel struct and returns the modified Rigel object.
// This method is typically used for method chaining during Rigel object creation.
func (r *Rigel) WithModule(module string) *Rigel {
	r.Module = module
	return r
}

// WithVersion sets the Version field of the Rigel struct and returns the modified Rigel object.
// This method is typically used for method chaining during Rigel object creation.
func (r *Rigel) WithVersion(version int) *Rigel {
	r.Version = version
	return r
}

// WithConfig sets the Config field of the Rigel struct and returns the modified Rigel object.
// This method is typically used for method chaining during Rigel object creation.
func (r *Rigel) WithConfig(config string) *Rigel {
	r.Config = config
	return r
}

// Default creates a new instance of Rigel with a default EtcdStorage instance.
func Default() (*Rigel, error) {
	_, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
	if err != nil {
		return nil, fmt.Errorf("failed to create default EtcdStorage: %w", err)
	}

	return nil, nil
}

// KeyExistsInSchema checks if a key exists in the schema.
func (r *Rigel) KeyExistsInSchema(ctx context.Context, key string) (bool, error) {
	schema, err := r.getSchema(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get schema: %w", err)
	}

	for _, field := range schema.Fields {
		if field.Name == key {
			return true, nil
		}
	}

	return false, nil
}

// Set sets a value of a config key in the storage.
func (r *Rigel) Set(ctx context.Context, configKey string, value string) error {
	// Check if the key exists in the schema
	exists, err := r.KeyExistsInSchema(ctx, configKey)
	if err != nil {
		return fmt.Errorf("failed to check if key exists in schema: %w", err)
	}
	if !exists {
		return &KeyNotFoundError{Key: configKey}
	}

	// Get the schema
	schema, err := r.getSchema(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}

	// Find the field in the schema
	var field *types.Field
	for _, f := range schema.Fields {
		if f.Name == configKey {
			field = &f
			break
		}
	}

	// Validate the value against the field's constraints
	if !validateValueAgainstConstraints(value, field) {
		return fmt.Errorf("value does not meet the constraints of the field")
	}

	// Construct the key for the parameter
	key := getConfKeyPath(r.App, r.Module, r.Version, r.Config, configKey)

	// Set the value in the storage
	err = r.Storage.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	// Update the value in the cache
	r.Cache.Set(key, value)

	return nil
}

// LoadConfig retrieves the configuration data associated with the provided configName.
// It then unmarshals this data into the provided configStruct.
//
// The configStruct parameter must be a pointer to a config struct used in the application.
// If it is not, an error will be returned.
// Non-pointer or non-struct types aren't supported due to type safety issues (e.g., unexpected fields in JSON)
// and modification restrictions, as non-pointer variables can't be updated by json.Unmarshal.
func (r *Rigel) LoadConfig(ctx context.Context, configStruct any) error {
	// Check if configStruct is a pointer to a struct
	val := reflect.ValueOf(configStruct)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("configStruct must be a pointer to a struct")
	}

	// Construct the configuration map
	configMap, err := r.constructConfigMap(ctx)
	if err != nil {
		return err
	}

	// Marshal the configuration map into a JSON string
	configJSON, err := json.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Unmarshal the JSON string into the provided configStruct
	err = json.Unmarshal(configJSON, configStruct)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config value: %w", err)
	}

	return nil
}

// AddSchema adds a new schema to the Rigel storage.
// If a schema with the same name and version already exists in the storage,
// AddSchema will override the existing schema with the new one.
func (r *Rigel) AddSchema(ctx context.Context, schema types.Schema) error {
	// Convert fields to JSON
	fieldsJson, err := json.Marshal(schema.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %v", err)
	}

	// Get the base schema path using the version from the schema
	baseSchemaPath := getSchemaPath(r.App, r.Module, schema.Version)

	// Store fields
	fieldsKey := baseSchemaPath + schemaFieldsKey
	err = r.Storage.Put(ctx, fieldsKey, string(fieldsJson))
	if err != nil {
		return fmt.Errorf("failed to store fields: %v", err)
	}

	// Store description
	descriptionKey := baseSchemaPath + schemaDescriptionKey
	err = r.Storage.Put(ctx, descriptionKey, schema.Description)
	if err != nil {
		return fmt.Errorf("failed to store description: %v", err)
	}

	return nil
}

// getSchema retrieves a schema from the storage based on the provided schemaName and schemaVersion.
func (r *Rigel) getSchema(ctx context.Context) (*types.Schema, error) {
	// Construct the base key for the schema
	schemaFieldsKey := getSchemaFieldsPath(r.App, r.Module, r.Version)

	fieldsStr, err := r.Storage.Get(ctx, schemaFieldsKey)
	if err != nil {
		return nil, err
	}
	var fields []types.Field
	err = json.Unmarshal([]byte(fieldsStr), &fields)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
	}

	// Construct the schema
	schema := &types.Schema{
		Fields: fields,
	}

	return schema, nil
}

// getConfigValue retrieves a configuration value from Rigel based on the provided schemaName, schemaVersion, and paramName.
func (r *Rigel) getConfigValue(ctx context.Context, paramName string) (string, error) {
	// Construct the key for the parameter
	key := getConfKeyPath(r.App, r.Module, r.Version, r.Config, paramName)

	// Retrieve the parameter value from the storage
	value, err := r.Storage.Get(ctx, key)
	if err != nil {
		return "", err
	}

	return value, nil
}

// constructConfigMap constructs a configuration map based on the provided schema, schemaName, and schemaVersion.
// constructConfigMap constructs a configuration map based on the Rigel object.
func (r *Rigel) constructConfigMap(ctx context.Context) (map[string]any, error) {
	// Retrieve the schema
	schema, err := r.getSchema(ctx)
	if err != nil {
		return nil, err
	}
	// Construct the configuration map
	config := make(map[string]any)
	for _, field := range schema.Fields {
		// Retrieve the configuration value for the field
		valueStr, err := r.getConfigValue(ctx, field.Name)
		if err != nil {
			return nil, err
		}

		// Convert the value to the correct type based on the field type
		value, err := convertToType(valueStr, field.Type)
		if err != nil {
			return nil, err
		}

		// Add the value to the configuration map
		config[field.Name] = value
	}
	return config, nil
}

type KeyNotFoundError struct {
	Key string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("key %s not found in config", e.Key)
}

// Get retrieves a value from the storage based on the provided key.
// It converts the retrieved value to the correct type based on the field type.
// If the field type is not "int" or "bool", the value is assumed to be a string.
// get retrieves a value from the cache or storage and returns it as a string.
func (r *Rigel) Get(ctx context.Context, configKey string) (string, error) {
	// Check if the key exists in the schema
	exists, err := r.KeyExistsInSchema(ctx, configKey)
	if err != nil {
		return "", fmt.Errorf("failed to check if key exists in schema: %w", err)
	}
	if !exists {
		return "", &KeyNotFoundError{Key: configKey}
	}

	// Construct the key for the parameter
	key := getConfKeyPath(r.App, r.Module, r.Version, r.Config, configKey)

	// Try to get the value from the cache
	value, found := r.Cache.Get(key)
	if found {
		return value, nil
	}

	// If the value is not in the cache, retrieve it from the storage
	valueStr, err := r.Storage.Get(ctx, key)
	if err != nil {
		return "", &KeyNotFoundError{Key: key}
	}

	// Store the value in the cache
	r.Cache.Set(key, valueStr)

	return valueStr, nil
}

func (r *Rigel) GetInt(ctx context.Context, configKey string) (int, error) {
	valueStr, err := r.Get(ctx, configKey)
	if err != nil {
		return 0, err
	}
	intValue, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value to int: %w", err)
	}
	return intValue, nil
}

func (r *Rigel) GetFloat(ctx context.Context, configKey string) (float64, error) {
	valueStr, err := r.Get(ctx, configKey)
	if err != nil {
		return 0, err
	}
	floatValue, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value to float: %w", err)
	}
	return floatValue, nil
}

func (r *Rigel) GetBool(ctx context.Context, configKey string) (bool, error) {
	valueStr, err := r.Get(ctx, configKey)
	if err != nil {
		return false, err
	}
	boolValue, err := strconv.ParseBool(valueStr)
	if err != nil {
		return false, fmt.Errorf("failed to convert value to bool: %w", err)
	}
	return boolValue, nil
}

func (r *Rigel) GetString(ctx context.Context, configKey string) (string, error) {
	valueStr, err := r.Get(ctx, configKey)
	if err != nil {
		return "", err
	}
	return valueStr, nil
}

// convertToType converts a string value to the specified type.
func convertToType(valueStr string, fieldType string) (interface{}, error) {
	switch fieldType {
	case "int":
		intValue, err := strconv.Atoi(valueStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to int: %w", err)
		}
		return intValue, nil
	case "bool":
		boolValue, err := strconv.ParseBool(valueStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to bool: %w", err)
		}
		return boolValue, nil
	case "float":
		floatValue, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to float: %w", err)
		}
		return floatValue, nil
	default: // "string"
		return valueStr, nil
	}
}

// WatchConfig starts watching for changes to any key in the specified configuration namespace in the storage.
// When a change is detected, it updates the corresponding key-value pair in the cache.
// The method takes the schemaName, schemaVersion, and configName
// to construct the base key for the configuration namespace.
func (r *Rigel) WatchConfig(ctx context.Context) error {
	// Construct the base key for the configuration
	baseKey := getConfPath(r.App, r.Module, r.Version, r.Config)

	events := make(chan types.Event)
	if err := r.Storage.Watch(ctx, baseKey, events); err != nil {
		return err
	}

	go func() {
		for event := range events {
			// Only update keys in the cache that have changed
			if _, found := r.Cache.Get(event.Key); found {
				r.Cache.Set(event.Key, event.Value)
			}
		}
	}()

	return nil
}
