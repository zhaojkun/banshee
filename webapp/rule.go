// Copyright 2015 Eleme Inc. All rights reserved.

package webapp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/eleme/banshee/models"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	sqlite3 "github.com/mattn/go-sqlite3"
)

// createRule request
type createRuleRequest struct {
	Pattern       string    `json:"pattern"`
	TrendUp       bool      `json:"trendUp"`
	TrendDown     bool      `json:"trendDown"`
	ThresholdMax  float64   `json:"thresholdMax"`
	ThresholdMin  float64   `json:"thresholdMin"`
	Comment       string    `json:"comment"`
	Level         int       `json:"level"`
	Disabled      bool      `json:"disabled"`
	DisabledFor   int       `json:"disabledFor"` // in Minute
	DisabledAt    time.Time `json:"disabledAt"`
	TrackIdle     bool      `json:"trackIdle"`
	NeverFillZero bool      `json:"neverFillZero"`
}

// createRule creates a rule.
func createRule(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	projectID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Request
	req := &createRuleRequest{
		Level:    models.RuleLevelLow,
		Disabled: false,
	}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validate
	if err := models.ValidateRulePattern(req.Pattern); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if len(req.Comment) <= 0 {
		ResponseError(w, ErrRuleNoComment)
		return
	}
	if projectID <= 0 {
		// ProjectID is invalid.
		ResponseError(w, ErrProjectID)
		return
	}
	if !req.TrendUp && !req.TrendDown && req.ThresholdMax == 0 && req.ThresholdMin == 0 {
		ResponseError(w, ErrRuleNoCondition)
		return
	}
	if err := models.ValidateRuleLevel(req.Level); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := db.Admin.DB().Where("project_id = ? AND pattern = ?", projectID, req.Pattern).First(&models.Rule{}).Error; err == nil {
		ResponseError(w, ErrDuplicateRulePattern)
		return
	}
	// Find project.
	proj := &models.Project{}
	if err := db.Admin.DB().First(proj, projectID).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Create rule.
	rule := &models.Rule{
		ProjectID:     projectID,
		Pattern:       req.Pattern,
		TrendUp:       req.TrendUp,
		TrendDown:     req.TrendDown,
		ThresholdMax:  req.ThresholdMax,
		ThresholdMin:  req.ThresholdMin,
		Comment:       req.Comment,
		Level:         req.Level,
		Disabled:      req.Disabled,
		DisabledFor:   req.DisabledFor,
		DisabledAt:    time.Now(),
		TrackIdle:     req.TrackIdle,
		NeverFillZero: req.NeverFillZero,
	}
	if err := db.Admin.DB().Create(rule).Error; err != nil {
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
			}
		}
		// Unexcepted error.
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	// Cache
	db.Admin.RulesCache.Put(rule)
	// Response
	rule.SetNumMetrics(db.Index.NumFilter(rule.Pattern))
	ResponseJSONOK(w, rule)
}

type ruleImportStatus struct {
	Rule   string
	Status error
}

func importProjectRules(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	projectID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil || projectID <= 0 {
		ResponseError(w, ErrProjectID)
		return
	}
	if err := db.Admin.DB().First(&models.Project{}, projectID).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Request
	var req []createRuleRequest
	if err := RequestFileBind(r, &req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	status := make([]ruleImportStatus, len(req), len(req))
	for i := range req {
		status[i].Rule = req[i].Pattern
		if err := models.ValidateRulePattern(req[i].Pattern); err != nil {
			status[i].Status = err
			continue
		}
		if len(req[i].Comment) <= 0 {
			status[i].Status = ErrRuleNoComment
			continue
		}
		if !req[i].TrendUp && !req[i].TrendDown && req[i].ThresholdMax == 0 && req[i].ThresholdMin == 0 {
			status[i].Status = ErrRuleNoCondition
			continue
		}
		if err := models.ValidateRuleLevel(req[i].Level); err != nil {
			status[i].Status = err
			continue
		}
		if err := db.Admin.DB().Where("project_id = ? AND pattern = ?", projectID, req[i].Pattern).First(&models.Rule{}).Error; err == nil {
			status[i].Status = ErrDuplicateRulePattern
			continue
		}

		rule := &models.Rule{
			ProjectID:     projectID,
			Pattern:       req[i].Pattern,
			TrendUp:       req[i].TrendUp,
			TrendDown:     req[i].TrendDown,
			ThresholdMax:  req[i].ThresholdMax,
			ThresholdMin:  req[i].ThresholdMin,
			Comment:       req[i].Comment,
			Level:         req[i].Level,
			Disabled:      req[i].Disabled,
			DisabledFor:   req[i].DisabledFor,
			DisabledAt:    req[i].DisabledAt,
			TrackIdle:     req[i].TrackIdle,
			NeverFillZero: req[i].NeverFillZero,
		}
		if err := db.Admin.DB().Create(rule).Error; err != nil {
			sqliteErr, ok := err.(sqlite3.Error)
			if ok {
				switch sqliteErr.ExtendedCode {
				case sqlite3.ErrConstraintNotNull:
					status[i].Status = ErrNotNull
					continue
				case sqlite3.ErrConstraintPrimaryKey:
					status[i].Status = ErrPrimaryKey
					continue
				}
			}
			status[i].Status = err
			continue
		}
		db.Admin.RulesCache.Put(rule)
	}
	ResponseJSONOK(w, status)
}

// deleteRule deletes a rule from a project.
func deleteRule(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Delete
	if err := db.Admin.DB().Delete(&models.Rule{ID: id}).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrRuleNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Cache
	db.Admin.RulesCache.Delete(id)
}

// editRule edits a rule
func editRule(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// id
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}

	// Request
	req := &createRuleRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validate
	if err := models.ValidateRulePattern(req.Pattern); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateRuleLevel(req.Level); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if len(req.Comment) <= 0 {
		ResponseError(w, ErrRuleNoComment)
		return
	}
	if !req.TrendUp && !req.TrendDown && req.ThresholdMax == 0 && req.ThresholdMin == 0 {
		ResponseError(w, ErrRuleNoCondition)
		return
	}

	rule := &models.Rule{}
	if db.Admin.DB().Where("id = ?", id).First(&rule).Error != nil {
		ResponseError(w, ErrRuleNotFound)
		return
	}

	rule.Comment = req.Comment
	rule.Level = req.Level
	rule.Pattern = req.Pattern
	rule.TrendUp = req.TrendUp
	rule.TrendDown = req.TrendDown
	rule.ThresholdMax = req.ThresholdMax
	rule.ThresholdMin = req.ThresholdMin
	rule.Disabled = req.Disabled
	rule.DisabledFor = req.DisabledFor
	rule.DisabledAt = time.Now()
	rule.TrackIdle = req.TrackIdle
	rule.NeverFillZero = req.NeverFillZero

	if db.Admin.DB().Save(rule).Error != nil {
		ResponseError(w, ErrRuleUpdateFailed)
		return
	}
	// Cache
	db.Admin.RulesCache.Delete(id)
	db.Admin.RulesCache.Put(rule)
	rule.SetNumMetrics(db.Index.NumFilter(rule.Pattern))
	ResponseJSONOK(w, rule)
}
