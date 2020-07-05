package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Api struct {
	DB *sqlx.DB
	Handler *Handler
	Validator *validator.Validate
	Path string
}

func NewApi(dir string, db *sqlx.DB ) *Api {
	a := new(Api)
	a.Validator = validator.New()
	a.DB = db
	a.Handler = NewHandler(a)
	a.Path = dir

	return a
}
