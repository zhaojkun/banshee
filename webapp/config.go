// Copyright 2015 Eleme Inc. All rights reserved.

package webapp

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// getConfig returns config.
func getConfig(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := cfg.Copy()
	c.Webapp.Auth[0] = "******"
	c.Webapp.Auth[1] = "******"
	ResponseJSONOK(w, c)
}

// getInterval returns config.interval.
type intervalResponse struct {
	Interval uint32 `json:"interval"`
}

func getInterval(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ResponseJSONOK(w, &intervalResponse{cfg.Interval})
}

// getLanguage returns config.webapp.language.
type languageResponse struct {
	Language string `json:"language"`
}

func getLanguage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ResponseJSONOK(w, &languageResponse{cfg.Webapp.Language})
}

// getPrivateDocURL returns config.webapp.privateDocUrl.
type privateDocURLResponse struct {
	PrivateDocURL string `json:"privateDocUrl"`
}

func getPrivateDocURL(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ResponseJSONOK(w, &privateDocURLResponse{cfg.Webapp.PrivateDocURL})
}

// getGraphiteURL returns config.webapp.graphiteUrl.
type getGraphiteURLResponse struct {
	GraphiteURL string `json:"graphiteUrl"`
}

func getGraphiteURL(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ResponseJSONOK(w, &getGraphiteURLResponse{cfg.Webapp.GraphiteURL})
}
