package types

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

type Span struct {
	TraceId      string     `json:traceId`
	Id           string     `json:spanId`
	ParentId     string     `json:parentSpanId`
	ClientStart  string     `json:startTimestamp`
	ClientEnd    string     `json:endTimestamp`
	RemoteStart  string     `json:startTimestamp`
	RemoteEnd    string     `json:endTimestamp`
	Host         string     `json:host`
	Method       string     `json:method`
	Url          string     `json:rurl`
	Headers      []KeyValue `json:headers`
	Parameters   []KeyValue `json:parameters`
	Body         string     `json:body`
	ResponseCode int        `json:responseCode`
}
