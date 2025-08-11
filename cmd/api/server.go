package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	mw "student_management_api/Golang/internal/api/middlewares"
	"student_management_api/Golang/internal/api/router"
	"student_management_api/Golang/internal/repository/sqlconnect"
	_ "time"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		return
	}
	_, err = sqlconnect.ConnectDB()

	if err != nil {
		fmt.Println("Error------: ", err)
		return
	}

	cert := "cert.pem"
	key := "key.pem"

	port := os.Getenv("API_PORT") 

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	// rl := mw.NewRateLimiter(5, time.Minute)
	// hppOptions := mw.HPPOptions{
	// 	CheckQuery:                  true,
	// 	CheckBody:                   true,
	// 	CheckBodyOnlyForContentType: "applicaiotn/x-www-form-urlencoded",
	// 	Whitelist:                   []string{"sortBy", "sortOrder", "name", "age", "class", "first_name", "last_name"},
	// }

	// secureMux := utils.ApplyingMiddleware(router.Router(), mw.Hpp(hppOptions), mw.Compression, mw.SecurityHeader, mw.ResponseTimeMiddleware, rl.Middleware, mw.Cors)
	jwtMIddleware := mw.MiddlewaresExcludePaths(mw.JWTMiddleware, "/execs/login",  "/execs/forgotpassword", "/execs/resetpassword/reset")
	secureMux := jwtMIddleware(mw.SecurityHeader(router.MainRouter()))

	// secureMux := mw.SecurityHeader(router.MainRouter())
	fmt.Println("Server is going to start")
	server := &http.Server{
		Addr:      port,
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server running on port :3000")
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Println("TLS server failed: ", err)
		log.Fatal(err)
	}

}
