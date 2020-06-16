package app

import (
	"ToDoGo/database"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type App struct {
	DB *sqlx.DB
	Handler *Handler
	Validator *validator.Validate
}

func NewApp() *App {
	a := new(App)
	a.Validator = validator.New()
	a.DB = database.Connect()
	a.Handler = NewHandler(a)

	return a
}
