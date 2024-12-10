package auth

import (
	"net/http"

	"github.com/gorilla/mux"
)


func CorsMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// log.Println("cors middleware ok!")

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE, PATCH")

			// Handle preflight (OPTIONS) request by returning 200 OK with the necessary headers
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call the next handler in the chain (e.g., your AuthMiddleware)
			next.ServeHTTP(w, r)
		})
	}
}
