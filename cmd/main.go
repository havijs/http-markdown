package main

import (
	"log"
	"net/http"

	"github.com/nvdsalehi/http-markdown/internal/app"
)

func main() {
	app := app.GetApp()
	app.Init()
	log.Fatal(http.ListenAndServe(":8080", app.Router))
}
