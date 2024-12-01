package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Kaese72/organization-registry/authentication"
	"github.com/Kaese72/riskie-lib/apierror"
	"go.elastic.co/apm/module/apmgorilla/v2"

	"github.com/Kaese72/finding-registry/internal/application"
	"github.com/Kaese72/finding-registry/rest/models"
	"github.com/gorilla/mux"
)

type restApplicationMux struct {
	application application.ApplicationLogic
}

func (appMux restApplicationMux) findingGetHandler(w http.ResponseWriter, r *http.Request) {
	organizationID := int(r.Context().Value(authentication.OrganizationIDKey).(float64))
	vars := mux.Vars(r)
	identifier, ok := vars["identifier"]
	if !ok {
		apierror.TerminalHTTPError(r.Context(), w, apierror.APIError{Code: http.StatusBadRequest, WrappedError: errors.New("missing identifier")})
		return
	}
	finding, err := appMux.application.ReadFinding(r.Context(), identifier, organizationID)
	if err != nil {
		apierror.TerminalHTTPError(r.Context(), w, err)
		return
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "   ")
	err = encoder.Encode(finding)
	if err != nil {
		apierror.TerminalHTTPError(r.Context(), w, err)
		return
	}
}

func (appMux restApplicationMux) findingsGetHandler(w http.ResponseWriter, r *http.Request) {
	organizationId := int(r.Context().Value(authentication.OrganizationIDKey).(float64))
	findings, err := appMux.application.ReadFindings(r.Context(), organizationId)
	if err != nil {
		apierror.TerminalHTTPError(r.Context(), w, err)
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
		apierror.TerminalHTTPError(r.Context(), w, err)
		return
	}
}

func (appMux restApplicationMux) findingsPostHandler(w http.ResponseWriter, r *http.Request) {
	organizationID := int(r.Context().Value(authentication.OrganizationIDKey).(float64))
	inputFinding := models.Finding{}
	err := json.NewDecoder(r.Body).Decode(&inputFinding)
	if err != nil {
		apierror.TerminalHTTPError(r.Context(), w, apierror.APIError{Code: http.StatusBadRequest, WrappedError: fmt.Errorf("error decoding request: %s", err.Error())})
		return
	}
	// Reset Identifier, just to be sure
	// FIXME should not be fixed here
	inputFinding.Identifier = ""
	findingR, err := appMux.application.PostFinding(r.Context(), inputFinding.ToIntermediary(), organizationID)
	if err != nil {
		apierror.TerminalHTTPError(r.Context(), w, err)
		return
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "   ")
	err = encoder.Encode(models.FindingFromIntermediary(findingR))
	if err != nil {
		apierror.TerminalHTTPError(r.Context(), w, err)
		return
	}
}

func InitMux(logic application.ApplicationLogic, jwtSecret string) *mux.Router {
	router := mux.NewRouter().PathPrefix("/finding-registry").Subrouter()
	apmgorilla.Instrument(router)
	appMux := restApplicationMux{application: logic}
	router.Use(authentication.DefaultJWTAuthentication(jwtSecret))
	router.HandleFunc("/findings/{identifier}", appMux.findingGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/findings", appMux.findingsGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/findings", appMux.findingsPostHandler).Methods(http.MethodPost)
	return router
}
