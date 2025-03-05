package main

import (
	"errors"
	"fmt"
	"io"
	"keyvaluestore/storage"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	db storage.DB
}

func main() {

	db, err := storage.NewInMemoryDB()

	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("DB successfully initialized")
	}

	handler := &Handler{db: db}

	router := mux.NewRouter()
	router.HandleFunc("/v1/key", handler.GetAllHandler).Methods("GET")
	router.HandleFunc("/v1/key/{key}", handler.GetHandler).Methods("GET")
	router.HandleFunc("/v1/key/{key}", handler.UpsertHandler).Methods("PUT")
	router.HandleFunc("/v1/key/{key}", handler.DeleteHandler).Methods("DELETE")

	log.Printf("serving on port 8080")

	err = http.ListenAndServe(":8080", router)
	log.Fatal(err)
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := h.db.Read(key)
	if errors.Is(err, storage.ErrorNoSuchKey) {
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
	value, err := h.db.ReadAll()

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
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := h.db.Delete(key)
	if errors.Is(err, storage.ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
