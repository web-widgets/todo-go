package data

import (
	"gorm.io/gorm"
)

type TaskUpdate struct {
	TaskProps
	HelperID int    `json:"targetId"`
	Bunch    []Task `json:"bunch"`
}

type TaskTemp struct {
	TaskProps
	ID       string `json:"id"`
	ParentID string `json:"parent,omitempty"`
}

type MoveInfo struct {
	ParentID  int    `json:"parent"`
	HelperID  int    `json:"targetId"`
	ProjectID int    `json:"project"`
	Operation string `json:"operation"`
	Reverse   bool   `json:"reverse"`
}

type PasteInfo struct {
	HelperID  int        `json:"targetId"`
	ParentID  int        `json:"parent"`
	ProjectID int        `json:"project"`
	Bunch     []TaskTemp `json:"bunch"`
}

type TasksDAO struct {
	db *gorm.DB
}

func NewTasksDAO(db *gorm.DB) *TasksDAO {
	return &TasksDAO{db}
}

func (d *TasksDAO) GetOne(id int) (*Task, error) {
	task := Task{}
	err := d.db.Find(&task, id).Error
	return &task, err
}

func (d *TasksDAO) GetAll() ([]Task, error) {
	tasks := make([]Task, 0)
	err := d.db.
		Order("project, `index` asc").
		Preload("AssignedUsers").
		Find(&tasks).Error

	for i, c := range tasks {
		tasks[i].AssignedUsersIDs = getIDs(c.AssignedUsers)
	}
	return tasks, err
}

func (d *TasksDAO) GetFromProject(id int) ([]Task, error) {
	var err error
	tasks := make([]Task, 0)
	err = d.db.
		Where("project = ?", id).
		Order("`index` asc").
		Preload("AssignedUsers").
		Find(&tasks).Error

	for i, c := range tasks {
		tasks[i].AssignedUsersIDs = getIDs(c.AssignedUsers)
	}
	return tasks, err
}

func (d *TasksDAO) Add(update *TaskUpdate) (int, error) {
	var err error
	var index int
	if update.HelperID != 0 {
		var helperTask *Task
		helperTask, err = d.GetOne(update.HelperID)
		if err != nil {
			return 0, err
		}

		if update.HelperID == update.ParentID {
			// add sub-task
			index, err = d.getMinIndex(helperTask.ProjectID, helperTask.ID)
		} else {
			// add task below
			index = helperTask.Index
			var direction int
			direction, err = d.updateIndex(helperTask.ProjectID, helperTask.ParentID, index, 1)
			if direction > 0 {
				index++
			}
		}
	} else {
		// add task at the start of the tree
		index, err = d.getMinIndex(update.ProjectID, 0)
	}
	if err != nil {
		return 0, err
	}
	task := update.toModel()
	task.Index = index
	err = d.db.Create(&task).Error

	return int(task.ID), err
}

func (d *TasksDAO) Update(id int, update *TaskUpdate) (err error) {
	task, err := d.GetOne(id)
	if err != nil {
		return
	}

	tx := d.openTX()
	defer d.closeTX(tx, err)

	task.Text = update.Text
	task.Checked = update.Checked
	task.DueDate = update.DueDate

	err = tx.Model(&task).Association("AssignedUsers").Clear()
	if err != nil {
		return
	}
	if len(update.AssignedUsersIDs) > 0 {
		users := make([]User, 0)
		err := d.db.Where("id IN(?)", update.AssignedUsersIDs).Find(&users).Error
		if err != nil {
			return err
		}
		task.AssignedUsers = users
	}
	err = tx.Save(&task).Error
	if err != nil {
		return
	}

	if len(update.Bunch) > 0 {
		for _, v := range update.Bunch {
			t := Task{}
			err = d.db.Find(&t, v.ID).Error
			if err != nil {
				return
			}
			err = tx.Save(&v).Error
			if err != nil {
				return
			}
		}
	}

	return
}

func (d *TasksDAO) Delete(id int) error {
	root, err := d.GetOne(id)
	if err != nil {
		return err
	}
	ids, err := d.getChildrenIDs(root.ProjectID, id)
	if err != nil {
		return err
	}
	ids = append(ids, id)

	err = d.db.Exec("DELETE FROM assigned_users WHERE task_id IN ?", ids).Error
	if err == nil {
		err = d.db.Where("id IN ?", ids).Delete(&Task{}).Error
	}
	if err == nil {
		_, err = d.updateIndex(root.ProjectID, root.ParentID, root.Index, -1)
	}

	return err
}

func (d *TasksDAO) Move(id int, info *MoveInfo) error {
	if id == info.ParentID || id == info.HelperID {
		return nil
	}

	switch info.Operation {
	case "project":
		return d.moveToProject(id, info)
	case "indent", "unindent":
		return d.shiftTask(id, info)
	}
	return d.moveTask(id, info)
}

func (d *TasksDAO) Paste(info *PasteInfo) (idPull map[string]int, err error) {
	if info == nil || len(info.Bunch) == 0 {
		return nil, nil
	}

	tx := d.openTX()
	defer d.closeTX(tx, err)

	helperTask, err := d.GetOne(info.HelperID)
	if err != nil {
		return
	}

	task := info.Bunch[0].toModel()
	task.ParentID = info.ParentID
	index := helperTask.Index

	var dir int
	dir, err = d.updateIndexTX(helperTask.ProjectID, helperTask.ParentID, index, 1, tx)
	if dir > 0 {
		index++
	}

	task.Index = index
	err = tx.Create(&task).Error
	if err != nil {
		return nil, err
	}

	idPull = make(map[string]int)
	idPull[info.Bunch[0].ID] = task.ID

	if len(info.Bunch) > 1 {
		indexPull := make(map[int]int)

		for i := 1; i < len(info.Bunch); i++ {
			v := info.Bunch[i]
			task := v.toModel()
			task.ParentID = idPull[v.ParentID]
			task.Index = indexPull[task.ParentID]

			err := tx.Create(&task).Error
			if err != nil {
				return nil, err
			}

			idPull[v.ID] = task.ID
			indexPull[task.ParentID]++
		}
	}

	return idPull, nil
}

// helpres

func (d TaskProps) toModel() *Task {
	return &Task{
		TaskProps: d,
	}
}

func (d *TasksDAO) openTX() *gorm.DB {
	return d.db.Begin()
}

func (d *TasksDAO) closeTX(tx *gorm.DB, err error) {
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
}

func (d *TasksDAO) moveToProject(id int, info *MoveInfo) error {
	var task Task
	err := d.db.Find(&task, id).Error
	if err != nil {
		return err
	}

	oldProject := task.ProjectID
	oldParent := task.ParentID
	oldIndex := task.Index

	task.ParentID = 0
	task.ProjectID = info.ProjectID
	task.Index, err = d.getMaxIndex(info.ProjectID, 0)
	if err != nil {
		return err
	}

	err = d.db.Save(task).Error
	if err != nil {
		return err
	}

	taskChildren, err := d.getChildrenIDs(oldProject, id)
	if err != nil {
		return err
	}

	if len(taskChildren) > 0 {
		err = d.db.Model(&Task{}).Where("id IN ?", taskChildren).Update("project", info.ProjectID).Error
		if err != nil {
			return err
		}
	}

	if err == nil {
		_, err = d.updateIndex(oldProject, oldParent, oldIndex, -1)
	}

	return nil
}

func (d *TasksDAO) moveTask(id int, info *MoveInfo) error {
	var task, helperTask *Task
	task, err := d.GetOne(id)
	if err != nil {
		return err
	}

	if info.Reverse {
		helperTask, err = d.GetOne(info.HelperID)
	} else {
		helperTask, err = d.getNextTaskByIndex(task.ProjectID, task.ParentID, task.Index)
	}
	if err != nil || helperTask == nil {
		return err
	}

	task.Index, helperTask.Index = helperTask.Index, task.Index
	err = d.db.Save(&task).Error
	if err == nil {
		err = d.db.Save(&helperTask).Error
	}
	return err
}

func (d *TasksDAO) shiftTask(id int, info *MoveInfo) error {
	var task, parentTask *Task
	task, err := d.GetOne(id)
	if err != nil {
		return err
	}

	var index int
	if info.Operation == "indent" {
		parentTask, err = d.GetOne(info.ParentID)
		if err != nil {
			return err
		}
		index, err = d.getMaxIndex(task.ProjectID, info.ParentID)
		if err == nil {
			_, err = d.updateIndex(task.ProjectID, task.ParentID, parentTask.Index, -1)
		}
	} else {
		parentTask, err = d.GetOne(task.ParentID)
		if err != nil {
			return err
		}
		var nextTask *Task
		nextTask, err = d.getNextTaskByIndex(task.ProjectID, parentTask.ParentID, parentTask.Index)
		if err != nil {
			return err
		}
		if nextTask == nil {
			index = parentTask.Index + 1
		} else {
			index = nextTask.Index
			var dir int
			dir, err = d.updateIndex(task.ProjectID, parentTask.ParentID, index-1, 1)
			if dir < 0 {
				index--
			}
		}

	}
	if err == nil {
		task.Index = index
		task.ParentID = info.ParentID
		err = d.db.Save(task).Error
	}

	return err
}

func (d *TasksDAO) getMinIndex(projectID, parentID int) (int, error) {
	task := Task{}
	err := d.db.
		Where("project = ? AND parent = ?", projectID, parentID).
		Order("`index` ASC").
		Take(&task).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return task.Index - 1, err
}

func (d *TasksDAO) getMaxIndex(projectID, parentID int) (int, error) {
	task := Task{}
	err := d.db.
		Where("project = ? AND parent = ?", projectID, parentID).
		Order("`index` DESC").
		Take(&task).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return task.Index + 1, err
}

func (d *TasksDAO) getNextTaskByIndex(projectID, parentID, index int) (*Task, error) {
	task := Task{}
	err := d.db.
		Where("project = ? AND parent = ? AND `index` > ?", projectID, parentID, index).
		Order("`index` ASC").
		Take(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &task, err
}

func (d *TasksDAO) getMinDistance(projectID, parentID, index int) (int, error) {
	var toEnd, toStart int64
	err := d.db.Model(&Task{}).
		Where("project = ? AND parent = ? AND `index` < ?", projectID, parentID, index+1).
		Count(&toStart).Error
	if err != nil {
		return 0, err
	}

	if toStart == 0 {
		return -1, nil
	}

	err = d.db.Model(&Task{}).
		Where("project = ? AND parent = ? AND `index` > ?", projectID, parentID, index-1).
		Count(&toEnd).Error
	if err != nil {
		return 0, err
	}

	if toEnd > toStart {
		return -1, nil
	}
	return 1, nil
}

func (d *TasksDAO) updateIndex(projectID, parentID, from, offset int) (dir int, err error) {
	return d.updateIndexTX(projectID, parentID, from, offset, d.db)
}

func (d *TasksDAO) updateIndexTX(projectID, parentID, from, offset int, tx *gorm.DB) (dir int, err error) {
	direction, err := d.getMinDistance(projectID, parentID, from)

	if direction < 0 {
		// update index to the start where 'index' <= 'from'
		err = tx.Model(&Task{}).
			Where("project = ? AND parent = ? AND `index` < ?", projectID, parentID, from+1).
			Update("index", gorm.Expr("`index` - ?", offset)).Error
	} else {
		// update index to the end where 'index' > 'from'
		err = tx.Model(&Task{}).
			Where("project = ? AND parent = ? AND `index` > ?", projectID, parentID, from).
			Update("index", gorm.Expr("`index` + ?", offset)).Error
	}

	return direction, nil
}

func (d *TasksDAO) getChildrenIDs(projectID, taskID int) ([]int, error) {
	arr, err := d.GetFromProject(projectID)
	if err != nil {
		return nil, err
	}

	return findChildren(arr, taskID), nil
}

func findChildren(arr []Task, id int) []int {
	if i := hasChild(arr, id); i == -1 {
		return []int{}
	}

	var storage []int
	for i := range arr {
		if arr[i].ParentID == id {
			storage = append(storage, arr[i].ID)
			storage = append(storage, findChildren(arr, arr[i].ID)...)
		}
	}
	return storage
}

func hasChild(arr []Task, id int) int {
	for i := range arr {
		if arr[i].ParentID == id {
			return i
		}
		i++
	}
	return -1
}

func getIDs(users []User) []int {
	ids := make([]int, len(users))
	for i, card := range users {
		ids[i] = card.ID
	}
	return ids
}
