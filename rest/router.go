package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Kaese72/finding-registry/internal/application"
	"github.com/Kaese72/finding-registry/rest/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func terminalHTTPError(w http.ResponseWriter, err error) {
	var apiError APIError
	if errors.As(err, &apiError) {
		if apiError.Code == 500 {
			// When an unknown error occurs, do not send the error to the client
			http.Error(w, "Internal Server Error", apiError.Code)
			log.Print(err.Error())
			return

		} else {
			bytes, intErr := json.MarshalIndent(apiError, "", "   ")
			if intErr != nil {
				// Must send a normal Error an not APIError just in case of eternal loop
				terminalHTTPError(w, fmt.Errorf("error encoding response: %s", intErr.Error()))
				return
			}
			http.Error(w, string(bytes), apiError.Code)
			return
		}
	} else {
		terminalHTTPError(w, APIError{Code: http.StatusInternalServerError, WrappedError: err})
		return
	}
}

type restApplicationMux struct {
	application application.ApplicationLogic
	jwtSecret   string
}

func (appMux restApplicationMux) findingGetHandler(w http.ResponseWriter, r *http.Request) {
	organizationID := int(r.Context().Value(organizationIDKey).(float64))
	vars := mux.Vars(r)
	identifier, ok := vars["identifier"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	finding, err := appMux.application.ReadFinding(identifier, organizationID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "   ")
	err = encoder.Encode(finding)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}
}

func (appMux restApplicationMux) findingsGetHandler(w http.ResponseWriter, r *http.Request) {
	organizationId := int(r.Context().Value(organizationIDKey).(float64))
	findings, err := appMux.application.ReadFindings(organizationId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}
	result := []models.Finding{}
	for index := range findings {
		result = append(result, models.FindingFromIntermediary(findings[index]))
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "   ")
	err = encoder.Encode(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}
}

func (appMux restApplicationMux) findingsPostHandler(w http.ResponseWriter, r *http.Request) {
	organizationID := int(r.Context().Value(organizationIDKey).(float64))
	inputFinding := models.Finding{}
	err := json.NewDecoder(r.Body).Decode(&inputFinding)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err.Error())
		return
	}
	// Reset Identifier, just to be sure
	// FIXME should not be fixed here
	inputFinding.Identifier = ""
	findingR, err := appMux.application.PostFinding(inputFinding.ToIntermediary(), organizationID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "   ")
	err = encoder.Encode(models.FindingFromIntermediary(findingR))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}
}

type contextKey string

const (
	userIDKey         contextKey = "userID"
	organizationIDKey contextKey = "organizationID"
)

func (app restApplicationMux) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(app.jwtSecret), nil
		})

		if err != nil {
			terminalHTTPError(w, APIError{Code: http.StatusUnauthorized, WrappedError: fmt.Errorf("error parsing token: %s", err.Error())})
			return
		}

		if !token.Valid {
			terminalHTTPError(w, APIError{Code: http.StatusUnauthorized, WrappedError: errors.New("invalid token")})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			terminalHTTPError(w, APIError{Code: http.StatusUnauthorized, WrappedError: errors.New("could not read claims")})
			return
		}

		userID, ok := claims[string(userIDKey)].(float64)
		if !ok {
			terminalHTTPError(w, APIError{Code: http.StatusUnauthorized, WrappedError: errors.New("could not read userId claim")})
			return
		}
		organizationID, ok := claims[string(organizationIDKey)].(float64)
		if !ok {
			terminalHTTPError(w, APIError{Code: http.StatusUnauthorized, WrappedError: errors.New("could not read organizationId claim")})
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, organizationIDKey, organizationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func InitMux(logic application.ApplicationLogic, jwtSecret string) *mux.Router {
	router := mux.NewRouter().PathPrefix("/finding-registry").Subrouter()
	appMux := restApplicationMux{application: logic, jwtSecret: jwtSecret}
	router.Use(appMux.authMiddleware)
	router.HandleFunc("/findings/{identifier}", appMux.findingGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/findings", appMux.findingsGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/findings", appMux.findingsPostHandler).Methods(http.MethodPost)
	return router
}
