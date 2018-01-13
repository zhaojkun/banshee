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

// getUsers returns all users.
func getUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var users []models.User
	if err := db.Admin.DB().Find(&users).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, users)
}

// getUser returns user by id.
func getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrUserID)
		return
	}
	// Query db.
	user := &models.User{}
	if err := db.Admin.DB().First(user, id).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrUserNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	ResponseJSONOK(w, user)
}

// createUser request
type createUserRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	EnableEmail bool   `json:"enableEmail"`
	Phone       string `json:"phone"`
	EnablePhone bool   `json:"enablePhone"`
	Universal   bool   `json:"universal"`
	RuleLevel   int    `json:"ruleLevel"`
}

// createUser creats a user.
func createUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Request
	req := &createUserRequest{
		EnableEmail: true,
		EnablePhone: true,
		Universal:   false,
		RuleLevel:   models.RuleLevelLow,
	}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validation
	if err := models.ValidateUserName(req.Name); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateUserEmail(req.Email); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateUserPhone(req.Phone); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateRuleLevel(req.RuleLevel); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	// Save
	user := &models.User{
		Name:        req.Name,
		Email:       req.Email,
		EnableEmail: req.EnableEmail,
		Phone:       req.Phone,
		EnablePhone: req.EnablePhone,
		Universal:   req.Universal,
		RuleLevel:   req.RuleLevel,
	}
	if err := db.Admin.DB().Create(user).Error; err != nil {
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
				ResponseError(w, ErrDuplicateUserName)
				return
			}
		}
		// Unexcepted.
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, user)
}

// deleteUser deletes a user.
func deleteUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrUserID)
		return
	}
	user := &models.User{ID: id}
	// Remove its projects.
	var projs []models.Project
	if err := db.Admin.DB().Model(user).Association("Projects").Find(&projs).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	if len(projs) > 0 {
		if err := db.Admin.DB().Model(user).Association("Projects").Delete(projs).Error; err != nil {
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Delete.
	if err := db.Admin.DB().Delete(user).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrUserNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	ResponseJSONOK(w, nil)
}

// updateUser request
type updateUserRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	EnableEmail bool   `json:"enableEmail"`
	Phone       string `json:"phone"`
	EnablePhone bool   `json:"enablePhone"`
	Universal   bool   `json:"universal"`
	RuleLevel   int    `json:"ruleLevel"`
}

// updateUser updates a user.
func updateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrUserID)
		return
	}
	// Request
	req := &updateUserRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validation
	if err := models.ValidateUserName(req.Name); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateUserEmail(req.Email); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateUserPhone(req.Phone); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if err := models.ValidateRuleLevel(req.RuleLevel); err != nil {
		ResponseError(w, NewValidationWebError(err))
	}
	// Find
	user := &models.User{}
	if err := db.Admin.DB().First(user, id).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrUserNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Patch
	user.Name = req.Name
	user.Email = req.Email
	user.EnableEmail = req.EnableEmail
	user.Phone = req.Phone
	user.EnablePhone = req.EnablePhone
	user.Universal = req.Universal
	user.RuleLevel = req.RuleLevel
	if err := db.Admin.DB().Save(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User not found.
			ResponseError(w, ErrUserNotFound)
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
				ResponseError(w, ErrDuplicateUserName)
				return
			}
		}
		// Unexcepted error.
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, user)
}

// getUserProjects gets usr projects.
func getUserProjects(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Get User.
	user := &models.User{}
	if err := db.Admin.DB().First(user, id).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrUserNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Query
	var projs []models.Project
	if user.Universal {
		// Get all projects for universal user.
		if err := db.Admin.DB().Find(&projs).Error; err != nil {
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	} else {
		// Get related projects for this user.
		if err := db.Admin.DB().Model(user).Association("Projects").Find(&projs).Error; err != nil {
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	ResponseJSONOK(w, projs)
}

type copyUserRequest struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func copyUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := &copyUserRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	var userA, userB models.User
	if err := db.Admin.DB().First(&userA, req.From).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrUserNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	if err := db.Admin.DB().First(&userB, req.To).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrUserNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// upgrade user B from user A
	if userA.EnableEmail {
		userB.EnableEmail = true
	}
	if userA.EnablePhone {
		userB.EnablePhone = true
	}
	if userA.Universal {
		userB.Universal = true
	}
	if userA.RuleLevel > userB.RuleLevel {
		userB.RuleLevel = userA.RuleLevel
	}
	err := db.Admin.DB().Save(&userB).Error
	if err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	// copy projects  to user b
	var projUserA, projUserB []projectUser
	err = db.Admin.DB().Table("project_users").Where("user_id = ?", userA.ID).Find(&projUserA).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	err = db.Admin.DB().Table("project_users").Where("user_id = ?", userB.ID).Find(&projUserB).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	projsMB := make(map[int]bool)
	for _, proj := range projUserB {
		projsMB[proj.ProjectID] = true
	}
	projs := make(map[int]bool)
	for _, proj := range projUserA {
		if _, ok := projsMB[proj.ProjectID]; !ok {
			projs[proj.ProjectID] = true
		}
	}
	for key := range projs {
		err := db.Admin.DB().Table("project_users").Save(&projectUser{
			UserID:    userB.ID,
			ProjectID: key,
		}).Error
		if err != nil {
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	ResponseJSONOK(w, userB)
}
