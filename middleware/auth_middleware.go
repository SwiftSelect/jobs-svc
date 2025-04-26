package middleware

import (
	"context"
	"jobs-svc/internal/clients"
	"net/http"
	"strings"
)

func AuthMiddleware(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Extract token from the header
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			// Validate token using the auth microservice
			userInfo, err := clients.ValidateToken(token, action)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			// Add user info to the request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "userInfo", userInfo)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
