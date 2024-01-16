package configsvc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/remiges-aniket/etcd"
	"github.com/remiges-aniket/trees"
	"github.com/remiges-aniket/utils"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"
)

// getSchemaResponse represents the structure for outgoing  responses.

type getConfigResponse struct {
	App         *string  `json:"app,omitempty"`
	Module      *string  `json:"module,omitempty"`
	Version     *int     `json:"ver,omitempty"`
	Config      *string  `json:"config,omitempty"`
	Description string   `json:"description,omitempty"`
	Values      []values `json:"values,omitempty"`
}

type values struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

func Config_get(c *gin.Context, s *service.Service) {
	lh := s.LogHarbour
	lh.Log("Config_get request received")

	client, ok := s.Dependencies["etcd"].(*etcd.EtcdStorage)
	if !ok {
		field := "etcd"
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(utils.INVALID_DEPENDENCY, &field)}))
		return
	}

	var response getConfigResponse
	var queryParams utils.GetConfigRequestParams
	err := c.ShouldBindQuery(&queryParams)
	if err != nil {
		var errCode string
		var fld string

		if strings.Contains(fmt.Sprint(err.Error()), "strconv.ParseInt") {
			errCode = "only_numbers_allowed"
			fld = "ver"
		} else {
			test := strings.Split(err.Error(), "'")
			fld = strings.Split(test[1], ".")[1]
			errCode = wscutils.ERRCODE_INVALID_REQUEST
		}
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(errCode, &fld)}))
		lh.Debug0().LogActivity("error while binding json request error:", err.Error)
		return
	}

	keyStr := utils.RIGELPREFIX + "/" + *queryParams.App + "/" + *queryParams.Module + "/" + strconv.Itoa(queryParams.Version) + "/" + *queryParams.Config

	getValue, err := client.GetWithPrefix(c, keyStr)
	if err != nil {
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(wscutils.ErrcodeMissing, nil, err.Error())}))
		lh.Debug0().LogActivity("error while get data from db error:", err.Error)
		return
	}
	// set response fields
	bindGetConfigResponse(&response, &queryParams, getValue)

	lh.Log(fmt.Sprintf("Record found: %v", map[string]any{"key with --prefix": keyStr, "value": response}))
	// te := make([]*etcdls.Node, 0)
	// arr, _ := etcdls.BuildTree(te, getValue)
	wscutils.SendSuccessResponse(c, wscutils.NewSuccessResponse(response))
}

// Config_list: handles the GET /configlist request

func Config_list(c *gin.Context, s *service.Service) {
	lh := s.LogHarbour
	lh.Log("v request received")

	client, ok := s.Dependencies["etcd"].(*etcd.EtcdStorage)
	if !ok {
		field := "etcd"
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(utils.INVALID_DEPENDENCY, &field)}))
		return
	}
	var response getConfigResponse
	var queryParams utils.GetConfigRequestParams
	err := c.ShouldBindQuery(&queryParams)
	if err != nil {
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(wscutils.ERRCODE_INVALID_REQUEST, nil, err.Error())}))
		lh.Debug0().LogActivity("error while binding json request error:", err.Error)
		return
	}

	keyStr := utils.RIGELPREFIX + "/" + *queryParams.App + "/" + *queryParams.Module + "/" + strconv.Itoa(queryParams.Version) + "/"

	getValue, err := client.GetWithPrefix(c, keyStr)
	if err != nil {
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(wscutils.ErrcodeMissing, nil, err.Error())}))
		lh.Debug0().LogActivity("error while get data from db error:", err.Error)
		return
	}

	// set response fields
	bindGetConfigResponse(&response, &queryParams, getValue)

	lh.Log(fmt.Sprintf("Record found: %v", map[string]any{"key with --prefix": keyStr, "value": response}))
	wscutils.SendSuccessResponse(c, wscutils.NewSuccessResponse(response))
}
func Config_list2(c *gin.Context, s *service.Service) {
	lh := s.LogHarbour
	lh.Log("Config_list Request Received")

	// Extracting etcdStorage and rigelTree from service dependency.

	etcd, ok := s.Dependencies["etcd"].(*etcd.EtcdStorage)
	if !ok {
		field := "etcd"
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(utils.INVALID_DEPENDENCY, &field)}))
		return
	}
	r := s.Dependencies["rTree"]
	rTree, ok := r.(*utils.Node)
	if !ok {
		field := "rigelTree"
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(utils.INVALID_DEPENDENCY, &field)}))
		return
	}

	container := &trees.Container{
		Etcd: etcd,
	}

	trees.Process(rTree, container)

	wscutils.SendSuccessResponse(c, &wscutils.Response{Status: "success", Data: container.ResponseData, Messages: []wscutils.ErrorMessage{}})
}

// bindGetConfigResponse is specifically used in Cinfig_get to bing and set the response
func bindGetConfigResponse(response *getConfigResponse, queryParams *utils.GetConfigRequestParams, getValue map[string]string) {
	response.App = queryParams.App
	response.Module = queryParams.Module
	response.Version = &queryParams.Version
	response.Config = queryParams.Config
	for key, vals := range getValue {

		arry := strings.Split(key, "/")
		keyStr := arry[len(arry)-1]
		if strings.EqualFold(keyStr, "description") {
			response.Description = vals
			continue
		} else {

			response.Values = append(response.Values, values{
				Name:  keyStr,
				Value: vals,
			})
		}

	}
}

func getValsForConfigCreateReqError(err validator.FieldError) []string {
	validationErrorVals := utils.GetErrorValidationMapByAPIName("config_create")
	return utils.CommonValidation(validationErrorVals, err)
}
