package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"student_management_api/Golang/internal/api/middlewares"
)

type User struct {
	Name string `json:"name"`
	Age  string `json:"age"`
	City string `json:"city"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello root route"))
	fmt.Println("Hello root route")
}

func teachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello get request from teachers page"))
		fmt.Println("hello get request from teachers page")
	case http.MethodPost:

		// Query Params /?key=value&sortby=email&sortorder=ASC
		fmt.Println(r.URL.Path)
		path := r.URL.Path
		newUser := strings.TrimPrefix(path, "/teachers/")
		userId := strings.TrimSuffix(newUser, "/")
		fmt.Println("Uers ID is: ", userId)

		queryParam := r.URL.Query()
		sortOrder := queryParam.Get("sortorder")

		if sortOrder == "" {
			sortOrder = "DESC"
		}

		fmt.Println("Get the sort by value: ", sortOrder)
		fmt.Println("Query parmater is: ", queryParam)
		w.Write([]byte("Hello post request from teachers page"))
		fmt.Println("Hello post request from teachers page")
	case http.MethodPut:
		w.Write([]byte("Hello put request from teachers page"))
		fmt.Println("Hello put request from teachers page")
	case http.MethodDelete:
		w.Write([]byte("Hello delete request from teachers page"))
		fmt.Println("Hello delete requets from teachers page")
	}
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {

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

func excecsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello get request from excecs  page"))
		fmt.Println("hello get request from excecs  page")
	case http.MethodPost:
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

func main() {


	
	cert := "cert.pem"
	key := "key.pem"

	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	
	mux.HandleFunc("/students/", studentsHandler)
	
	mux.HandleFunc("/teachers/", teachersHandler)
	
	mux.HandleFunc("/exec/", excecsHandler)
	

	port := ":3000"

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	server := &http.Server{
		Addr: port,
		Handler: middlewares.SecurityHeader(mux) ,
		TLSConfig: tlsConfig,
	}



	fmt.Println("Server running on port :3000")
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatal(err)
	}

}
