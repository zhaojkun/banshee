// Copyright 2016 Eleme Inc. All rights reserved.

package webapp

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage/eventdb"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"time"
)

func getEventsByProjectID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Options
	n, err := strconv.Atoi(r.URL.Query().Get("past"))
	if err != nil {
		n = 3600 * 24 // 1 day
	}
	past := uint32(n)
	if past > cfg.Expiration {
		ResponseError(w, ErrEventPast)
		return
	}
	level, err := strconv.Atoi(r.URL.Query().Get("level"))
	if err != nil {
		level = 0 // low
	}
	if err := models.ValidateRuleLevel(level); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	end := uint32(time.Now().Unix())
	start := end - past
	ews, err := db.Event.GetByProjectID(id, level, start, end)
	if err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	if ews == nil {
		ews = make([]eventdb.EventWrapper, 0) // Note: Avoid js null
	}
	ResponseJSONOK(w, ews)
}
