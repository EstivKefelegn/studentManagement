package router

import (
	"net/http"
	"student_management_api/Golang/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)

	mux.HandleFunc("/students", handlers.StudentsHandler)

	mux.HandleFunc("GET /teachers", handlers.GetTeachersHndler)
	mux.HandleFunc("POST /teachers", handlers.AddTeachersHandler)
	mux.HandleFunc("PUT /teachers", handlers.UpdateTeacherHadler)
	mux.HandleFunc("PATCH /teachers", handlers.PatchTeachersHandler)
	mux.HandleFunc("DELETE /teachers", handlers.DeleteTeachersHandler)

	mux.HandleFunc("PUT /teachers/{id}", handlers.UpdateTeacherHadler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PatchOneTeachersHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteTeacherHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetTeacherHndler)

	mux.HandleFunc("GET /students", handlers.GetStudentsHndler)
	mux.HandleFunc("POST /students", handlers.AddStudentsHandler)
	mux.HandleFunc("PUT /students", handlers.UpdateStudentHadler)
	mux.HandleFunc("PATCH /students", handlers.PatchStudentsHandler)
	mux.HandleFunc("DELETE /students", handlers.DeleteStudentsHandler)

	mux.HandleFunc("PUT /students/{id}", handlers.UpdateStudentHadler)
	mux.HandleFunc("PATCH /students/{id}", handlers.PatchOneStudentsHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteStudentHandler)
	mux.HandleFunc("GET /students/{id}", handlers.GetStudentHndler)

	mux.HandleFunc("GET /execs/", handlers.ExcecsHandler)

	return mux
}
