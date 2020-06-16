package main

import (
	"ToDoGo/app"
	"log"
)

func main() {
	a := app.NewApp()
	log.Fatal(a.Handler.Echo.Start(":8081"))
}




