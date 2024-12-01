package jwt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

func JWTMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("auth middleware ok!")

			// log.Println(r.Header)

			token, err := VerifyAccessToken(r)
			if err != nil {
				accessTokenErr := err

				token, err = VerifyRefreshToken(r)
				if err != nil {
					http.Error(w, fmt.Sprintf("access token err: %v\nrefresh token err: %v", accessTokenErr, err), http.StatusForbidden)
					return
				}
			}

			claims, ok := token.Claims.(jwt.MapClaims)

			if ok && token.Valid {
				_, ok := claims["tokenUuid"].(string)
				if !ok {
					http.Error(w, "error token uuid", http.StatusForbidden)
					return
				}

				_, ok = claims["userId"]

				if !ok {
					http.Error(w, "error user id", http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
			}
		})
	}
}
