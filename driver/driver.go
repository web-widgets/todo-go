package driver

import (
	"fmt"
	"web-widgets/todo-go/config"

	"gorm.io/gorm"
)

type DBDriver interface {
	GetConnection(cfg *config.DBConfig) (*gorm.DB, error)
	DataDown(db *gorm.DB)
}

func NewDriver(name string) (DBDriver, error) {
	switch name {
	case "mysql":
		return MySqlDriver{}, nil
	case "sqlite":
		return SQLiteDriver{}, nil
	}
	return nil, fmt.Errorf("unknown driver name: " + name)
}

func mustExec(db *gorm.DB, sql string) {
	err := db.Exec(sql).Error
	if err != nil {
		panic(err)
	}
}
