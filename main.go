package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jecitDev/api/handler"
)

func errPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "api paramater not recognition"}`))
}

func main() {
	r := mux.NewRouter()
	/*
		address design :  x.com:port/{prefix}/{app}/{entity}
		JSON body parameter (template), POST Method


		{prefix} lib
		api - design for external apps(public) or mobile
		db  - design for sync db, etc
		mdl - design for modular internal apps
	*/
	//                    modul
	mdl := r.PathPrefix("/core").Subrouter()
	mdl.HandleFunc("", errPost)
	mdl.HandleFunc("/{a}", errPost)
	mdl.HandleFunc("/{a}/{b}/{c}", errPost)
	mdl.HandleFunc("/{app}/{entity}", handler.GoPostData).Methods(http.MethodPost)

	log.Fatal(http.ListenAndServe(":443", r))
}
