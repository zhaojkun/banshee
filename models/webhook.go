// Copyright 2015 Eleme Inc. All rights reserved.

package models

// WebHook is the alerter message receiver.
type WebHook struct {
	// ID in db.
	ID int `gorm:"primary_key" json:"id"`
	// Name
	Name string `sql:"index;not null;unique" json:"name"`
	//Type
	Type string `json:"type"`
	// Email
	URL string `json:"url"`
	// WebHooks can been subscribed by many projects.
	Projects []*Project `gorm:"many2many:project_webhooks" json:"-"`
}
