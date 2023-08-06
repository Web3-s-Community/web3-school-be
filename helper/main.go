package main

import (
	"autopilot-helper/helper/handler"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
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

	demoHandler := handler.DemoHandler{}

	http.HandleFunc("/", demoHandler.Handle)

	port := 8092
	fmt.Printf("Start serving HTTP at port: %v\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		fmt.Println("Error when server HTTP: " + err.Error())
	}
}