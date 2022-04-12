package data

import (
	"gorm.io/gorm"
)

func NewProjectsDAO(db *gorm.DB) *ProjectsDAO {
	return &ProjectsDAO{db}
}

type ProjectTemp struct {
	ID   *int   `json:"id"`
	Name string `json:"label"`
}

type ProjectUpdate struct {
	Name string `json:"label"`
}

type ProjectsDAO struct {
	db *gorm.DB
}

func (d *ProjectsDAO) GetOne(id int) (Project, error) {
	project := Project{}
	err := d.db.Find(&project, id).Error
	return project, err
}

func (d *ProjectsDAO) GetAll() ([]ProjectTemp, error) {
	projects := make([]Project, 0)
	err := d.db.Find(&projects).Error

	if err == nil && len(projects) > 0 {
		temp := make([]ProjectTemp, len(projects)+1)
		temp[0] = ProjectTemp{nil, "No project"}
		for i := range projects {
			temp[i+1].ID = &projects[i].ID
			temp[i+1].Name = projects[i].Name
		}
		return temp, nil
	}

	return nil, err
}

func (d *ProjectsDAO) Add(update *ProjectUpdate) (int, error) {
	project := Project{}
	update.fillModel(&project)

	err := d.db.Save(&project).Error
	return project.ID, err
}

func (d *ProjectsDAO) Update(id int, update *ProjectUpdate) error {
	project, err := d.GetOne(id)
	if err != nil {
		return err
	}

	update.fillModel(&project)
	err = d.db.Save(&project).Error
	return err
}

func (d *ProjectsDAO) Delete(id int) error {
	tasks := make([]Task, 0)
	err := d.db.Select("id").Where("project = ?", id).Find(&tasks).Error
	if err != nil {
		return err
	}

	ids := make([]int, len(tasks))
	for i := range tasks {
		ids[i] = tasks[i].ID
	}

	err = d.db.Exec("DELETE FROM assigned_users WHERE task_id IN ?", ids).Error
	if err != nil {
		return err
	}
	err = d.db.Where("id IN ?", ids).Delete(&Task{}).Error
	if err != nil {
		return err
	}
	err = d.db.Delete(&Project{}, id).Error

	return err
}

func (d *ProjectUpdate) fillModel(ev *Project) {
	if ev != nil {
		ev.Name = d.Name
	}
}
