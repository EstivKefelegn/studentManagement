package utils

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

func GenerateInsertQuery(tableName string, model interface{}) string {
	modelType := reflect.TypeOf(model)
	var columns, placeholders string
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		fmt.Println(dbTag)
		dbTag = strings.TrimSuffix(dbTag, ",omitempty")
		if dbTag != "" && dbTag != "id" {
			if columns != "" {
				columns += ", "
				placeholders += ", "
			}
			columns += dbTag
			placeholders += "?"
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columns, placeholders)
}

func GetStructValues(model interface{}) []interface{} {
	modelValue := reflect.ValueOf(model)
	modelType := modelValue.Type()
	values := []interface{}{}
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		if dbTag != "" && dbTag != "id,omitempty" {
			values = append(values, modelValue.Field(i).Interface())
		}

	}
	log.Println("Values: ", values)
	return values
}

func AddSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 { // if sortvalue exists
		query += " ORDER BY"
		for i, param := range sortParams {
			// /?sortby=last_name:asc
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !IsValidSortField(field) || !IsValidSortOrder(order) {
				continue
			}

			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

func AddFilter(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"calss":      "class",
		"subject":    "subject",
	}

	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}

	return query, args
}

func IsValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func IsValidSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"calss":      true,
		"subject":    true,
	}

	return validFields[field]
}
