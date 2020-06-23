package app

import (
	"ToDoGo/model"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/labstack/echo"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Handler struct {
	App *App
	Echo *echo.Echo
	Chi *chi.Mux
}

func NewHandler(app *App) *Handler {
	h := new(Handler)
	e := echo.New()
	c := chi.NewRouter()

	h.App = app

	c.Use(middleware.RequestID)
	c.Use(middleware.RealIP)
	c.Use(middleware.Logger)
	c.Use(middleware.Recoverer)

	c.Route("/", func(r chi.Router) {
		r.Get("/", h.getTodosByChi)
		r.Get("/{id}", h.getTodosWithIDByChi)
		r.Post("/", h.postTodoByChi)
		r.Patch("/{id}", h.patchTodoByChi)
		r.Delete("/{id}", h.deleteTodoByChi)
		r.Post("/file", h.getImagePath)
	})

	h.FileServer(c)

	e.GET("/", h.getTodos)
	e.GET("/:id", h.getTodosWithID)
	e.POST("/", h.postTodo)
	e.PATCH("/:id", h.patchTodo)
	e.DELETE("/:id", h.deleteTodo)

	h.Echo = e
	h.Chi = c
	return h
}
//HTTP

func (h *Handler) FileServer(router *chi.Mux) {
	root := filepath.Join(h.App.Path, "src")
	fs := http.FileServer(http.Dir(root))

	router.Get("/files/*", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		if _, err := os.Stat(root + r.RequestURI); os.IsNotExist(err) {
			http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})
}

func (h *Handler) getImagePath(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(0)
	id, err := strconv.ParseInt(r.FormValue("id"), 10,64)
	if err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	tempFile, err := ioutil.TempFile("./src/files/todo", "todo-*.jpg")

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}

	_, err = tempFile.Write(fileBytes)
	if err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}

	todo := new(model.Todo)
	todo.ID = id

	err = h.addImagePathDB(todo, tempFile.Name())
	if err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}

	h.renderJSON(w, http.StatusOK, todo)
}

func (h *Handler) renderJSON(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) getTodosByChi(w http.ResponseWriter, r *http.Request) {
	response := new(model.Response)
	todos := h.getTodosFromDB()
	response.Data = todos
	response.TotalCount = int64(len(todos))
	h.renderJSON(w,http.StatusOK, response)
}

func (h *Handler) getTodos(c echo.Context) error {
	response := new(model.Response)
	todos := h.getTodosFromDB()
	response.Data = todos
	response.TotalCount = int64(len(todos))
	return c.JSON(http.StatusOK, response)
}

func (h *Handler) getTodosWithIDByChi(w http.ResponseWriter, r *http.Request) {
	query := chi.URLParam(r, "id")
	if query != "" {
		id, err := strconv.ParseInt(query, 10,64)
		if err != nil {
			h.renderJSON(w, http.StatusBadRequest, nil)
		}
		todo, err := h.getTodoFromDB(id)
		if err != nil {
			h.renderJSON(w, http.StatusBadRequest, nil)
		}

		h.renderJSON(w, http.StatusOK, model.Response{
			Data: todo,
		})
	}
}

func (h *Handler) getTodosWithID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10,64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, nil)
	}
	todo, err := h.getTodoFromDB(id)
	if err != nil {
		return c.JSON(http.StatusNotFound,nil)
	}

	return c.JSON(http.StatusOK, model.Response{
		Data: todo,
	})
}

func (h *Handler) postTodoByChi(w http.ResponseWriter, r *http.Request) {
	response := new(model.Response)
	todo := new(model.Todo)
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		h.renderJSON(w, http.StatusBadRequest, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusBadRequest),
		})
		return
	}

	if err = h.App.Validator.Struct(todo); err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}

	err = h.insertTodoToDB(todo)
	if err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
	}
	fmt.Printf("Here")
	response.Data = todo
	h.renderJSON(w, http.StatusCreated, response)
}

func (h *Handler) postTodo(c echo.Context) (err error) {
	response := new(model.Response)
	todo := new(model.Todo)
	if err = c.Bind(todo); err != nil {
		return err
	}

	if err = h.App.Validator.Struct(todo); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
	}

	err = h.insertTodoToDB(todo)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
	}

	response.Data = todo
	return c.JSON(http.StatusCreated, response)
}

func (h *Handler) patchTodoByChi(w http.ResponseWriter, r *http.Request) {
	todo := model.NewTodo()
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
	}

	query := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(query, 10,64)
	if err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
	}

	todo.ID = id
	h.patchTodoDB(todo.ID, todo)
	h.renderJSON(w, http.StatusOK, todo)
}

func (h *Handler) patchTodo(c echo.Context) (err error) {
	todo := model.NewTodo()
	if err = c.Bind(todo); err != nil {
		return
	}
	if err = h.App.Validator.Struct(todo); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
	}
	id, err := strconv.Atoi(c.Param("id"))
	todo.ID = int64(id)
	h.patchTodoDB(todo.ID, todo)
	return c.JSON(http.StatusOK, todo)
}

func (h *Handler) deleteTodoByChi(w http.ResponseWriter, r *http.Request) {
	query := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(query, 10,64)
	if err != nil {
		h.renderJSON(w, http.StatusUnprocessableEntity, model.Response{
			Errors: err.Error(),
			Detail: http.StatusText(http.StatusUnprocessableEntity),
		})
	}

	err = h.deleteTodoDB(id)
	if err != nil {
		h.renderJSON(w, http.StatusNotFound, model.Response{
			Detail: http.StatusText(http.StatusNotFound),
		})
	}

	h.renderJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) deleteTodo(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10,64)
	if err != nil {
		return err
	}

	err = h.deleteTodoDB(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, model.Response{
			Detail: http.StatusText(http.StatusNotFound),
		})
	}

	return c.JSON(http.StatusNoContent, nil)
}

//MARK: Database
func (h *Handler) getTodosFromDB() []model.Todo {
	rows, err := h.App.DB.Queryx("SELECT * FROM todos")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var todos []model.Todo
	for rows.Next() {
		var tmp model.Todo
		err := rows.StructScan(&tmp)
		if err != nil {
			log.Fatal(err)
		}
		todos = append(todos, tmp)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return todos
}

func (h *Handler) getTodoFromDB(id int64) (*model.Todo, error) {
	todo := model.NewTodo()
	err := h.App.DB.QueryRowx("SELECT * FROM todos WHERE id = $1", id).StructScan(todo)
	if err != nil {
		return nil, err
	}

	return todo, nil
}

func (h *Handler) insertTodoToDB(todo *model.Todo) error {
	tx, err := h.App.DB.Beginx()
	if err != nil {
		return err
	}
	err = tx.QueryRowx("INSERT INTO todos (description) VALUES ($1) RETURNING id, is_done", todo.Description).Scan(&todo.ID, &todo.IsDone)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (h *Handler) patchTodoDB(id int64, todo *model.Todo) error {
	tx, err := h.App.DB.Beginx()
	if err != nil {
		return err
	}

	err = tx.QueryRowx("UPDATE todos SET is_done = $1 WHERE id = $2 RETURNING *", todo.IsDone, id).StructScan(todo)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (h *Handler) deleteTodoDB(id int64) error {
	tx, err := h.App.DB.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (h *Handler) addImagePathDB(todo *model.Todo, path string) error {
	tx, err := h.App.DB.Beginx()
	if err != nil {
		return err
	}

	err = tx.QueryRowx("UPDATE todos SET image_url = $1 WHERE id = $2 RETURNING *", path, todo.ID).StructScan(todo)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}




