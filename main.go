package main

import (
	"fmt"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

type ToDo struct {
	ID int `db:"id"`
	Description string `db:"description"`
	IsDone bool `db:"is_done"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "sadihakan1"
	dbname   = "learning"
)

var db *sqlx.DB

func main() {
	connectToDatabase()

	e := echo.New()
	e.GET("/", getTodos)
	e.GET("/:id", getTodosWithID)
	e.POST("/", postTodo)
	e.PATCH("/:id", patchTodo)
	e.DELETE("/:id", deleteTodo)
	e.Logger.Fatal(e.Start(":8080"))
}

func getTodos(c echo.Context) error {
	todos := getTodosFromDB()
	return c.JSON(http.StatusOK, todos)
}

func getTodosWithID(c echo.Context) error {
	param := c.Param("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusOK, "Can't find id:" + param)
	}
	todo := getTodoFromDB(id)
	return c.JSON(http.StatusOK, todo)
}

func postTodo(c echo.Context) error {
	description := c.FormValue("description")
	insertTodoToDB(description)
	return c.String(http.StatusOK, "Done")
}

func patchTodo(c echo.Context) error {
	param := c.Param("id")
	isDone := c.FormValue("isDone")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusOK, "Can't find id:" + param)
	}
	isDoneBool, err := strconv.ParseBool(isDone)
	patchTodoDB(id,isDoneBool)
	return c.String(http.StatusOK, "Done")
}

func deleteTodo(c echo.Context) error {
	param := c.Param("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusOK, "Can't find id:" + param)
	}
	deleteTodoDB(id)
	return c.String(http.StatusOK, "Done")
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

func getTodoFromDB(id int) ToDo {
	fmt.Println(id)
	//Neden $1 şeklinde $0 değil ???
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
	return todos[0]
}

func insertTodoToDB(description string) {
	tx := db.MustBegin()
	tx.MustExec("INSERT INTO todos (description) VALUES ($1)", description)
	tx.Commit()
}

func patchTodoDB(id int, isDone bool) {
	tx := db.MustBegin()
	tx.MustExec("UPDATE todos SET is_done = $1 WHERE id = $2", isDone, id)
	tx.Commit()
}

func deleteTodoDB(id int) {
	tx := db.MustBegin()
	tx.MustExec("DELETE FROM todos WHERE id = $1;", id)
	tx.Commit()
}

