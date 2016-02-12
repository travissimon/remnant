package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
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

type headerField string

const (
	traceId      string = "remnant-trace-id"
	spanId              = "remnant-span-id"
	parentSpanId        = "remnant-parent-span-id"
)

type RemnantClient struct {
	remnantUrl string
	span       *LocalSpan
}

func (rc *RemnantClient) getLogUrl() string {
	return rc.remnantUrl + "/v1/log"
}

func (rc *RemnantClient) getLocalServiceUrl() string {
	return rc.remnantUrl + "/v1/client-span"
}

func (rc *RemnantClient) getRemoteRequestUrl() string {
	return rc.remnantUrl + "/v1/remote-span"
}

func NewRemnantClient(remnantUrl string, req *http.Request) (*RemnantClient, error) {
	rand.Seed(time.Now().UnixNano())

	cl := &RemnantClient{}
	cl.span = &LocalSpan{}

	cl.remnantUrl = remnantUrl
	if strings.HasSuffix(remnantUrl, "/") {
		cl.remnantUrl = remnantUrl[:len(remnantUrl)-1]
	}

	cl.span.TraceId = req.Header.Get(traceId)
	cl.span.Id = req.Header.Get(spanId)
	cl.span.ParentId = req.Header.Get(parentSpanId)

	cl.span.Start = getTimestamp()

	req.ParseForm()
	cl.span.Host = req.Host
	cl.span.Method = req.Method
	cl.span.Url = req.URL.String()
	headers := make([]KeyValue, 0, 10)
	for k, v := range req.Header {
		headers = append(headers, KeyValue{Key: k, Value: strings.Join(v, ",")})
	}
	cl.span.Headers = headers
	params := make([]KeyValue, 0, 10)
	for k, v := range req.PostForm {
		params = append(params, KeyValue{Key: k, Value: strings.Join(v, ",")})
	}
	cl.span.Parameters = params

	var body []byte
	var err error
	if body, err = ioutil.ReadAll(req.Body); err != nil {
		return nil, err
	}
	cl.span.Body = string(body)

	return cl, nil
}

func (cl *RemnantClient) EndSpan() {
	cl.span.End = getTimestamp()
	spanJson, err := json.Marshal(cl.span)
	if err != nil {
		fmt.Printf("Error marshalling span: %s\n", err.Error)
		return
	}
	go http.Post(cl.getLocalServiceUrl(), "application/json", bytes.NewReader(spanJson))
}

// generates rfc3339 formated timestrings
func getTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// rand str impl - see http://stackoverflow.com/questions/22892120
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// generates a span id - 32 random alpha-numeric chars
func generateId() string {
	size := 32
	b := make([]byte, size)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := size-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func addRequestHeaders(rc *RemnantClient, req *http.Request, spanId string) {
	req.Header.Add(traceId, rc.span.TraceId)
	req.Header.Add(spanId, spanId)
	req.Header.Add(parentSpanId, rc.span.Id)
}

func (rc *RemnantClient) LogDebug(msg string) {
	rc.log(msg, "", "debug")
}
func (rc *RemnantClient) LogInfo(msg string) {
	rc.log(msg, "", "info")
}
func (rc *RemnantClient) LogWarning(msg string) {
	rc.log(msg, "", "warning")
}
func (rc *RemnantClient) LogError(msg string) {
	rc.log(msg, "", "error")
}
func (rc *RemnantClient) LogException(msg string) {
	rc.log(msg, "", "exception")
}

func (rc *RemnantClient) log(msg, stackTrace, level string) {
	logMsg := &LogMessage{}
	logMsg.TraceId = rc.span.TraceId
	logMsg.SpanId = rc.span.Id
	logMsg.Severity = level
	logMsg.Message = msg
	logMsg.StackTrace = stackTrace

	msgJson, err := json.Marshal(logMsg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling log message: %s\n", err.Error())
	}
	go http.Post(rc.getLogUrl(), "application/json", bytes.NewReader(msgJson))
}

func (rc *RemnantClient) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// mark current client start
	spanId := generateId()

	span := &RemoteSpan{}
	span.TraceId = rc.span.TraceId
	span.Id = spanId
	span.ParentId = rc.span.Id
	span.Start = getTimestamp()

	addRequestHeaders(rc, req, spanId)
	resp, err = http.Get(url)
	span.End = getTimestamp()

	spanJson, err := json.Marshal(span)

	fmt.Printf("calling to %s\n", rc.getRemoteRequestUrl())
	go http.Post(rc.getRemoteRequestUrl(), "application/json", bytes.NewReader(spanJson))

	return resp, err
}
