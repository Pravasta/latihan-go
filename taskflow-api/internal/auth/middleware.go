package auth

import (
	"fmt"
	"net/http"
	"taskflow-api/internal/common"
)

type AuthMiddleware struct {
	jwt *JWTService
}

func NewAuthMiddleware(jwt *JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		// 1. Get the Authorization header from the request
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			fmt.Println("[AuthMiddleware] Authorization header is missing")
			common.WriteError(w, http.StatusUnauthorized, "Authorization header is missing")
			return
		}

		// 2. Validate the token
		part := "Bearer "
		if len(authHeader) <= len(part) || authHeader[:len(part)] != part {
			fmt.Println("[AuthMiddleware] Invalid Authorization header format")
			common.WriteError(w, http.StatusUnauthorized, "Invalid Authorization header format")
			return
		}

		// 3. Extract the token from the header
		token := authHeader[len(part):]

		if token == "" {
			fmt.Println("[AuthMiddleware] Token is missing")
			common.WriteError(w, http.StatusUnauthorized, "Token is missing")
			return
		}

		// 4. Validate the token and get the user ID
		userID, err := m.jwt.Parse(token)
		if err != nil {
			fmt.Println("[AuthMiddleware] Failed to parse token:", err)
			common.WriteError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// 5. Set the user ID in the request context
		ctx := common.SetUserID(r.Context(), userID)
		r = r.WithContext(ctx)

		// 6. Call the next handler
		next.ServeHTTP(w, r)
	})
}
