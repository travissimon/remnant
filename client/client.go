package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/travissimon/remnant/types"
)

// timestamps should be formatted according to rfc3339:
// http://tools.ietf.org/html/rfc3339
// e.g. 2006-01-02T15:04:05Z07:00

type headerField string

const (
	traceIdHeader      string = "remnant-trace-id"
	spanIdHeader              = "remnant-span-id"
	parentSpanIdHeader        = "remnant-parent-span-id"
)

type Client interface {
	LogDebug(msg string)
	LogInfo(msg string)
	LogWarning(msg string)
	LogError(msg string)
	LogException(msg, exception string)

	Get(url string) (resp *http.Response, err error)
	Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error)
	PostForm(url string, data url.Values) (resp *http.Response, err error)
}

type RemnantHandlerFunc func(http.ResponseWriter, *http.Request, Client)

func GetInstrumentedHandler(remnantUrl string, f RemnantHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cl, err := NewRemnantClient(remnantUrl, r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Could not create Remnant client: ", err.Error())
		} else {
			cl.EndSpan()
		}
		f(w, r, cl)
	}
}

type RemnantClient struct {
	remnantUrl string
	Span       *types.Span
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
	cl.Span = &types.Span{}

	cl.remnantUrl = remnantUrl
	if strings.HasSuffix(remnantUrl, "/") {
		cl.remnantUrl = remnantUrl[:len(remnantUrl)-1]
	}

	cl.Span.TraceId = req.Header.Get(traceIdHeader)
	if cl.Span.TraceId == "" {
		cl.Span.TraceId = generateId()
	}
	cl.Span.Id = req.Header.Get(spanIdHeader)
	if cl.Span.Id == "" {
		cl.Span.Id = generateId()
	}
	cl.Span.ParentId = req.Header.Get(parentSpanIdHeader)
	if cl.Span.ParentId == "" {
		cl.Span.Id = generateId()
	}

	cl.Span.ClientStart = getTimestamp()

	req.ParseForm()
	cl.Span.Host = req.Host
	cl.Span.Method = req.Method
	cl.Span.Url = req.URL.String()
	headers := make([]types.KeyValue, 0, 10)
	for k, v := range req.Header {
		headers = append(headers, types.KeyValue{Key: k, Value: strings.Join(v, ",")})
	}
	cl.Span.Headers = headers
	params := make([]types.KeyValue, 0, 10)
	for k, v := range req.PostForm {
		params = append(params, types.KeyValue{Key: k, Value: strings.Join(v, ",")})
	}
	cl.Span.Parameters = params

	var body []byte
	var err error
	if body, err = ioutil.ReadAll(req.Body); err != nil {
		return nil, err
	}
	cl.Span.Body = string(body)

	return cl, nil
}

func (cl *RemnantClient) EndSpan() {
	cl.Span.ClientEnd = getTimestamp()
	spanJson, err := json.Marshal(cl.Span)
	if err != nil {
		fmt.Printf("Error marshalling span: %s\n", err.Error)
		return
	}
	go http.Post(cl.getLocalServiceUrl(), "application/json", bytes.NewReader(spanJson))
}

// generates rfc3339 formated timestrings
func getTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
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
	req.Header.Add(traceIdHeader, rc.Span.TraceId)
	req.Header.Add(spanIdHeader, spanId)
	req.Header.Add(parentSpanIdHeader, rc.Span.Id)
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
func (rc *RemnantClient) LogException(msg, stackTrace string) {
	rc.log(msg, stackTrace, "exception")
}

func (rc *RemnantClient) log(msg, stackTrace, level string) {
	logMsg := &types.LogMessage{}
	logMsg.TraceId = rc.Span.TraceId
	logMsg.SpanId = rc.Span.Id
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

	span := &types.Span{}
	span.TraceId = rc.Span.TraceId
	span.Id = spanId
	span.ParentId = rc.Span.Id
	span.RemoteStart = getTimestamp()

	addRequestHeaders(rc, req, spanId)
	httpClient := &http.Client{}
	resp, err = httpClient.Do(req)
	span.RemoteEnd = getTimestamp()
	span.ResponseCode = resp.StatusCode

	// asynchronously post span to server so that we don't incur the cost
	// of waiting for a response
	go rc.postSpan(span)

	return resp, err
}

func (client *RemnantClient) postSpan(span *types.Span) {
	spanJson, err := json.Marshal(span)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling json span: %s\n", err.Error())
	}
	http.Post(client.getRemoteRequestUrl(), "application/json", bytes.NewReader(spanJson))
}

func (rc *RemnantClient) Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	return nil, nil
}

func (rc *RemnantClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return nil, nil
}
