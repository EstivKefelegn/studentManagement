package middlewares

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"student_management_api/Golang/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)


func JWTMiddleware(next http.Handler) http.Handler {
	fmt.Println("--------JWT Middleware------------")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("++++++++++ Inside JWT Middleware ++++++++++")

		token, err := r.Cookie("Bearer")
		if err != nil {
			http.Error(w, "Authorization Header MIssing", http.StatusUnauthorized)
			utils.ErrorHandler(err, "Authorization Header MIssing")
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET")

		parsedToken, err := jwt.Parse(token.Value, func(token *jwt.Token) (any, error) {
			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing methid: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		// , jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Token Expired", http.StatusUnauthorized)
				utils.ErrorHandler(err, "Token Expired")
				return
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				http.Error(w, "Error Malformed", http.StatusUnauthorized)
				utils.ErrorHandler(err, "Error Malformed")
				return
			}

			utils.ErrorHandler(err, "")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if parsedToken.Valid {
			log.Println("Valid JWT")
		} else {
			http.Error(w, "Invalid Login Token", http.StatusUnauthorized)
			log.Println("Invalid JWT")
		}
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if ok {
			fmt.Println(claims["uid"], claims["exp"], claims["role"])
		} else {
			fmt.Println(err)
		}

		ctx := context.WithValue(r.Context(), utils.ContextKey("role"), claims["role"])
		ctx = context.WithValue(ctx, utils.ContextKey("expiresAt"), claims["exp"])
		ctx = context.WithValue(ctx, utils.ContextKey("username"), claims["user"])
		ctx = context.WithValue(ctx, utils.ContextKey("userId"), claims["uid"])	

		fmt.Println(ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
		fmt.Println("Sent response from JWT middleware")
	})
}
