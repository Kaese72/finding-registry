package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Kaese72/finding-registry/internal/application"
	"github.com/Kaese72/finding-registry/rest/models"
	"github.com/gorilla/mux"
)

type restApplicationMux struct {
	application application.ApplicationLogic
}

func (appMux restApplicationMux) findingGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	identifier, ok := vars["identifier"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	finding, err := appMux.application.ReadFinding(identifier)
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
	findings, err := appMux.application.ReadFindings()
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
	findingR, err := appMux.application.PostFinding(inputFinding.ToIntermediary())
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

func InitMux(logic application.ApplicationLogic) *mux.Router {
	router := mux.NewRouter()
	appMux := restApplicationMux{application: logic}
	router.HandleFunc("/findings/{identifier}", appMux.findingGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/findings", appMux.findingsGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/findings", appMux.findingsPostHandler).Methods(http.MethodPost)
	return router
}
