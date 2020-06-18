package main

import (
	"ToDoGo/app"
	"log"
	"net/http"
)

func main() {
	a := app.NewApp()
	//log.Fatal(a.Handler.Echo.Start(":8081"))
	log.Fatal(http.ListenAndServe(":8081", a.Handler.Chi))

}




