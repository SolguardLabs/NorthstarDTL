package api

import (
	"encoding/json"
	"net/http"

	"github.com/solguardlabs/northstardtl/src/domain"
)

func NewHTTPHandler(service *Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /snapshot", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, service.Snapshot())
	})
	mux.HandleFunc("POST /quote", func(w http.ResponseWriter, r *http.Request) {
		var intent domain.Intent
		if err := json.NewDecoder(r.Body).Decode(&intent); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		quotes, err := service.QuoteIntent(intent)
		if err != nil {
			writeError(w, statusFor(err), err)
			return
		}
		writeJSON(w, quotes)
	})
	mux.HandleFunc("POST /submit", func(w http.ResponseWriter, r *http.Request) {
		var request domain.SubmitIntentRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		response, err := service.SubmitIntent(request)
		if err != nil {
			writeError(w, statusFor(err), err)
			return
		}
		writeJSON(w, response)
	})
	mux.HandleFunc("POST /execute", func(w http.ResponseWriter, r *http.Request) {
		var request domain.ExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		response, err := service.Execute(request)
		if err != nil {
			writeError(w, statusFor(err), err)
			return
		}
		writeJSON(w, response)
	})
	mux.HandleFunc("POST /route", func(w http.ResponseWriter, r *http.Request) {
		var patch domain.RoutePatch
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		route, err := service.UpdateRoute(patch)
		if err != nil {
			writeError(w, statusFor(err), err)
			return
		}
		writeJSON(w, route)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("content-type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	if converted, ok := domain.AsDomainError(err); ok {
		_ = json.NewEncoder(w).Encode(converted)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "internal", "message": err.Error()})
}

func statusFor(err error) int {
	if converted, ok := domain.AsDomainError(err); ok {
		switch converted.Code {
		case domain.ErrInvalidIntent:
			return http.StatusBadRequest
		case domain.ErrInsufficientFunds, domain.ErrInsufficientRoute, domain.ErrExposureLimit:
			return http.StatusConflict
		case domain.ErrRouteNotFound, domain.ErrTicketNotFound:
			return http.StatusNotFound
		default:
			return http.StatusUnprocessableEntity
		}
	}
	return http.StatusInternalServerError
}
