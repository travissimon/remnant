package data

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/travissimon/remnant/types"
)

var db *sql.DB

func init() {
	database, err := sql.Open("postgres", "host=localhost dbname=tsimon sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database\n")
		return
	}
	err = database.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not ping database\n")
		return
	}
	db = database

	prepareSqlStatements()
}

func InsertSpan(span types.Span) {
	headers := convertKVArray(span.Headers)
	parameters := convertKVArray(span.Parameters)

	_, err := insertSpanStatement.Exec(
		span.TraceId,
		span.Id,
		span.ParentId,
		getTimestamp(span.ClientStart),
		getTimestamp(span.ClientEnd),
		getTimestamp(span.RemoteStart),
		getTimestamp(span.RemoteEnd),
		span.Host,
		span.Method,
		span.Url,
		headers,
		parameters,
		span.Body,
		span.ResponseCode)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting span: %s\n", err.Error())
	}
}

func InsertLog(log types.LogMessage) {
	_, err := insertLogStatement.Exec(
		log.TraceId,
		log.SpanId,
		log.Severity,
		log.Message,
		log.StackTrace)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting span: %s\n", err.Error())
	}
}

func prepareSqlStatements() {
	prepareSelectSpanByIdStatement()
	prepareInsertSpanStatement()
	prepareInsertLogStatement()
}

var selectSpanByIdStatement *sql.Stmt
var insertSpanStatement *sql.Stmt
var insertLogStatement *sql.Stmt

func prepareSelectSpanByIdStatement() {
	stmt, err := db.Prepare(`SELECT span_id FROM spans WHERE span_id = $1`)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating select span by id statement: %s\n", err.Error())
	}

	selectSpanByIdStatement = stmt
}

func prepareInsertSpanStatement() {
	stmt, err := db.Prepare(`
	INSERT INTO spans (
		trace_id,
		span_id,
		parent_span_id,
		client_start,
		client_end,
		remote_start,
		remote_end,
		host,
		method,
		url,
		headers,
		parameters,
		body,
		response_code
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11,
		$12,
		$13,
		$14
	)`)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating insert client prepared statement: %s\n", err.Error())
	}

	insertSpanStatement = stmt
}

func prepareInsertLogStatement() {
	stmt, err := db.Prepare(`
	INSERT INTO logs (
		trace_id,
		span_id,
		severity,
		message,
		stack_trace
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5

	)`)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating insert log prepared statement: %s\n", err.Error())
	}

	insertLogStatement = stmt
}

func getTimestamp(t string) interface{} {
	if t == "" {
		return nil
	}
	return t
}

func convertKVArray(keyValues []types.KeyValue) string {
	var buffer bytes.Buffer

	var upperBound = len(keyValues) - 1
	for idx, kv := range keyValues {
		buffer.WriteString(kv.Key)
		buffer.WriteString("=\"")
		buffer.WriteString(kv.Value)
		buffer.WriteString("\"")
		if idx != upperBound {
			buffer.WriteString(", ")
		}
	}

	return buffer.String()
}
