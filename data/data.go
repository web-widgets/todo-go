package data

import "time"

type TaskProps struct {
	Text           string     `json:"text"`
	Checked        bool       `json:"checked"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	CompletionDate *time.Time `json:"completion_date,omitempty"`
	CreationDate   *time.Time `json:"creation_date,omitempty"`
	EditionDate    *time.Time `json:"edited_date,omitempty"`
	ParentID       int        `json:"parent,omitempty" gorm:"column:parent"`
	ProjectID      int        `json:"project,omitempty" gorm:"column:project"`

	AssignedUsers    []User `gorm:"many2many:assigned_users" json:"-"`
	AssignedUsersIDs []int  `gorm:"-" json:"assigned"`
}

type Task struct {
	TaskProps
	ID    int `json:"id"`
	Index int `json:"index"`
}

type Project struct {
	ID   int    `json:"id"`
	Name string `json:"label" gorm:"column:label"`
}

type User struct {
	ID     int    `json:"id"`
	Name   string `json:"label" gorm:"column:label"`
	Avatar string `json:"avatar" gorm:"column:path"`

	AssignedCards []Task `gorm:"many2many:assigned_users" json:"-"`
}

type Tag struct {
	ID   int
	Name string
}
