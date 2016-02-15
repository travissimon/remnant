package dummydata

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/travissimon/remnant/types"
)

func init() {
}

func prettyPrint(obj interface{}) string {
	json, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling json: %s\n", err.Error())
	}
	return string(json)
}

func InsertSpan(span types.Span) {
	fmt.Printf("Insert span: %s\n\n", prettyPrint(span))
}

func InsertLog(log types.LogMessage) {
	fmt.Printf("Insert log: %s\n\n", prettyPrint(log))
}
