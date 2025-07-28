package router

import (
	"net/http"
	"student_management_api/Golang/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)

	mux.HandleFunc("/students", handlers.StudentsHandler)

	mux.HandleFunc("GET /teachers/", handlers.GetTeachersHndler)
	mux.HandleFunc("POST /teachers/", handlers.AddTeachersHandler)
	mux.HandleFunc("PUT /teachers/", handlers.UpdateTeacherHadler)
	mux.HandleFunc("PATCH /teachers/", handlers.PatchTeachersHandler)
	mux.HandleFunc("DELETE /teachers/", handlers.DeleteTeachersHandler)

	mux.HandleFunc("PUT /teachers/{id}", handlers.UpdateTeacherHadler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PatchOneTeachersHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteTeacherHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetTeacherHndler)

	mux.HandleFunc("/execs/", handlers.ExcecsHandler)

	return mux
}
