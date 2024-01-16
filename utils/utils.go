package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/remiges-tech/alya/wscutils"
)

type Entity int

const (
	Country Entity = iota
	State
	PostalCode
)

const (
	DIALTIMEOUT        = 5 * time.Second
	RIGELPREFIX        = "/remiges/rigel"
	INVALID_DEPENDENCY = "invalid_dependency"
)

type Status int

type Operation int

const (
	UPSERT Operation = iota
	DELETE
	ACTIVE
	INACTIVE
)

// query parameter
type GetSchemaRequestParameters struct {
	App     string `form:"app"  binding:"required"`
	Module  string `form:"module" binding:"required"`
	Version int    `form:"ver" binding:"required"`
}

type GetConfigRequestParams struct {
	App     *string `form:"app"  binding:"required"`
	Module  *string `form:"module" binding:"required"`
	Version int     `form:"ver" binding:"required"`
	Config  *string `form:"config" binding:"required"`
}

type CreateConfigRequest struct {
	App         *string `json:"app" validate:"required"`
	Module      *string `json:"module" validate:"required"`
	Version     *int    `json:"ver" validate:"required"`
	Config      *string `json:"config" validate:"required"`
	Description string  `json:"description,omitempty"`
	Name        *string `json:"name" validate:"required"`
	Value       *string `json:"value" validate:"required"`
}

type AuditTrail struct {
	CreatedAt  time.Time
	CreatedBy  string
	ApprovedAt time.Time
	ApprovedBy string
	ModifiedAt time.Time
	ModifiedBy string
	AuditEntry []AuditEntry
}

type AuditEntry struct {
	Status      Status
	UpdatedBy   string
	UpdatedAt   time.Time
	FromAddress string
}

type AppConfig struct {
	DBConnURL        string `json:"db_conn_url"`
	DBHost           string `json:"db_host"`
	DBPort           int    `json:"db_port"`
	DBUser           string `json:"db_user"`
	DBPassword       string `json:"db_password"`
	DBName           string `json:"db_name"`
	AppServerPort    string `json:"app_server_port"`
	KeycloakURL      string `json:"keycloak_url"`
	KeycloakClientID string `json:"keycloak_client_id"`
}

type IDResponse struct {
	ID *int64
}

type GetRequest struct {
	ID      int  `json:"id" validate:"required"`
	Past    bool `json:"past"`
	Current bool `json:"current"`
	Future  bool `json:"future"`
}

type Environment string

func (env Environment) IsValid() bool {
	switch env {
	case DevEnv, ProdEnv, UATEnv:
		return true
	}
	return false
}

const (
	DevEnv  Environment = "dev_env"
	ProdEnv Environment = "prod_env"
	UATEnv  Environment = "uat_env"
)

// CommonValidation is a generic function which setup standard validation utilizing
// validator package and Maps the errorVals based on the map parameter and
// return []errorVals
func CommonValidation(validationErrorVals map[string]string, err validator.FieldError) []string {
	var vals []string
	switch err.Tag() {
	case "Required":
		vals = append(vals, "NotProvided")
	}
	return vals
}

func GetErrorValidationMapByAPIName(apiName string) map[string]string {
	var validationsMap = make(map[string]map[string]string)
	validationsMap["config_create"] = map[string]string{
		"Required": "Not_Provided",
	}
	// below is one more example ::

	// validationsMap["country_draft_forward"] = map[string]string{
	// 	"IDmin": "length must be greater than one",
	// }
	return validationsMap[apiName]
}

type Node struct {
	Name     string
	Children map[string]*Node
	IsLeaf   bool
	FullPath string
	Value    string
}

func NewNode(name string) *Node {
	return &Node{
		Name:     name,
		Children: make(map[string]*Node),
		IsLeaf:   false,
		FullPath: "",
		Value:    "",
	}
}

// add nodes corresponding to the path in the node tree
func (n *Node) AddPath(path string, val string) {
	// []parts := split the path on '/'
	var parts []string
	parts = strings.Split(path, "/")
	fullPath := ""

	current := n
	// fmt.Printf("root: %v \n", current.Name)
	for i, part := range parts {
		if i == 0 {
			continue
		}
		_, exists := current.Children[part]
		// fmt.Printf("check exist %v \n", part)
		if !exists {
			current.Children[part] = NewNode(part)
			// fmt.Printf("part miss: %v \n", part)
			// fmt.Printf("part node: %v \n", current.Children[part])
		}
		current = current.Children[part]
		if i == len(parts)-1 {
			current.Value = val
			fullPath = fullPath + part
		} else {
			fullPath = fullPath + part + "/"
		}
		current.FullPath = fullPath
		// fmt.Printf("current node: %v \n", current.FullPath)
	}

}

func (n *Node) Ls(path string) []*Node {
	var parts []string
	fmt.Printf("path inside ls: %v", path)
	parts = strings.Split(path, "/")

	current := n

	for i, part := range parts {
		if i == 0 {
			continue
		}
		child, exists := current.Children[part]
		if !exists {
			//fmt.Errorf("%v not valid child", part)
			wscutils.NewErrorResponse(" Insvalid child")
		}
		current = child
	}
	var nodes []*Node
	for _, v := range current.Children {
		nodes = append(nodes, v)
	}
	return nodes
}
