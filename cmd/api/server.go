package main

import (
	"crypto/tls"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	mw "student_management_api/Golang/internal/api/middlewares"
	"student_management_api/Golang/internal/api/router"
	"student_management_api/Golang/pkg/utils"
	"time"
	_ "time"

	"github.com/joho/godotenv"
)

//go:embed .env
var envFile embed.FS

func loadEnvFromEmbededFile() {
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}

	tempfile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating temp .env file %v:", err)
	}

	defer os.Remove(tempfile.Name())

	_, err = tempfile.Write(content)
	if err != nil {
		log.Fatalf("Error closing tempfile %v:", err)
	}

	err = tempfile.Close()
	if err != nil {
		log.Fatalf("Error closing tempfile %v:", err)
	}

	err = godotenv.Load(tempfile.Name())
	if err != nil {
		log.Fatalf("Error loading .env file %v:", err)
	}
}

func main() {
	//Only in production, for running a source code
	// err := godotenv.Load()
	// if err != nil {
	// 	return
	// }

	// Load envirenment variable
	loadEnvFromEmbededFile()
	fmt.Println("Environment variable CERT_FILE", os.Getenv("CERT_FILE"))

	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	port := os.Getenv("API_PORT")

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	rl := mw.NewRateLimiter(5, time.Minute)
	hppOptions := mw.HPPOptions{
		CheckQuery:                  true,
		CheckBody:                   true,
		CheckBodyOnlyForContentType: "applicaiotn/x-www-form-urlencoded",
		Whitelist:                   []string{"sortBy", "sortOrder", "name", "age", "class", "first_name", "last_name"},
	}

	// secureMux := jwtMIddleware(mw.SecurityHeader(router.MainRouter()))
	// secureMux := mw.XSSMiddleware(router.MainRouter())
	router := router.MainRouter()
	jwtMIddleware := mw.MiddlewaresExcludePaths(mw.JWTMiddleware, "/execs/login", "/execs/forgotpassword", "/execs/resetpassword/reset")
	secureMux := utils.ApplyingMiddleware(router, mw.Hpp(hppOptions), mw.SecurityHeader, mw.Compression, mw.XSSMiddleware, jwtMIddleware, mw.ResponseTimeMiddleware, rl.Middleware, mw.Cors)

	// secureMux := mw.SecurityHeader(router.MainRouter())
	fmt.Println("Server is going to start")
	server := &http.Server{
		Addr:      port,
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server running on port :3000")
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Println("TLS server failed: ", err)
		log.Fatal(err)
	}

}
