package remnant

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
)

// timestamps should be formatted according to rfc3339:
// http://tools.ietf.org/html/rfc3339
// e.g. 2006-01-02T15:04:05Z07:00

type KeyValue struct {
	Key   string `json:key`
	Value string `json:value`
}

type LogMessage struct {
	TraceId    string `json:traceId`
	SpanId     string `json:spanId`
	Severity   string `json:severity`
	Message    string `json:message`
	StackTrace string `json:stackTrace`
}

type LocalSpan struct {
	TraceId    string     `json:traceId`
	Id         string     `json:spanId`
	ParentId   string     `json:parentSpanId`
	Start      string     `json:startTimestamp`
	End        string     `json:endTimestamp`
	Host       string     `json:host`
	Method     string     `json:method`
	Url        string     `json:rurl`
	Headers    []KeyValue `json:headers`
	Parameters []KeyValue `json:parameters`
	Body       string     `json:body`
}

type RemoteSpan struct {
	TraceId      string `json:traceId`
	Id           string `json:spanId`
	ParentId     string `json:parentSpanId`
	Start        string `json:startTimestamp`
	End          string `json:endTimestamp`
	ResponseCode int    `json:responseCode`
}

func swaggerIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Remnant Index Handler %s\n", time.Now().Local())
	http.ServeFile(w, r, "swagger/remnant.swagger.json")
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Remnant Log handler: %s", time.Now().UTC().Format(time.RFC3339))
	// decode log message
	var logMsg LogMessage
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
	fmt.Printf("Remnant Client span: %s", time.Now().UTC().Format(time.RFC3339))
	// decode log message
	var localSpan LocalSpan
	err := json.NewDecoder(r.Body).Decode(&localSpan)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	// Do something with it?
	fmt.Printf("Message: %s\n", localSpan)
	fmt.Fprintf(w, "Message: %s\n", localSpan)
}

func remoteSpanHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Remnant Remote span: %s", time.Now().UTC().Format(time.RFC3339))
	// decode log message
	var remoteSpan RemoteSpan
	err := json.NewDecoder(r.Body).Decode(&remoteSpan)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	// Do something with it?
	fmt.Printf("Message: %s\n", remoteSpan)
	fmt.Fprintf(w, "Message: %s\n", remoteSpan)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Remnant healthz %s\n", time.Now().Local())
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
	http.HandleFunc("/", swaggerIndexHandler)

	fmt.Printf("Starting combined mode Remnant server on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
