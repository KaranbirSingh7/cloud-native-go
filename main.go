package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// var store = make(map[string]string)

var store = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

var ErrorNoSuchKey = fmt.Errorf("no such key")

func main() {
	r := mux.NewRouter()

	log.Println("starting server")

	// routes
	r.HandleFunc("/healthz", healthz)
	r.HandleFunc("/v1/key/{key}", keyVauleGetHandler).Methods("GET")
	r.HandleFunc("/v1/key/{key}", keyVaulePutHandler).Methods("PUT")
	r.HandleFunc("/v1/key/{key}", keyVauleDeleteHandler).Methods("DELETE")

	// listen on port :8080
	log.Fatal(http.ListenAndServe(":8080", r))

}

func keyVauleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	Delete(key)

	w.WriteHeader(http.StatusOK)
}
func keyVauleGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := Get(key)
	if err != nil {
		if errors.Is(err, ErrorNoSuchKey) {
			http.Error(
				w,
				err.Error(),
				http.StatusNotFound,
			)
			return
		}
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, value)
}

func keyVaulePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// key is in URL path
	key := vars["key"]

	// value in body content
	value, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
	defer r.Body.Close()

	if err = Put(key, string(value)); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// 201
	w.WriteHeader(http.StatusCreated)
}

func Put(key, value string) error {
	store.Lock()
	store.m[key] = value
	store.Unlock()
	log.Println(store.m)
	return nil
}

func Delete(key string) {
	store.Lock()
	delete(store.m, key)
	store.Unlock()
	log.Println("Deleted key", key)
}

func Get(key string) (string, error) {
	store.RLock()
	value, ok := store.m[key]
	store.RUnlock()

	if !ok {
		return "", ErrorNoSuchKey
	}

	return value, nil
}

func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "UP")
}
