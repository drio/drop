package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

//go:embed templates
var templateFS embed.FS

func main() {
	appName := "drop"
	dbPath := "drop.sqlite"
	port := "9191"
	domain := "driohq.net"
	inFly := false
	if _, ok := os.LookupEnv("FLY_APP_NAME"); ok {
		dbPath = "/data/drop.sqlite"
		inFly = true
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath))
	exitOnError(err)
	model, err := NewSQLModel(db)
	exitOnError(err)

	server, err := NewServer(ServerOpts{
		model:   model,
		logger:  log.Default(),
		appName: appName,
		inFly:   inFly,
		port:    port,
		domain:  domain,
	})
	exitOnError(err)

	fmt.Printf("listening :%s", server.port)
	exitOnError(http.ListenAndServe("0.0.0.0:9191", server))
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
