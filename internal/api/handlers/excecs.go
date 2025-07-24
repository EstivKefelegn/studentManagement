package handlers

import (
	"fmt"
	"net/http"
)

func ExcecsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello get request from excecs  page"))
		fmt.Println("hello get request from excecs  page")
	case http.MethodPost:

		// fmt.Println("Query: ", r.URL.Query())
		// fmt.Println("Name: ", r.URL.Query().Get("name"))
		// fmt.Println("Name: ", r.URL.Query().Get("name"))

		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Something went wrong %v", err)
			return
		}
		fmt.Println("Form from POST methods =========> ", r.Form)
		w.Write([]byte("Hello post request from excecs  page"))
		fmt.Println("Hello post request from excecs  page")
	case http.MethodPut:
		w.Write([]byte("Hello put request from excecs  page"))
		fmt.Println("Hello put request from excecs  page")
	case http.MethodDelete:
		w.Write([]byte("Hello delete request from excecs  page"))
		fmt.Println("Hello delete requets from excecs  page")
	}
}
