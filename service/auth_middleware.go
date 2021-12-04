package service

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type JWTAuthMiddleware struct {
	VerifyBytes []byte
}

func (m *JWTAuthMiddleware) GetMiddleWare(r *mux.Router) func(next http.Handler) http.Handler {
	if len(m.VerifyBytes) == 0 {
		log.Fatalf("cannot use the middleware without public key payload")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			token, err := extractToken(req)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, err.Error())
				return
			}
			userId, err := m.validateAccessToken(token)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, err.Error())
				return
			}
			// TODO
			log.Println("jwt user: ", userId)

			next.ServeHTTP(w, req)
		})
	}
}

func (m *JWTAuthMiddleware) validateAccessToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			log.Println("Unexpected signing method in auth token")
			return nil, errors.New("unexpected signing method in auth token")
		}

		verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(m.VerifyBytes)
		if err != nil {
			log.Println("unable to parse public key:", err)
			return nil, err
		}

		return verifyKey, nil
	})

	if err != nil {
		log.Println("unable to parse claims", err)
		return "", err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid || claims.Id == "" {
		return "", errors.New("invalid token: authentication failed")
	}
	return claims.Id, nil
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	authHeaderContent := strings.Split(authHeader, " ")
	if len(authHeaderContent) != 2 {
		return "", errors.New("token not provided or malformed")
	}
	return authHeaderContent[1], nil
}
