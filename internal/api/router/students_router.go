package router

import (
	"net/http"
	"student_management_api/Golang/internal/api/handlers"
)

func StudentRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /students", handlers.GetStudentsHndler)
	mux.HandleFunc("POST /students", handlers.AddStudentsHandler)
	mux.HandleFunc("PUT /students", handlers.UpdateStudentHadler)
	mux.HandleFunc("PATCH /students", handlers.PatchStudentsHandler)
	mux.HandleFunc("DELETE /students", handlers.DeleteStudentsHandler)

	mux.HandleFunc("PUT /students/{id}", handlers.UpdateStudentHadler)
	mux.HandleFunc("PATCH /students/{id}", handlers.PatchOneStudentsHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteStudentHandler)
	mux.HandleFunc("GET /students/{id}", handlers.GetStudentHndler)

	return mux
}
