// Copyright 2015 Eleme Inc. All rights reserved.

package webapp

import (
	"net/http"
	"strconv"

	"github.com/eleme/banshee/models"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	"github.com/mattn/go-sqlite3"
)

// getWebHooks returns all webhooks.
func getWebHooks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var webhooks []models.WebHook
	if err := db.Admin.DB().Find(&webhooks).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, webhooks)
}

// getWebHook returns webhook by id.
func getWebHook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrWebHookID)
		return
	}
	// Query db.
	webhook := &models.WebHook{}
	if err := db.Admin.DB().First(webhook, id).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrWebHookNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	ResponseJSONOK(w, webhook)
}

// createWebHook request
type createWebHookRequest struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	RuleLevel int    `json:"ruleLevel"`
	Universal bool   `json:"universal"`
}

// createWebHook creats a webhook.
func createWebHook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Request
	req := &createWebHookRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validation
	if err := models.ValidateUserName(req.Name); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateWebHookURL(req.URL); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateRuleLevel(req.RuleLevel); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	// Save
	webhook := &models.WebHook{
		Name:      req.Name,
		Type:      req.Type,
		URL:       req.URL,
		Universal: req.Universal,
		RuleLevel: req.RuleLevel,
	}
	if err := db.Admin.DB().Create(webhook).Error; err != nil {
		// Write errors.
		sqliteErr, ok := err.(sqlite3.Error)
		if ok {
			switch sqliteErr.ExtendedCode {
			case sqlite3.ErrConstraintNotNull:
				ResponseError(w, ErrNotNull)
				return
			case sqlite3.ErrConstraintPrimaryKey:
				ResponseError(w, ErrPrimaryKey)
				return
			case sqlite3.ErrConstraintUnique:
				ResponseError(w, ErrDuplicateWebHookName)
				return
			}
		}
		// Unexcepted.
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, webhook)
}

// deleteWebHook deletes a webhook.
func deleteWebHook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrUserID)
		return
	}
	webhook := &models.WebHook{ID: id}
	// Remove its projects.
	var projs []models.Project
	if err := db.Admin.DB().Model(webhook).Association("Projects").Find(&projs).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	if err := db.Admin.DB().Model(webhook).Association("Projects").Delete(projs).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	// Delete.
	if err := db.Admin.DB().Delete(webhook).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrWebHookNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
}

// updateWebHook request
type updateWebHookRequest struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	RuleLevel int    `json:"ruleLevel"`
	Universal bool   `json:"universal"`
}

// updateWebHook updates a webhook.
func updateWebHook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrUserID)
		return
	}
	// Request
	req := &updateWebHookRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validation
	if err := models.ValidateUserName(req.Name); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateWebHookURL(req.URL); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateRuleLevel(req.RuleLevel); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	// Find
	webhook := &models.WebHook{}
	if err := db.Admin.DB().First(webhook, id).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrWebHookNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Patch
	webhook.Name = req.Name
	webhook.Type = req.Type
	webhook.URL = req.URL
	webhook.Universal = req.Universal
	webhook.RuleLevel = req.RuleLevel
	if err := db.Admin.DB().Save(webhook).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User not found.
			ResponseError(w, ErrWebHookNotFound)
			return
		}
		// Write errors.
		sqliteErr, ok := err.(sqlite3.Error)
		if ok {
			switch sqliteErr.ExtendedCode {
			case sqlite3.ErrConstraintNotNull:
				ResponseError(w, ErrNotNull)
				return
			case sqlite3.ErrConstraintUnique:
				ResponseError(w, ErrDuplicateWebHookName)
				return
			}
		}
		// Unexcepted error.
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, webhook)
}

// getWebHookProjects gets projects associate with the webhook.
func getWebHookProjects(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Get User.
	webhook := &models.WebHook{}
	if err := db.Admin.DB().First(webhook, id).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrWebHookNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Query
	var projs []models.Project
	if webhook.Universal {
		if err := db.Admin.DB().Find(&projs).Error; err != nil {
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	} else {
		// Get related projects for this webhook.
		if err := db.Admin.DB().Model(webhook).Association("Projects").Find(&projs).Error; err != nil {
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	ResponseJSONOK(w, projs)
}
