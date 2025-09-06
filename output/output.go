package output

import (
	"encoding/json"
	"fmt"
	"os"
)

type ProgressReport struct {
	Title             string      `json:"title"`
	Message           string      `json:"message"`
	Percentage        float64     `json:"percentage"`
	Current           int         `json:"current"`
	Total             int         `json:"total"`
	EntityData        interface{} `json:"entity_data"`
	ExtraMessage      string      `json:"extra_message"`
	ExtraMessageColor string      `json:"extra_message_color"`
}

type ErrorReport struct {
	Message      string `json:"message"`
	LongMessage  string `json:"long_message"`
	ErrorMessage string `json:"error"`
}
type ErrorStruct struct {
	Message string `json:"error"`
}

func Error(message string) {
	ErrorJson(ErrorStruct{
		Message: message,
	})
}

func ErrorMessage(message, longMessage, errorMessage string) {
	ErrorJson(ErrorReport{
		Message:      message,
		LongMessage:  longMessage,
		ErrorMessage: errorMessage,
	})
}
func Message(message string) {
	os.Stdout.WriteString(fmt.Sprintf("{\"message\": \"%s\"}", message))
}
func Progress(title string, message string, percentage float64, current int, total int, entityData interface{}) {
	Json(ProgressReport{
		Title:      title,
		Message:    message,
		Percentage: percentage,
		Current:    current,
		Total:      total,
		EntityData: entityData,
	})
}
func Json(obj interface{}) {
	objJson, err := json.Marshal(obj)
	if err != nil {
		Error(err.Error())
	}
	os.Stdout.WriteString(string(objJson))
	os.Stdout.WriteString("\n")
}
func ErrorJson(obj interface{}) {
	objJson, err := json.Marshal(obj)
	if err != nil {
		Error(err.Error())
	}
	os.Stderr.WriteString(string(objJson))
	os.Stderr.WriteString("\n")
}
