// Copyright 2015 Eleme Inc. All rights reserved.

package models

// Project is a rules group.
type Project struct {
	// ID in db.
	ID int `json:"id"`
	// Name
	Name string `sql:"not null;unique" json:"name"`
	// Project may have many rules, they shouldn't be shared.
	Rules []*Rule `json:"-"`
	// Project may have many users, they shouldn't be shared.
	Users []*User `gorm:"many2many:project_users" json:"-"`
	// Project may have many webhooks
	WebHooks []*WebHook `gorm:"many2many:project_webhooks" json:"-"`
	// Silent time range in one day.
	EnableSilent    bool `json:"enableSilent"`
	SilentTimeStart int  `json:"silentTimeStart"`
	SilentTimeEnd   int  `json:"silentTimeEnd"`
}
