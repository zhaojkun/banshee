// Copyright 2015 Eleme Inc. All rights reserved.

package webapp

import (
	"fmt"
	"net/http"
)

// WebError is errors for web operations.
type WebError struct {
	// HTTP status code
	Code int `json:"code"`
	// Message
	Msg string `json:"msg"`
}

// Errors
var (
	// Common
	ErrBadRequest = NewWebError(http.StatusBadRequest, "Bad request")
	ErrNotNull    = NewWebError(http.StatusBadRequest, "Null value")
	ErrPrimaryKey = NewWebError(http.StatusForbidden, "Primarykey voilated")
	ErrUnique     = NewWebError(http.StatusForbidden, "Value should be unique")
	ErrNotFound   = NewWebError(http.StatusNotFound, "Not found")
	// Project
	ErrProjectID               = NewWebError(http.StatusBadRequest, "Bad project id")
	ErrProjectNotFound         = NewWebError(http.StatusNotFound, "Project not found")
	ErrDuplicateProjectName    = NewWebError(http.StatusForbidden, "Duplicate project name")
	ErrDuplicateProjectUser    = NewWebError(http.StatusForbidden, "Duplicate user to project")
	ErrProjectUniversalUser    = NewWebError(http.StatusForbidden, "Cannot add universal user to project")
	ErrProjectUniversalWebHook = NewWebError(http.StatusForbidden, "Cannot add universal webhook to project")
	ErrDuplicateProjectWebHook = NewWebError(http.StatusForbidden, "Duplicate webhook to project")

	// User
	ErrUserID            = NewWebError(http.StatusBadRequest, "Bad user id")
	ErrUserNotFound      = NewWebError(http.StatusNotFound, "User not found")
	ErrDuplicateUserName = NewWebError(http.StatusForbidden, "Duplicate user name")

	// WebHook
	ErrWebHookID            = NewWebError(http.StatusBadRequest, "Bad webhook id")
	ErrWebHookNotFound      = NewWebError(http.StatusNotFound, "WebHook not found")
	ErrDuplicateWebHookName = NewWebError(http.StatusForbidden, "Duplicate webhook name")

	// Rule
	ErrRuleID               = NewWebError(http.StatusBadRequest, "Bad rule id")
	ErrDuplicateRulePattern = NewWebError(http.StatusForbidden, "Duplicate rule pattern")
	ErrRuleNotFound         = NewWebError(http.StatusNotFound, "Rule not found")
	ErrRuleNoCondition      = NewWebError(http.StatusBadRequest, "No condition specified")
	ErrRuleNoComment        = NewWebError(http.StatusBadRequest, "No rule comment")
	ErrRuleCommentNotValid  = NewWebError(http.StatusBadRequest, "Rule comment is not valid, empty?")
	ErrRuleUpdateFailed     = NewWebError(http.StatusBadRequest, "Failed to update rule")
	// Metric
	ErrMetricNotFound = NewWebError(http.StatusNotFound, "Metric not found")
	// Event
	ErrEventPast      = NewWebError(http.StatusBadRequest, "Bad event time past")
	ErrEventTimeRange = NewWebError(http.StatusBadRequest, "Bad event time range")
)

// NewWebError creates a WebError.
func NewWebError(code int, text string) *WebError {
	return &WebError{code, text}
}

// Error returns the string format of the WebError.
func (err *WebError) Error() string {
	return fmt.Sprintf("[%d]: %s", err.Code, err.Msg)
}

// NewUnexceptedWebError returns an unexcepted WebError.
func NewUnexceptedWebError(err error) *WebError {
	return NewWebError(http.StatusInternalServerError, err.Error())
}

// NewValidationWebError creates a validation WebError from error.
func NewValidationWebError(err error) *WebError {
	return NewWebError(http.StatusBadRequest, err.Error())
}
