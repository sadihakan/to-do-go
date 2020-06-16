package app

import (
	"ToDoGo/model"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	App *App
	Echo *echo.Echo
}

func NewHandler(app *App) *Handler {
	h := new(Handler)
	h.App = app
	e := echo.New()

	e.GET("/", h.getTodos)
	e.GET("/:id", h.getTodosWithID)
	e.POST("/", h.postTodo)
	e.PATCH("/:id", h.patchTodo)
	e.DELETE("/:id", h.deleteTodo)

	h.Echo = e

	return h
}

func (h *Handler) getTodos(c echo.Context) error {
	response := new(model.Response)
	todos := h.getTodosFromDB()
	response.Data = todos
	response.TotalCount = int64(len(todos))
	return c.JSON(http.StatusOK, response)
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




