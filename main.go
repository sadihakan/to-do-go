package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ToDo struct {
	ID          int64  `db:"id" json:"id"`
	Description string `db:"description" json:"description"`
	IsDone      bool   `db:"is_done" json:"is_done"`
}

type Response struct {
	Data       interface{} `json:"data"`
	TotalCount int64       `json:"total_count"`
	Status     bool        `json:"status"`
}

type PostTodo struct {
	Description string `json:"description" form:"description"`
}

type RequestMessage struct {
	Message string `json:"message"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "sadihakan1"
	dbname   = "learning"
)

var db *sqlx.DB
var validate *validator.Validate

func main() {
	connectToDatabase()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	e.GET("/", getTodos)
	e.GET("/:id", getTodosWithID)
	e.POST("/", postTodo)
	e.PATCH("/:id", patchTodo)
	e.DELETE("/:id", deleteTodo)
	e.Logger.Fatal(e.Start(":8081"))
}

func getTodos(c echo.Context) error {
	response := new(Response)
	todos := getTodosFromDB()
	response.Data = todos
	response.TotalCount = int64(len(todos))
	return c.JSON(http.StatusOK, response)
}

func getTodosWithID(c echo.Context) error {
	param := c.Param("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		message := new(RequestMessage)
		message.Message = "Can't find id:"+param
		return c.JSON(http.StatusBadRequest, message)
	}
	todo, isDone := getTodoFromDB(id)
	if isDone {
		return c.JSON(http.StatusOK, todo)
	} else {
		message := new(RequestMessage)
		message.Message = "User not exist"
		return c.JSON(http.StatusBadRequest, message)
	}
}

func postTodo(c echo.Context) (err error) {
	response := new(Response)
	body := new(PostTodo)
	todo := new(ToDo)
	if err = c.Bind(body); err != nil {
		return
	}
	todo.Description = body.Description
	err = insertTodoToDB(todo)
	if err != nil {
		message := new(RequestMessage)
		message.Message = err.Error()
		return c.JSON(http.StatusBadRequest, message)
	}
	response.Data = todo
	response.Status = true
	return c.JSON(http.StatusCreated, response)
}

func patchTodo(c echo.Context) (err error) {
	body := new(ToDo)
	if err = c.Bind(body); err != nil {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	body.ID = int64(id)
	patchTodoDB(body)
	return c.JSON(http.StatusOK, body)
}

func deleteTodo(c echo.Context) error {
	param := c.Param("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusOK, "Can't find id:"+param)
	}
	deleteTodoDB(id)
	message := new(RequestMessage)
	message.Message = "Todo deleted"
	return c.JSON(http.StatusBadRequest, message)
}

func connectToDatabase() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err = sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Successfully connected!!!")
}

func getTodosFromDB() []ToDo {
	rows, err := db.Queryx("SELECT * FROM todos")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var todos []ToDo
	for rows.Next() {
		var tmp ToDo
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

func getTodoFromDB(id int) (ToDo, bool) {
	fmt.Println(id)
	rows, err := db.Queryx("SELECT * FROM todos WHERE id = $1", id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var todos []ToDo
	for rows.Next() {
		var tmp ToDo
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

	validateErr := validate.Struct(todos)
	if validateErr != nil {
		return ToDo{
			ID:          0,
			Description: "",
			IsDone:      false,
		}, false

	}
	return todos[0], true
}

func insertTodoToDB(todo *ToDo) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	err = tx.QueryRowx("INSERT INTO todos (description) VALUES ($1) RETURNING id", todo.Description).Scan(&todo.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func patchTodoDB(todo *ToDo) error {
	tx, err := db.Beginx()
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Println(todo)
	err = tx.QueryRowx("UPDATE todos SET is_done = $1 WHERE id = $2 RETURNING *", todo.IsDone, todo.ID).Scan(&todo.ID,&todo.IsDone,&todo.Description)
	if err != nil {
		log.Fatal(err)
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func deleteTodoDB(id int) {
	tx := db.MustBegin()
	tx.MustExec("DELETE FROM todos WHERE id = $1;", id)
	tx.Commit()
}
