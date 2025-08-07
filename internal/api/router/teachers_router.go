package router

import (
	"net/http"
	"student_management_api/Golang/internal/api/handlers"
)

func TeachersRouter() *http.ServeMux {
	
	mux := http.NewServeMux()

	mux.HandleFunc("GET /teachers", handlers.GetTeachersHndler)
	mux.HandleFunc("POST /teachers", handlers.AddTeachersHandler)
	mux.HandleFunc("PUT /teachers", handlers.UpdateTeacherHadler)
	mux.HandleFunc("PATCH /teachers", handlers.PatchTeachersHandler)
	mux.HandleFunc("DELETE /teachers", handlers.DeleteTeachersHandler)

	mux.HandleFunc("PUT /teachers/{id}", handlers.UpdateTeacherHadler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PatchOneTeachersHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteTeacherHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetTeacherHndler)

	mux.HandleFunc("GET /teachers/{id}/students", handlers.GetStudentsByTeacherID)
	mux.HandleFunc("GET /teachers/{id}/studentscount", handlers.GetStudentCountByTeacherID)

	return mux
}
