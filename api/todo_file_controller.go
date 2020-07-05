package api

import (
	"ToDoGo/model"
	"fmt"
	"github.com/go-chi/chi"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"github.com/google/uuid"
)

type TodoFileController struct {
	*Api
}

func (c TodoFileController) Index(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "todoID"), 10,64)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusBadRequest, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusBadRequest),
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

	c.Handler.renderJSON(w,http.StatusOK, &model.Response{
		Data: todoFiles,
		TotalCount: int64(len(todoFiles)),
	})
}

func (c TodoFileController) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(0)
	id, err := strconv.ParseInt(chi.URLParam(r, "todoID"), 10,64)
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
		c.Handler.renderJSON(w, http.StatusNotFound, nil)
		return
	}

	f, handler, err := r.FormFile("file")
	if err != nil {
		c.Handler.renderJSON(w, http.StatusUnprocessableEntity, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}
	defer f.Close()

	uid := uuid.New()

	os.MkdirAll(filepath.Join(c.Path, "files", "todo", fmt.Sprintf("%d", todo.ID), uid.String()), os.ModePerm)

	createdFile, err := os.Create(filepath.Join(c.Path, "files", "todo", fmt.Sprintf("%d", todo.ID), uid.String(), handler.Filename))
	if err != nil {
		c.Handler.renderJSON(w, http.StatusInternalServerError, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	_, err = io.Copy(createdFile, f)

	if err != nil {
		c.Handler.renderJSON(w, http.StatusInternalServerError, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	todoFile := model.NewTodoFile()
	err = c.DB.QueryRowx("INSERT INTO todo_files (todo_id, path) VALUES ($1, $2) RETURNING *", todo.ID, filepath.Join("todo", fmt.Sprintf("%d", todo.ID), uid.String(), handler.Filename)).StructScan(todoFile)
	if err != nil {
		os.Remove(createdFile.Name())
		c.Handler.renderJSON(w, http.StatusUnprocessableEntity, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}

	c.Handler.renderJSON(w, http.StatusOK, &model.Response{
		Data: todoFile,
	})
}

func (c TodoFileController) Delete(w http.ResponseWriter, r *http.Request) {
	todoID, err := strconv.ParseInt(chi.URLParam(r, "todoID"), 10,64)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusBadRequest, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusBadRequest),
		})
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10,64)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusBadRequest, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusBadRequest),
		})
		return
	}

	todoFile := model.NewTodoFile()
	err = c.DB.QueryRowx("SELECT * FROM todo_files WHERE id = $1 AND todo_id = $2", id, todoID).StructScan(todoFile)
	if err != nil {
		c.Handler.renderJSON(w, http.StatusNotFound, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusBadRequest),
		})
		return
	}

	if err := os.Remove(filepath.Join(c.Path, todoFile.Path)); err != nil {
		c.Handler.renderJSON(w, http.StatusInternalServerError, &model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	if _, err = c.DB.Exec("DELETE FROM todo_files WHERE id = $1", todoFile.ID); err != nil {
		c.Handler.renderJSON(w, http.StatusNotFound, &model.Response{
			Detail: http.StatusText(http.StatusNotFound),
		})
		return
	}

	c.Handler.renderJSON(w, http.StatusNoContent, nil)
}