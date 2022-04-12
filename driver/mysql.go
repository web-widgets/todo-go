package driver

import (
	"fmt"
	"web-widgets/todo-go/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySqlDriver struct{}

func (d MySqlDriver) GetConnection(c *config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	return db, err
}

func (d MySqlDriver) DataDown(db *gorm.DB) {
	mustExec(db, "SET FOREIGN_KEY_CHECKS = 0")
	mustExec(db, "TRUNCATE TABLE tags")
	mustExec(db, "TRUNCATE TABLE tasks")
	mustExec(db, "TRUNCATE TABLE users")
	mustExec(db, "TRUNCATE TABLE projects")
	mustExec(db, "TRUNCATE TABLE assigned_users")
	mustExec(db, "SET FOREIGN_KEY_CHECKS = 1")
}
