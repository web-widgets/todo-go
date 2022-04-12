package data

import (
	"log"
	"web-widgets/todo-go/config"
	"web-widgets/todo-go/driver"

	"gorm.io/gorm"
)

type DAO struct {
	db *gorm.DB

	Tasks    *TasksDAO
	Users    *UsersDAO
	Projects *ProjectsDAO
	Tags     *TagsDAO
}

func (d *DAO) GetDB() *gorm.DB {
	return d.db
}

func NewDAO(config *config.DBConfig, url string) *DAO {
	driver, err := driver.NewDriver(config.Type)
	if err != nil {
		log.Fatal(err)
	}

	db, err := driver.GetConnection(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	db.AutoMigrate(&Task{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Project{})
	db.AutoMigrate(&Tag{})

	dao := &DAO{
		db:       db,
		Tasks:    NewTasksDAO(db),
		Users:    NewUsersDAO(db),
		Projects: NewProjectsDAO(db),
		Tags:     NewTagsDAO(db),
	}

	if config.ResetOnStart {
		driver.DataDown(db)
		dataUp(dao)
	}

	return dao
}
