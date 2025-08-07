package router

import (
	"net/http"
)

func MainRouter() *http.ServeMux {

	tRouter := TeachersRouter()
	sRouter := StudentRouter()

	tRouter.Handle("/", sRouter)
	return tRouter	
	// mux := http.NewServeMux()

	// mux.HandleFunc("/", handlers.RootHandler)

	// mux.HandleFunc("/students", handlers.StudentsHandler)

	// mux.HandleFunc("GET /execs/", handlers.ExcecsHandler)

	// return mux
}
