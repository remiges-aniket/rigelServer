package configsvc

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/remiges-aniket/utils"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"
	"github.com/remiges-tech/logharbour/logharbour"
	"github.com/remiges-aniket/rigel"
)

func Config_update(c *gin.Context, s *service.Service) {
	l := s.LogHarbour
	l.Log("Starting execution of Config_set()")

	var configupdate configupdate
	err := wscutils.BindJSON(c, &configupdate)
	if err != nil {
		l.LogActivity("error while binding json", err)
		return
	}

	// Validate the user creation request
	validationErrors := validateConfigupdate(configupdate, c)
	if len(validationErrors) > 0 {
		l.LogDebug("Validation errors:", logharbour.DebugInfo{Variables: map[string]any{"validationErrors": validationErrors}})
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, validationErrors))
		return
	}

	// Extracting Rigel client from service dependency and initializing with values from request parameters.
	rigelClient := s.Dependencies["rigel"]
	r, ok := rigelClient.(*rigel.Rigel)
	if !ok {
		str := "rigelClient"
		l.Debug0().LogDebug("Invalid Rigel Client Dependency:", logharbour.DebugInfo{Variables: map[string]any{"rigelClient": rigelClient}})
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(utils.INVALID_DEPENDENCY, &str)}))
		return
	}
	r.WithApp(configupdate.App).WithModule(configupdate.Module).WithVersion(configupdate.Ver).WithConfig(configupdate.Config)

	for _, v := range configupdate.Values {
		err = r.Set(c, v.Name, v.Value)
		if err != nil {
			l.LogActivity("error while setting value in etcd:", err)
			wscutils.SendErrorResponse(c, wscutils.NewErrorResponse("unable_to_set"))
			return
		}
	}
	wscutils.SendSuccessResponse(c, &wscutils.Response{Status: wscutils.SuccessStatus, Data: "data set successfully", Messages: []wscutils.ErrorMessage{}})
}

// validateConfigupdate performs validation for the Configupdate.
func validateConfigupdate(config configupdate, c *gin.Context) []wscutils.ErrorMessage {
	// Validate the request body
	validationErrors := wscutils.WscValidate(config, config.getValsForUser)

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return validationErrors
}

// getValsForUser returns validation error details based on the field and tag.
func (config *configupdate) getValsForUser(err validator.FieldError) []string {
	return nil
}
