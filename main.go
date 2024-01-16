package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/remiges-aniket/configsvc"
	"github.com/remiges-aniket/etcd"
	"github.com/remiges-aniket/rigel"
	"github.com/remiges-aniket/schemaserv"
	"github.com/remiges-aniket/utils"
	"github.com/remiges-tech/alya/config"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"
	"github.com/remiges-tech/logharbour/logharbour"
)

func main() {

	appConfig, environment := setConfigEnvironment(utils.DevEnv)
	// Logger setup
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fallbackWriter := logharbour.NewFallbackWriter(logFile, os.Stdout)
	lctx := logharbour.NewLoggerContext(logharbour.Info)
	l := logharbour.NewLogger(lctx, "rigel", fallbackWriter)

	// Open the error types file
	file, err := os.Open("./errortypes.yaml")
	if err != nil {
		log.Fatalf("Failed to open error types file: %v", err)
	}
	defer file.Close()
	// Load the error types
	wscutils.LoadErrorTypes(file)
	// Router
	r := gin.Default()
	if environment == utils.DevEnv {
		r.Use(corsMiddleware())
	}

	// Create a new EtcdStorage instance
	etcdStorage, err := etcd.NewEtcdStorage([]string{fmt.Sprint(appConfig.DBHost + ":" + strconv.Itoa(appConfig.DBPort))})

	if err != nil {
		log.Fatalf("Failed to create EtcdStorage: %v", err)
		wscutils.NewErrorResponse("Failed to create EtcdStorage")
		return
	}

	//Create a new Rigel instance
	rigelClient := rigel.NewWithStorage(etcdStorage)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), utils.DIALTIMEOUT)
	defer cancel()

	// Get all keys from etcd
	allkeys, err := etcdStorage.GetWithPrefix(ctx, "/")
	if err != nil {
		log.Fatalf("etcd interaction failed: %v", err)
		return
	}

	// Build a Rigel STree
	rTree := utils.NewNode("")
	for k, v := range allkeys {
		rTree.AddPath(k, v)
	}

	// Services

	s := service.NewService(r).
		WithLogHarbour(l).
		WithDependency("appConfig", appConfig).
		WithDependency("rTree", rTree).
		WithDependency("etcd", etcdStorage).
		WithDependency("rigel", rigelClient)

	// Config Services
	s.RegisterRoute(http.MethodGet, "/configget", configsvc.Config_get)
	s.RegisterRoute(http.MethodGet, "/configlist", configsvc.Config_list)
	s.RegisterRoute(http.MethodPost, "/configset", configsvc.Config_set)
	s.RegisterRoute(http.MethodPost, "/configupdate", configsvc.Config_update)

	// Schema Services
	s.RegisterRoute(http.MethodGet, "/getschema", schemaserv.HandleGetSchemaRequest)
	s.RegisterRoute(http.MethodGet, "/schemalist", schemaserv.HandleGetSchemaListRequest)

	r.Run(":" + appConfig.AppServerPort)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}

func setConfigEnvironment(environment utils.Environment) (utils.AppConfig, utils.Environment) {
	var appConfig utils.AppConfig
	if !environment.IsValid() {
		log.Fatal("environment params is not valid")
	}
	switch environment {
	case utils.DevEnv:
		config.LoadConfigFromFile("./config_dev.json", &appConfig)
	case utils.ProdEnv:
		config.LoadConfigFromFile("./config_prod.json", &appConfig)
	case utils.UATEnv:
		config.LoadConfigFromFile("./config_uat.json", &appConfig)
	}
	return appConfig, environment
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
