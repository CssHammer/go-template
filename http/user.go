package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/CssHammer/go-template/models"
	"github.com/CssHammer/go-template/service"
)

func (s *HTTPService) getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)[ParamID]
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "id must be int", http.StatusBadRequest)
		return
	}

	user, err := s.service.GetUser(r.Context(), int(idInt))
	if err != nil {
		switch {
		case errors.As(err, &service.ErrNotValidRequest{}):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			s.log.Error("service: get user", zap.Error(err))
		}
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		s.log.Error("encode user", zap.Error(err))
	}
}

func (s *HTTPService) postUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "can't decode body", http.StatusBadRequest)
		return
	}

	err = s.service.CreateUser(r.Context(), user)
	if err != nil {
		switch {
		case errors.As(err, &service.ErrNotValidRequest{}):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			s.log.Error("service: create user", zap.Error(err))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
