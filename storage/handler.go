package storage

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	db     DB
	logger TransactionLogger
}

func NewHandler(db DB, logger TransactionLogger) (Handler, error) {
	return Handler{db: db, logger: logger}, nil
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := h.db.Get(key)
	if errors.Is(err, ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, value)
}

func (h *Handler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	value, err := h.db.GetAll()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, value)
}

func (h *Handler) UpsertHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer r.Body.Close()
	h.db.Upsert(key, string(value))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.WritePut(key, string(value))
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := h.db.Delete(key)
	if errors.Is(err, ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.WriteDelete(key)
}
