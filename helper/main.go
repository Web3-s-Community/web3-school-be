package main

import (
	"autopilot-helper/helper/handler"
	"autopilot-helper/helper/pkg/model"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file: " + err.Error())
	}
}

func main() {
	startService()
}
func startService() {
	fmt.Println("Starting service")
	dbConnStr := os.Getenv("DB_CONNSTR")
	if dbConnStr == "" {
		log.Fatalln("Cannot get th_connstr")
	}
	db, err := model.CreateConnectionWithConnString(dbConnStr)
	if err != nil {
		panic(1)
	}
	dbManager := model.NewDbManagerWDB(db)

	demoHandler := handler.DemoHandler{}
	challengesHandler := handler.ChallengesHandler{
		DbManager: dbManager,
	}
	challengeDetailHandler := handler.ChallengeDetailHandler{
		DbManager: dbManager,
	}

	http.HandleFunc("/", demoHandler.Handle)
	http.HandleFunc("/challenges", challengesHandler.Handle)
	http.HandleFunc("/challenge", challengeDetailHandler.Handle)

	port := 8092
	fmt.Printf("Start serving HTTP at port: %v\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		fmt.Println("Error when server HTTP: " + err.Error())
	}
}
