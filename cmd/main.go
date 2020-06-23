package main

import (
	"ToDoGo/app"
	"log"
	"net/http"
	"os"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	a := app.NewApp(dir)
	//log.Fatal(a.Handler.Echo.Start(":8081"))
	log.Fatal(http.ListenAndServe(":8081", a.Handler.Chi))

}




