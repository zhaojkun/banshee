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

type getProjectsResult struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	NumRules int    `json:"numRules"`
}

// getProjects returns all projects.
func getProjects(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var results []getProjectsResult
	err := db.Admin.DB().Table("projects").
		Joins("LEFT JOIN rules ON rules.project_id = projects.id").
		Select("projects.id as id, projects.name as name, count(rules.id) as num_rules").
		Group("projects.id").Scan(&results).Error
	if err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, results)
}

// getProject returns project by id.
func getProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Query db.
	proj := &models.Project{}
	if err := db.Admin.DB().First(proj, id).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	ResponseJSONOK(w, proj)
}

// createProject request
type createProjectRequest struct {
	Name string `json:"name"`
}

// createProject creates a project.
func createProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
	}
	//Todo check teamid
	// Request
	req := &createProjectRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validate
	if err := models.ValidateProjectName(req.Name); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	// Save.
	proj := &models.Project{Name: req.Name, TeamID: id}
	if err := db.Admin.DB().Create(proj).Error; err != nil {
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
				ResponseError(w, ErrDuplicateProjectName)
				return
			}
		}
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, proj)
}

// updateProject request
type updateProjectRequest struct {
	Name            string `json:"name"`
	EnableSilent    bool   `json:"enableSilent"`
	SilentTimeStart int    `json:"silentTimeStart"`
	SilentTimeEnd   int    `json:"silentTimeEnd"`
}

// updateProject updates a project.
func updateProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Request
	req := &updateProjectRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validate
	if err := models.ValidateProjectName(req.Name); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	if !req.EnableSilent && (req.SilentTimeStart != 0 || req.SilentTimeEnd != 0) {
		// Validate if silent is disabled and start and end both are zero.
		if err := models.ValidateProjectSilentRange(req.SilentTimeStart, req.SilentTimeEnd); err != nil {
			ResponseError(w, NewValidationWebError(err))
			return
		}
	}
	// Find
	proj := &models.Project{}
	if err := db.Admin.DB().First(proj, id).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Patch.
	proj.Name = req.Name
	proj.EnableSilent = req.EnableSilent
	proj.SilentTimeStart = req.SilentTimeStart
	proj.SilentTimeEnd = req.SilentTimeEnd
	if err := db.Admin.DB().Save(proj).Error; err != nil {
		if err == gorm.RecordNotFound {
			// Not found.
			ResponseError(w, ErrProjectNotFound)
			return
		}
		// Writer errors.
		sqliteErr, ok := err.(sqlite3.Error)
		if ok {
			switch sqliteErr.ExtendedCode {
			case sqlite3.ErrConstraintNotNull:
				ResponseError(w, ErrNotNull)
				return
			case sqlite3.ErrConstraintUnique:
				ResponseError(w, ErrDuplicateProjectName)
				return
			}
		}
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, proj)
}

// deleteProject deletes a project.
func deleteProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	proj := &models.Project{ID: id}
	// Delete Its Rules
	var rules []models.Rule
	if err := db.Admin.DB().Model(proj).Related(&rules).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	for i := 0; i < len(rules); i++ {
		db.Admin.DB().Delete(&rules[i])
		db.Admin.RulesCache.Delete(rules[i].ID)
	}
	// Delete Its user relationships.
	var users []models.User
	if err := db.Admin.DB().Model(proj).Association("Users").Find(&users).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	if err := db.Admin.DB().Model(proj).Association("Users").Delete(users).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	// Delete.
	if err := db.Admin.DB().Delete(proj).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
}

// getProjectRules gets project rules.
func getProjectRules(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Query
	var rules []models.Rule
	if err := db.Admin.DB().Model(&models.Project{ID: id}).Related(&rules).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	for i := 0; i < len(rules); i++ {
		rules[i].SetNumMetrics(len(db.Index.Filter(rules[i].Pattern)))
	}
	ResponseJSONOK(w, rules)
}

// getProjectUsers gets project users.
func getProjectUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Query
	var users []models.User
	if err := db.Admin.DB().Model(&models.Project{ID: id}).Association("Users").Find(&users).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	// Universals
	var univs []models.User
	if err := db.Admin.DB().Where("universal = ?", true).Find(&univs).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	users = append(users, univs...)
	ResponseJSONOK(w, users)
}

// addProjectUserRequest is the request of addProjectUser
type addProjectUserRequest struct {
	Name string `json:"name"`
}

// projectUser is the tempory select result for table `project_users`
type projectUser struct {
	UserID    int `sql:"user_id"`
	ProjectID int `sql:"project_id"`
}

// addProjectUser adds a user to a project.
func addProjectUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Request
	req := &addProjectUserRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Find user.
	user := &models.User{}
	if err := db.Admin.DB().Where("name = ?", req.Name).First(user).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrUserNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	if user.Universal {
		ResponseError(w, ErrProjectUniversalUser)
		return
	}
	// Find proj
	proj := &models.Project{}
	if err := db.Admin.DB().First(proj, id).Error; err != nil {
		if err == gorm.RecordNotFound {
			ResponseError(w, ErrNotFound)
			return
		}
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	// Note: Gorm only insert values to join-table if the primary key not in
	// the join-table. So we select the record at first here.
	if err := db.Admin.DB().Table("project_users").Where("user_id = ? and project_id = ?", user.ID, proj.ID).Find(&projectUser{}).Error; err == nil {
		ResponseError(w, ErrDuplicateProjectUser)
		return
	}
	// Append user.
	if err := db.Admin.DB().Model(proj).Association("Users").Append(user).Error; err != nil {
		if err == gorm.RecordNotFound {
			// User or Project not found.
			ResponseError(w, ErrNotFound)
			return
		}
		// Duplicate primay key.
		sqliteErr, ok := err.(sqlite3.Error)
		if ok {
			switch sqliteErr.ExtendedCode {
			case sqlite3.ErrConstraintUnique:
				ResponseError(w, ErrDuplicateProjectUser)
				return
			case sqlite3.ErrConstraintPrimaryKey:
				ResponseError(w, ErrDuplicateProjectUser)
				return
			}
		}
		// Unexcepted error.
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
}

// deleteProjectUser deletes a user from a project.
func deleteProjectUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	userID, err := strconv.Atoi(ps.ByName("user_id"))
	if err != nil {
		ResponseError(w, ErrUserID)
		return
	}
	// Find user.
	user := &models.User{}
	if err := db.Admin.DB().First(user, userID).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrUserNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Find proj.
	proj := &models.Project{}
	if err := db.Admin.DB().First(proj, id).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Delete user.
	if err := db.Admin.DB().Model(proj).Association("Users").Delete(user).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
}
