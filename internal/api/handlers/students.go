package handlers

import (
	"fmt"
	"net/http"
)

func StudentsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello get request from students page"))
		fmt.Println("hello get request from students page")
	case http.MethodPost:
		w.Write([]byte("Hello post request from student page"))
		fmt.Println("Hello post request from students page")
	case http.MethodPut:
		w.Write([]byte("Hello put request from student page"))
		fmt.Println("Hello put request from students page")
	case http.MethodDelete:
		w.Write([]byte("Hello delete request from student page"))
		fmt.Println("Hello delete requets from students page")
	}

}
