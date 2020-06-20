package model

type Todo struct {
	ID          int64  `db:"id" json:"id"`
	Description string `db:"description" json:"description" validate:"required"`
	IsDone      bool   `db:"is_done" json:"is_done"`
	ImageURL  	string `db:"image_url" json:"imageURL"`
}

func NewTodo() *Todo {
	return &Todo{}
}
