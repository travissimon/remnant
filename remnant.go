package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/travissimon/remnant/client"
)

func swaggerIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s Remnant Index Handler\n", time.Now().Local())
	fmt.Printf("Request url: %s\n", r.URL)
	http.ServeFile(w, r, "swagger/remnant.swagger.json")
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s Remnant Log handler", time.Now().UTC().Format(time.RFC3339))
	// decode log message
	var logMsg client.LogMessage
	err := json.NewDecoder(r.Body).Decode(&logMsg)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	// Do something with it?
	fmt.Printf("Message: %s\n", logMsg)
	fmt.Fprintf(w, "Message: %s\n", logMsg)
}

func clientSpanHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s Remnant Client span\n", time.Now().UTC().Format(time.RFC3339))
	// decode log message
	var localSpan client.LocalSpan
	err := json.NewDecoder(r.Body).Decode(&localSpan)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	// Do something with it?
	fmt.Printf("cl span: %s\n", localSpan)
	fmt.Fprintf(w, "cl span: %s\n", localSpan)
}

func remoteSpanHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s Remnant Remote span\n", time.Now().UTC().Format(time.RFC3339))
	// decode log message
	var remoteSpan client.RemoteSpan
	err := json.NewDecoder(r.Body).Decode(&remoteSpan)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	// Do something with it?
	fmt.Printf("rm span: %s\n", remoteSpan)
	fmt.Fprintf(w, "rm span: %s\n", remoteSpan)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s Remnant healthz\n", time.Now().Local())
	fmt.Fprintf(w, "OK")
}

var db *sql.DB

func initServer() error {
	database, err := sql.Open("postgres", "host=localhost dbname=tsimon sslmode=disable")
	if err != nil {
		log.Fatal("Error connecting to database\n")
		return err
	}

	db = database

	return nil
}

func main() {
	var port = flag.String("port", "7777", "Define what TCP port to bind to")
	flag.Parse()

	err := initServer()
	if err != nil {
		log.Fatal("Error initialising server")
		return
	}
	http.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))
	http.HandleFunc("/v1/log", logHandler)
	http.HandleFunc("/v1/client-span", clientSpanHandler)
	http.HandleFunc("/v1/remote-span", remoteSpanHandler)
	http.HandleFunc("/healthz", healthzHandler)

	fmt.Printf("Starting combined mode Remnant server on port %s\n", *port)
	http.ListenAndServe(":"+*port, nil)
}
