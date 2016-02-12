package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/travissimon/remnant"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type headerField string

const (
	traceId      string = "remnant-trace-id"
	spanId              = "remnant-span-id"
	parentSpanId        = "remnant-parent-span-id"
)

type RemnantClient struct {
	remnantUrl string
	span       *remnant.LocalSpan
}

func (rc *RemnantClient) getLocalServiceUrl() string {
	return rc.remnantUrl + "/local-span"
}

func (rc *RemnantClient) getRemoteRequestUrl() string {
	return rc.remnantUrl + "/remote-span"
}

func NewRemnantClient(remnantUrl string, req *http.Request) (*RemnantClient, error) {
	rand.Seed(time.Now().UnixNano())

	cl := &RemnantClient{}
	cl.span = &remnant.LocalSpan{}

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
	headers := make([]remnant.KeyValue, 0, 10)
	for k, v := range req.Header {
		headers = append(headers, remnant.KeyValue{Key: k, Value: strings.Join(v, ",")})
	}
	cl.span.Headers = headers
	params := make([]remnant.KeyValue, 0, 10)
	for k, v := range req.PostForm {
		params = append(params, remnant.KeyValue{Key: k, Value: strings.Join(v, ",")})
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
	fmt.Printf("%s %s: %s\n", getTimestamp(), level, msg)
}

func (rc *RemnantClient) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// mark current client start
	spanId := generateId()

	span := &remnant.RemoteSpan{}
	span.TraceId = rc.span.TraceId
	span.Id = spanId
	span.ParentId = rc.span.Id
	span.Start = getTimestamp()

	addRequestHeaders(rc, req, spanId)
	resp, err = http.Get(url)
	span.End = getTimestamp()

	spanJson, err := json.Marshal(span)
	go http.Post(rc.getRemoteRequestUrl(), "application/json", bytes.NewReader(spanJson))

	return resp, err
}
