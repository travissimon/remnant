package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LogMessage struct {
	TraceId    string `json:traceId`
	SpanId     string `json:spanId`
	Severity   string `json:severity`
	Message    string `json:message`
	StackTrace string `json:stackTrace`
}

type HttpHeaderValue struct {
	Key   string `json:key`
	Value string `json:value`
}

type ClientStart struct {
	TraceId           string            `json:traceId`
	SpanId            string            `json:spanId`
	ParentSpanId      string            `json:parentSpanId`
	IpAddress         string            `json:ipAddress`
	RequestUrl        string            `json:requestUrl`
	RequestParameters []HttpHeaderValue `json:requestParameters`
}

type SpanReference struct {
	TraceId      string `json:traceId`
	SpanId       string `json:spanId`
	parentSpanId string `json:parentSpanId`
}

type RemoteEnd struct {
	TraceId          string `json:traceId`
	SpanId           string `json:spanId`
	ParentSpanId     string `json:parentSpanId`
	HttpResponseCode int    `json:httpResponseCode`
}

func swaggerIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Index Handler %s\n", time.Now().Local())
	http.ServeFile(w, r, "swagger/remnant.swagger.json")
}

func logHandler(w http.ResponseWriter, r *http.Request) {
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
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("healthz %s\n", time.Now().Local())
	fmt.Fprintf(w, "OK")
}

func main() {
	http.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))
	http.HandleFunc("/v1/log", logHandler)
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/", swaggerIndexHandler)

	fmt.Printf("Starting combined mode Remnant server on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
