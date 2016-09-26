// Copyright 2016 Eleme Inc. All rights reserved.

package webapp

import (
	"net/http"

	"github.com/eleme/banshee/util/log"
)

// Logger output webapp visited infomation
func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		log.Infof("%s %s", r.Method, r.URL.String())
	})
}
