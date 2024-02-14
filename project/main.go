package main

import (
	"api/database"
	"api/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool  *pgxpool.Pool
	jb    database.JobQueue
	cache *database.Cache
)

func criaPessoa(w http.ResponseWriter, r *http.Request) {
	var pessoa model.Pessoa

	pessoa.ID = uuid.NewString()
	err := json.NewDecoder(r.Body).Decode(&pessoa)

	if err != nil {
		fmt.Println("err")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	apelidoExists, err := cache.GetApelido(pessoa.Apelido)
	if err != nil || apelidoExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := cache.Set(&pessoa); err != nil {
		fmt.Printf("Error setting pessoa in cache: %v", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	jb <- database.Job{Payload: &pessoa}

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Location", "/pessoas/"+pessoa.ID)
	w.WriteHeader(http.StatusCreated)
}

func getPessoa(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var pessoa *model.Pessoa

	pessoa, _ = cache.Get(params["id"])

	if pessoa == nil {
		err := pool.QueryRow(context.Background(),
			`select id, nome, apelido, to_char(nascimento, 'yyyy-mm-dd'), stack 
			from pessoa where pessoa.id = $1`,
			params["id"],
		).Scan(&pessoa.ID, &pessoa.Nome, &pessoa.Apelido, &pessoa.Nascimento, &pessoa.Stack)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	_ = json.NewEncoder(w).Encode(&pessoa)
}

func getPessoaByTerm(w http.ResponseWriter, r *http.Request) {
	var pessoas []model.Pessoa

	term := strings.ToLower(r.URL.Query().Get("t"))

	if len(term) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	cacheRows, _ := database.TryFindItemInCache(term, cache)

	if len(cacheRows) > 0 {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&cacheRows)
		return
	}

	pessoas = database.GetPessoaByTermInPostgres(term, pool)

	if len(pessoas) > 0 {
		if len(pessoas) > 0 {
			go func() {
				if err := cache.SetSearch(term, pessoas); err != nil {
					fmt.Printf("Error setting values in cache: %v", err)
				}
			}()
		}
	}

	_ = json.NewEncoder(w).Encode(&pessoas)

}

func countPessoas(w http.ResponseWriter, r *http.Request) {

	var count int

	err := pool.QueryRow(context.Background(),
		"select count(*) from pessoa",
	).Scan(&count)

	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%v", count)
}

func main() {
	uuid.EnableRandPool()

	r := mux.NewRouter()

	r.HandleFunc("/pessoas", criaPessoa).Methods("POST")
	r.HandleFunc("/pessoas/{id}", getPessoa).Methods("GET")
	r.HandleFunc("/pessoas", getPessoaByTerm).Methods("GET")
	r.HandleFunc("/contagem-pessoas", countPessoas).Methods("GET")

	pool = database.Connect()

	jb = database.CreateJobQueue()
	dispatcher := database.CreateDispatcher(pool, jb)
	cache = database.CreateCache()
	go dispatcher.Run()

	srv := &http.Server{
		Addr: "0.0.0.0:" + os.Getenv("HTTP_PORT"),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 2,
		ReadTimeout:  time.Second * 2,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	log.Fatal(srv.ListenAndServe())

}
