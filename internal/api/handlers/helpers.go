package handlers

import (
	"errors"
	"reflect"
	"strings"
	"student_management_api/Golang/pkg/utils"
)

func CheckEmptyFields(value interface{}) error {
	val := reflect.ValueOf(value)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String && field.String() == "" {
			return utils.ErrorHandler(errors.New("all fields are required"), "all fields are required")

		}
	}
	return nil
}

func GetFieldNames(model interface{}) []string {
	val := reflect.TypeOf(model)
	fields := []string{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		// fieldToAdd := strings.TrimSuffix(field.Tag.Get("json"), ".omitempty")
		fieldToAdd := strings.Split(field.Tag.Get("json"), ",")[0]
		fields = append(fields, fieldToAdd)
	}
	return fields
}
