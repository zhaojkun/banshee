package webapp

import (
	"net/http"
	"strconv"

	"github.com/eleme/banshee/models"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	"github.com/mattn/go-sqlite3"
)

type getTeamsResult struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	NumProjects int    `json:"numProjects"`
}

// getProjects returns all projects.
func getTeams(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var results []getTeamsResult
	err := db.Admin.DB().Table("teams").
		Joins("LEFT JOIN projects ON projects.team_id = teams.id").
		Select("teams.id as id, teams.name as name, count(projects.id) as num_projects").
		Group("teams.id").Scan(&results).Error
	if err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, results)
}

// getProject returns project by id.
func getTeam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Query db.
	team := &models.Team{}
	if err := db.Admin.DB().First(team, id).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	ResponseJSONOK(w, team)
}

// createProject request
type createTeamRequest struct {
	Name string `json:"name"`
}

// createProject creates a project.
func createTeam(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Request
	req := &createTeamRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validate
	if err := models.ValidateTeamName(req.Name); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	// Save.
	team := &models.Team{Name: req.Name}
	if err := db.Admin.DB().Create(team).Error; err != nil {
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
	ResponseJSONOK(w, team)
}

// updateProject request
type updateTeamRequest struct {
	Name string `json:"name"`
}

// updateProject updates a project.
func updateTeam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Request
	req := &updateTeamRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validate
	if err := models.ValidateProjectName(req.Name); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	// Find
	team := &models.Team{}
	if err := db.Admin.DB().First(team, id).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Patch.
	team.Name = req.Name
	if err := db.Admin.DB().Save(team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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
	ResponseJSONOK(w, team)
}

// deleteProject deletes a project.
func deleteTeam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	team := &models.Team{ID: id}
	// Delete Its Rules
	var projs []models.Project

	if err := db.Admin.DB().Model(team).Related(&projs).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	for i := 0; i < len(projs); i++ {
		db.Admin.DB().Delete(&projs[i])
	}
	// Delete.
	if err := db.Admin.DB().Delete(team).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
}

type getTeamProjectsResult struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	NumRules int    `json:"numRules"`
}

// getProjectRules gets project rules.
func getTeamProjects(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	var results []getTeamProjectsResult
	err = db.Admin.DB().Table("projects").
		Joins("LEFT JOIN rules ON rules.project_id = projects.id").
		Where("team_id = ?", id).
		Select("projects.id as id, projects.name as name, count(rules.id) as num_rules").
		Group("projects.id").Scan(&results).Error
	if err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSONOK(w, results)
}
