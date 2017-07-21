// Copyright 2015 Eleme Inc. All rights reserved.

// Package webapp implements a simple http web server to visualize detection
// results and to manage alert rules.
package webapp

import (
	"fmt"
	"net/http"

	"github.com/eleme/banshee/config"
	"github.com/eleme/banshee/filter"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/util/log"
	"github.com/julienschmidt/httprouter"
)

// Globals
var (
	// Config
	cfg *config.Config
	// Storage
	db *storage.DB
	// Filter
	flt *filter.Filter
)

// Init globals.
func Init(c *config.Config, d *storage.DB) {
	cfg = c
	db = d
}

// Start http server.
func Start(c *config.Config, d *storage.DB, f *filter.Filter) {
	// Init globals.
	cfg = c
	db = d
	flt = f
	// Auth
	auth := newAuthHandler(cfg.Webapp.Auth[0], cfg.Webapp.Auth[1])
	// Routes
	router := httprouter.New()
	// Api
	router.GET("/api/config", auth.handler(getConfig))
	router.GET("/api/interval", getInterval)
	router.GET("/api/privateDocUrl", getPrivateDocURL)
	router.GET("/api/graphiteUrl", getGraphiteURL)
	router.GET("/api/language", getLanguage)
	router.GET("/api/teams", getTeams)
	router.GET("/api/team/:id", getTeam)
	router.POST("/api/team", auth.handler(createTeam))
	router.PATCH("/api/team/:id", auth.handler(updateTeam))
	router.DELETE("/api/team/:id", auth.handler(deleteTeam))
	router.GET("/api/team/:id/projects", getTeamProjects)
	router.POST("/api/team/:id/project", auth.handler(createProject))
	router.GET("/api/projects", getProjects)
	router.GET("/api/project/:id", getProject)
	router.PATCH("/api/project/:id", auth.handler(updateProject))
	router.DELETE("/api/project/:id", auth.handler(deleteProject))
	router.GET("/api/project/:id/rules", auth.handler(getProjectRules))
	router.POST("/api/project/:id/rules", auth.handler(importProjectRules))
	router.GET("/api/project/:id/users", auth.handler(getProjectUsers))
	router.POST("/api/project/:id/user", auth.handler(addProjectUser))
	router.DELETE("/api/project/:id/user/:user_id", auth.handler(deleteProjectUser))
	router.GET("/api/project/:id/webhooks", auth.handler(getProjectWebHooks))
	router.POST("/api/project/:id/webhook", auth.handler(addProjectWebHook))
	router.DELETE("/api/project/:id/webhook/:webhook_id", auth.handler(deleteProjectWebHook))
	router.GET("/api/project/:id/events", auth.handler(getEventsByProjectID))
	router.GET("/api/events", getEvents)
	router.GET("/api/users", auth.handler(getUsers))
	router.POST("/api/users/copy", auth.handler(copyUser))
	router.GET("/api/user/:id", auth.handler(getUser))
	router.POST("/api/user", auth.handler(createUser))
	router.DELETE("/api/user/:id", auth.handler(deleteUser))
	router.PATCH("/api/user/:id", auth.handler(updateUser))
	router.GET("/api/user/:id/projects", auth.handler(getUserProjects))
	router.GET("/api/webhooks", auth.handler(getWebHooks))
	router.GET("/api/webhook/:id", auth.handler(getWebHook))
	router.POST("/api/webhook", auth.handler(createWebHook))
	router.DELETE("/api/webhook/:id", auth.handler(deleteWebHook))
	router.PATCH("/api/webhook/:id", auth.handler(updateWebHook))
	router.GET("/api/webhook/:id/projects", auth.handler(getWebHookProjects))
	router.POST("/api/project/:id/rule", auth.handler(createRule))
	router.DELETE("/api/rule/:id", auth.handler(deleteRule))
	router.PATCH("/api/rule/:id", auth.handler(editRule))
	router.GET("/api/metric/rules/:name", getMetricRules)
	router.GET("/api/metric/indexes", getMetricIndexes)
	router.GET("/api/metric/data", getMetrics)
	router.GET("/api/info", getInfo)
	router.GET("/api/version", getVersion)
	// Static
	router.NotFound = newStaticHandler(http.Dir(cfg.Webapp.Static), auth)
	// Serve
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Webapp.Port)
	log.Infof("webapp is listening and serving on %s..", addr)
	log.Fatal(http.ListenAndServe(addr, Logger(router)))
}
