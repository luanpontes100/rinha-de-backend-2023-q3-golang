package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) createPerson(w http.ResponseWriter, r *http.Request) {
	var p pessoa
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}
	defer r.Body.Close()
	if err := p.validate(); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err := p.createPerson(a.DB); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/pessoas/%s", p.Id.String()))
	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) getPerson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	p := pessoa{}
	if err := p.getPerson(a.DB, id); err != nil {
		switch err.Error() {
		case "no rows in result set":
			respondWithError(w, http.StatusNotFound, "Pessoa não encontrada")
		default:
			fmt.Println("ERROR IN REQUEST getPerson with ID " + id + ". Error: " + err.Error())
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) searchPeople(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("t")
	if len(term) == 0 {
		respondWithError(w, http.StatusBadRequest, "Termo de busca não informado")
		return
	}
	people, err := pessoas{}.searchPeople(a.DB, term)
	if err != nil {
		fmt.Println("ERROR IN REQUEST searchPeople with term " + term + ". Error: " + err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, people)
}

func (a *App) getCountPeople(w http.ResponseWriter, r *http.Request) {
	count, err := pessoas{}.totalPeople(a.DB)
	if err != nil {
		fmt.Println("ERROR IN REQUEST getCountPeople. Error: " + err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Fprintf(w, "%d", count)
}
