package api

import (
	"ToDoGo/model"
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type TodoController struct {
	*Api
}

func (c TodoController) Index(w http.ResponseWriter, r *http.Request) {
	todos := make([]*model.Todo, 0)
	todoIDs := make([]int64, 0)
	todoFilesMap := make(map[int64][]model.TodoFile)

	rows, err := c.DB.Queryx("SELECT * FROM todos")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tmp model.Todo
			if err := rows.StructScan(&tmp); err == nil {
				todos = append(todos, &tmp)
				todoIDs = append(todoIDs, tmp.ID)
			}
		}
	}

	if len(todoIDs) > 0 {
		query, args, err := sqlx.In("SELECT * FROM todo_files WHERE todo_id IN (?);", todoIDs)
		if err == nil {
			query = c.DB.Rebind(query)
			rows, err := c.DB.Queryx(query, args...)

			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var tmp model.TodoFile
					if err := rows.StructScan(&tmp); err == nil {
						todoFilesMap[tmp.TodoID] = append(todoFilesMap[tmp.TodoID], tmp)
					}
				}
			}
		}
	}

	for _, todo := range todos {
		todo.Files = make([]model.TodoFile, 0)
		if files, ok := todoFilesMap[todo.ID]; ok {
			todo.Files = files
		}
	}

	c.Handler.renderJSON(w,http.StatusOK, &model.Response{
		Data: todos,
		TotalCount: int64(len(todos)),
	})
}

func (c TodoController) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusBadRequest, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusBadRequest),
		})
		return
	}

	todo := model.NewTodo()
	err = c.DB.QueryRowx("SELECT * FROM todos WHERE id = $1", id).StructScan(todo)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusNotFound, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusNotFound),
		})
		return
	}

	c.Handler.renderJSON(w, http.StatusOK, &model.Response{
		Data: todo,
	})
}

func (c TodoController) Create(w http.ResponseWriter, r *http.Request) {
	todo := new(model.Todo)

	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		c.Handler.renderJSON(w, http.StatusBadRequest, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusBadRequest),
		})
		return
	}

	if err := c.Validator.Struct(todo); err != nil {
		c.Handler.renderJSON(w, http.StatusUnprocessableEntity, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}

	err := c.DB.QueryRowx("INSERT INTO todos (description) VALUES ($1) RETURNING id, is_done", todo.Description).Scan(&todo.ID, &todo.IsDone)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusUnprocessableEntity, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}

	c.Handler.renderJSON(w, http.StatusCreated, &model.Response{
		Data: todo,
	})
}

func (c TodoController) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusBadRequest, nil)
		return
	}

	todo := model.NewTodo()

	if err = c.DB.QueryRowx("SELECT * FROM todos WHERE id = $1", id).StructScan(todo); err != nil {
		c.Handler.renderJSON(w, http.StatusNotFound, nil)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		c.Handler.renderJSON(w, http.StatusUnprocessableEntity, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
	}

	err = c.DB.QueryRowx("UPDATE todos SET is_done = $1 WHERE id = $2 RETURNING *", todo.IsDone, id).StructScan(todo)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusUnprocessableEntity, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}

	c.Handler.renderJSON(w, http.StatusOK, &model.Response{
		Data: todo,
	})
}

func (c TodoController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusBadRequest, nil)
		return
	}

	todo := model.NewTodo()

	if err = c.DB.QueryRowx("SELECT * FROM todos WHERE id = $1", id).StructScan(todo); err != nil {
		c.Handler.renderJSON(w, http.StatusNotFound, &model.Response{
			Detail: http.StatusText(http.StatusNotFound),
		})
		return
	}

	todoFiles := make([]model.TodoFile, 0)
	rows, err := c.DB.Queryx("SELECT * FROM todo_files WHERE todo_id = $1", id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tmp model.TodoFile
			if err := rows.StructScan(&tmp); err == nil {
				todoFiles = append(todoFiles, tmp)
			}
		}
	}

	for _, f := range todoFiles {
		os.Remove(filepath.Join(c.Path, f.Path))
	}

	if _, err = c.DB.Exec("DELETE FROM todos WHERE id = $1", id); err != nil {
		c.Handler.renderJSON(w, http.StatusNotFound, &model.Response{
			Detail: http.StatusText(http.StatusNotFound),
		})
		return
	}

	c.Handler.renderJSON(w, http.StatusNoContent, nil)
}

