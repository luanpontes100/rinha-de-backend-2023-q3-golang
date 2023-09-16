package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Router *mux.Router
	DB     *pgxpool.Pool
}

func (a *App) Initialize(host, user, password, dbname string) {
	db_config, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", user, password, host, dbname))
	if err != nil {
		panic(err)
	}
	db_config.MaxConns = 100 // hardcoded, check pg's max_connection. must be half of that per instance
	db_config.MinConns = 10

	a.DB, err = pgxpool.NewWithConfig(context.Background(), db_config)
	if err != nil {
		panic(err)
	}

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

func (a *App) Run(port string) {
	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/pessoas", a.createPerson).Methods("POST")
	a.Router.HandleFunc("/pessoas", a.searchPeople).Methods("GET")
	a.Router.HandleFunc("/pessoas/{id}", a.getPerson).Methods("GET")
	a.Router.HandleFunc("/contagem-pessoas", a.getCountPeople).Methods("GET")
}
