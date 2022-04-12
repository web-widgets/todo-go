package main

import (
	"net/http"
	"web-widgets/todo-go/data"

	"github.com/go-chi/chi"
)

func initRoutes(r chi.Router, dao *data.DAO) {

	r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) {
		data, err := dao.Tasks.GetAll()
		sendResponse(w, data, err)
	})

	r.Get("/tasks/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		data, err := dao.Tasks.GetFromProject(id)
		sendResponse(w, data, err)
	})

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		data, err := dao.Users.GetAll()
		sendResponse(w, data, err)
	})

	r.Get("/projects", func(w http.ResponseWriter, r *http.Request) {
		data, err := dao.Projects.GetAll()
		sendResponse(w, data, err)
	})

	r.Post("/tasks", func(w http.ResponseWriter, r *http.Request) {
		task := data.TaskUpdate{}
		err := parseFormObject(w, r, &task)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}
		id, err := dao.Tasks.Add(&task)
		sendResponse(w, &Response{id}, err)
	})

	r.Put("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		task := data.TaskUpdate{}
		err := parseFormObject(w, r, &task)
		id := numberParam(r, "id")
		if err == nil {
			err = dao.Tasks.Update(id, &task)
		}
		sendResponse(w, nil, err)
	})

	r.Delete("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		err := dao.Tasks.Delete(id)
		sendResponse(w, nil, err)
	})

	r.Post("/projects", func(w http.ResponseWriter, r *http.Request) {
		project := data.ProjectUpdate{}
		err := parseFormObject(w, r, &project)
		var id int
		if err == nil {
			id, err = dao.Projects.Add(&project)
		}
		sendResponse(w, &Response{id}, err)
	})

	r.Put("/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		project := data.ProjectUpdate{}
		err := parseFormObject(w, r, &project)
		if err == nil {
			err = dao.Projects.Update(id, &project)
		}
		sendResponse(w, nil, err)
	})

	r.Delete("/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		err := dao.Projects.Delete(id)
		sendResponse(w, nil, err)
	})

	r.Put("/move/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := numberParam(r, "id")
		info := data.MoveInfo{}
		err := parseFormObject(w, r, &info)
		if err == nil {
			err = dao.Tasks.Move(id, &info)
		}
		sendResponse(w, nil, err)
	})

	r.Get("/tags", func(w http.ResponseWriter, r *http.Request) {
		data, err := dao.Tags.GetAll()
		sendResponse(w, data, err)
	})

	r.Post("/clone", func(w http.ResponseWriter, r *http.Request) {
		info := data.PasteInfo{}
		err := parseFormObject(w, r, &info)
		var pull map[string]int
		if err == nil {
			pull, err = dao.Tasks.Paste(&info)
		}
		sendResponse(w, pull, err)
	})

}

func sendResponse(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		format.Text(w, 500, err.Error())
	} else {
		if data == nil {
			data = Response{}
		}
		format.JSON(w, 200, data)
	}
}
