package main

import (
	"keyvaluestore/storage"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	db, err := storage.NewInMemoryDB()

	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("DB successfully initialized")
	}

	logger, err := storage.InitializeTransactionLogger(db)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("TransactionLogger successfully initialized")
	}

	handler, err := storage.NewHandler(db, logger)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Handler successfully initialized")
	}

	router := mux.NewRouter()
	router.HandleFunc("/v1/key", handler.GetAllHandler).Methods("GET")
	router.HandleFunc("/v1/key/{key}", handler.GetHandler).Methods("GET")
	router.HandleFunc("/v1/key/{key}", handler.UpsertHandler).Methods("PUT")
	router.HandleFunc("/v1/key/{key}", handler.DeleteHandler).Methods("DELETE")

	log.Printf("serving on port 8080")

	err = http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", router)
	log.Fatal(err)
}
