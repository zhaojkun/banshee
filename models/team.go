package models

// Team is a projects group.
type Team struct {
	// ID in db.
	ID int `json:"id"`
	// Name
	Name string `sql:"not null;unique" json:"name"`
	// Project may have many rules, they shouldn't be shared.
	Project []*Project `json:"-"`
}
