package model

type TodoFile struct {
	ID     int64 `db:"id" json:"id" validate:"required"`
	TodoID int64 `db:"todo_id" json:"todo_id" validate:"required"`
	Path   string  `db:"path" json:"path" validate:"required"`
}

func NewTodoFile() *TodoFile {
	return  &TodoFile{}
}
