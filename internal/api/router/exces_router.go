package router

import (
	"net/http"
	"student_management_api/Golang/internal/api/handlers"
)

func ExcecRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /execs", handlers.GetExecsHndler)
	mux.HandleFunc("POST /execs", handlers.AddExecsHandler)
	mux.HandleFunc("PATCH /execs", handlers.PatchExecsHandler)

	mux.HandleFunc("GET /execs/{id}", handlers.GetOneExecHandler)
	mux.HandleFunc("PATCH /execs/{id}", handlers.PatchOneExecsHandler)
	mux.HandleFunc("DELETE /execs/{id}", handlers.DeleteExecHandler)
	// mux.HandleFunc("POST /execs/{id}/updatepassword", handlers.AddExecsHandler)

	// mux.HandleFunc("POST /execs/login", handlers.UpdateExecHadler)
	// mux.HandleFunc("POST /execs/logout", handlers.UpdateExecHadler)
	// mux.HandleFunc("POST /execs/forgotpassword", handlers.UpdateExecHadler)
	// mux.HandleFunc("POST /execs/resetpassword/reset/{resetcode}", handlers.UpdateExecHadler)

	return mux

}
