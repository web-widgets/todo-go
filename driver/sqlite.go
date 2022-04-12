package driver

import (
	"web-widgets/todo-go/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteDriver struct{}

func (d SQLiteDriver) GetConnection(c *config.DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(c.Path), &gorm.Config{})
	return db, err
}

func (d SQLiteDriver) DataDown(db *gorm.DB) {
	mustExec(db, "DELETE FROM tags")
	mustExec(db, "DELETE FROM tasks")
	mustExec(db, "DELETE FROM users")
	mustExec(db, "DELETE FROM projects")
	mustExec(db, "DELETE FROM assigned_users")
}
