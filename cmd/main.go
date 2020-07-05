package main

import (
	"ToDoGo/api"
	"ToDoGo/database"
	"log"
	"net/http"
	"os"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	db := database.Connect()
	a := api.NewApi(dir, db)
	//log.Fatal(a.Handler.Echo.Start(":8081"))
	log.Fatal(http.ListenAndServe(":8081", a.Handler.Chi))

}




