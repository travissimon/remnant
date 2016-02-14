package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	_ "github.com/lib/pq"
	"github.com/travissimon/remnant/data"
	"github.com/travissimon/remnant/types"
)

func swaggerIndexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "swagger/remnant.swagger.json")
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	// decode log message
	var logMsg types.LogMessage
	err := json.NewDecoder(r.Body).Decode(&logMsg)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	data.InsertLog(logMsg)
}

func spanHandler(w http.ResponseWriter, r *http.Request) {
	// decode log message
	var span types.Span
	err := json.NewDecoder(r.Body).Decode(&span)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	if span.TraceId == span.Id {
		// Initiating span
		data.InsertSpan(span)
	} else {
		mergeSpan(span)
	}
}

var spanCache map[string]types.Span
var spanMutex = &sync.Mutex{}

func prettyPrint(obj interface{}) string {
	json, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling json: %s\n", err.Error())
	}
	return string(json)
}

func writeMismatch(fieldName, existing, new string) {
	fmt.Fprintf(os.Stderr, "Span merge, mismatched field: %s\n", fieldName)
	fmt.Fprintf(os.Stderr, "existing: %s\n", existing)
	fmt.Fprintf(os.Stderr, "new: %s\n", new)
}

func mergeSpan(span types.Span) {
	spanMutex.Lock()
	sp, exists := spanCache[span.Id]
	if !exists {
		// either client or remote half of this span is missing -
		// cache it until we get all the details
		spanCache[span.Id] = span
		spanMutex.Unlock()
		return
	}

	spanMutex.Unlock()

	// remove span from cache
	delete(spanCache, span.Id)

	// copy across missing values
	if sp.TraceId == "" {
		sp.TraceId = span.TraceId
	} else if sp.TraceId != span.TraceId {
		writeMismatch("TraceId", sp.TraceId, span.TraceId)
	}
	if sp.ParentId == "" {
		sp.ParentId = span.ParentId
	} else if span.ParentId != "" && sp.ParentId != span.ParentId {
		writeMismatch("ParentId", sp.ParentId, span.ParentId)
	}
	if sp.ClientStart == "" {
		sp.ClientStart = span.ClientStart
	}
	if sp.ClientEnd == "" {
		sp.ClientEnd = span.ClientEnd
	}
	if sp.RemoteStart == "" {
		sp.RemoteStart = span.RemoteStart
	}
	if sp.RemoteEnd == "" {
		sp.RemoteEnd = span.RemoteEnd
	}
	if sp.Host == "" {
		sp.Host = span.Host
	} else if span.Host != "" && sp.Host != span.Host {
		writeMismatch("Host", sp.Host, span.Host)
	}
	if sp.Method == "" {
		sp.Method = span.Method
	} else if span.Method != "" && sp.Method != span.Method {
		writeMismatch("Method", sp.Method, span.Method)
	}
	if sp.Url == "" {
		sp.Url = span.Url
	} else if span.Url != "" && sp.Url != span.Url {
		writeMismatch("Url", sp.Url, span.Url)
	}
	if sp.Headers == nil || len(sp.Headers) == 0 {
		sp.Headers = span.Headers
	}
	if sp.Parameters == nil || len(sp.Parameters) == 0 {
		sp.Parameters = span.Parameters
	}
	if sp.Body == "" {
		sp.Body = span.Body
	} else if span.Body != "" && sp.Body != span.Body {
		writeMismatch("Body", sp.Body, span.Body)
	}
	if sp.ResponseCode == 0 {
		sp.ResponseCode = span.ResponseCode
	} else if span.ResponseCode != 0 && sp.ResponseCode != span.ResponseCode {
		writeMismatch("ResponseCode", string(sp.ResponseCode), string(span.ResponseCode))
	}

	data.InsertSpan(sp)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func main() {
	var port = flag.String("port", "7777", "Define what TCP port to bind to")
	flag.Parse()

	spanCache = make(map[string]types.Span)

	http.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))
	http.HandleFunc("/v1/log", logHandler)
	http.HandleFunc("/v1/client-span", spanHandler)
	http.HandleFunc("/v1/remote-span", spanHandler)
	http.HandleFunc("/healthz", healthzHandler)

	fmt.Printf("Starting Remnant server on port %s\n", *port)
	http.ListenAndServe(":"+*port, nil)
}
