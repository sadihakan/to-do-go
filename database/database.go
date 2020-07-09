package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

var (
	host     = ""
	port     = 0
	user     = ""
	password = ""
	dbname   = ""
)

func Connect() *sqlx.DB {
	setSecrets()
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		log.Fatalln(err)
	}
	db.SetMaxIdleConns(15)
	db.SetMaxOpenConns(15)

	return db
}

func setSecrets() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}


	host = os.Getenv("HOST")
	port, _ = strconv.Atoi(os.Getenv("PORT"))
	user = os.Getenv("DB_USER")
	password = os.Getenv("PASSWORD")
	dbname = os.Getenv("DB_NAME")

	fmt.Println(host,port,password,user,dbname)
}


