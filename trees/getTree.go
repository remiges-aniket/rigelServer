package trees

import (
	"context"
	"strconv"

	"github.com/remiges-aniket/etcd"
	"github.com/remiges-aniket/rigel"
	"github.com/remiges-aniket/utils"
	"github.com/remiges-tech/alya/wscutils"
)

// Container is used by getSchemaList handler to create and enrich
// schema data it will eventually return. This is required to maintain the state
// of the record while it goes through multiple iterations: first for list of apps,
// then the list of modules inside the app and then versions.
// Container holds the state for each iteration.
type Container struct {
	appName      string
	moduleName   string
	version      int
	Config       string
	Description  string
	Etcd         *etcd.EtcdStorage
	ResponseData []any
}

type GetConfigListResponse struct {
	App         string `json:"app"`
	Module      string `json:"module"`
	Ver         int    `json:"ver"`
	Config      string `json:"config"`
	Description string `json:"description"`
}

func Process(rTree *utils.Node, c *Container) {
	appNodes := rTree.Ls(utils.RIGELPREFIX)
	for _, n := range appNodes {
		workOnApps(n, rTree, c)
	}
}

func workOnApps(n *utils.Node, rTree *utils.Node, c *Container) {
	appName := n.Name
	moduleNodes := rTree.Ls(utils.RIGELPREFIX + "/" + appName)
	c.appName = appName
	// modules
	for _, m := range moduleNodes {
		workOnModules(m, rTree, c)
	}
}

func workOnModules(m *utils.Node, rTree *utils.Node, c *Container) {
	mName := m.Name
	versionNodes := rTree.Ls(utils.RIGELPREFIX + "/" + c.appName + "/" + mName)
	c.moduleName = mName

	for _, v := range versionNodes {
		workOnVersions(v, rTree, c)
	}
}

func workOnVersions(v *utils.Node, rTree *utils.Node, c *Container) {
	vName := v.Name
	vInt, err := strconv.Atoi(vName)
	if err != nil {
		wscutils.NewErrorResponse("invalid version")
		return
	}

	c.version = vInt
	configNodes := rTree.Ls(utils.RIGELPREFIX + "/" + c.appName + "/" + c.moduleName + "/" + vName)

	for _, conf := range configNodes {
		workOnConfigs(conf, rTree, c)
	}

}

func workOnConfigs(conf *utils.Node, rTree *utils.Node, c *Container) {
	c.Config = conf.Name
	GenerateResponse(c)
}

func GenerateResponse(c *Container) {
	descr := getConfigDescr(c)
	response := GetConfigListResponse{
		App:         c.appName,
		Ver:         c.version,
		Module:      c.moduleName,
		Config:      c.Config,
		Description: descr,
	}
	c.ResponseData = append(c.ResponseData, response)
}

func getDescr(t *Container) string {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), utils.DIALTIMEOUT)
	defer cancel()

	descr, err := t.Etcd.Get(ctx, rigel.GetSchemaDescriptionPath(t.appName, t.moduleName, t.version)) // vInt))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			wscutils.NewErrorResponse("description get timed out")
		} else {
			wscutils.NewErrorResponse("description get failed")
		}
		return ""
	}

	return descr
}

func getConfigDescr(t *Container) string {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), utils.DIALTIMEOUT)
	defer cancel()

	descr, err := t.Etcd.Get(ctx, rigel.GetConfKeyPath(t.appName, t.moduleName, t.version, t.Config, "description")) // vInt))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			wscutils.NewErrorResponse("description get timed out")
		} else {
			wscutils.NewErrorResponse("description get failed")
		}
		return ""
	}

	return descr
}
