package main

import (
	"net/http"
	"log"
	"context"
	"github.com/gorilla/mux"
	"fmt"
	"os"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Amospk2/rinha-de-backend-2023-q3/project/database"
)

var (
	db *pgxpool.Pool
)

func isInvalidBody(pessoa Pessoa) (bool){
	
	if len(pessoa.Nome) == 0 || len(pessoa.Apelido)  == 0 || len(pessoa.Nascimento) == 0 {
		return true
	}

	if len(pessoa.Nome) > 100 || len(pessoa.Apelido) > 32 {
		return true
	}
	
	_, err := time.Parse("2006-01-02", pessoa.Nascimento)

	if err != nil {
		return true
	}

	for _, v := range pessoa.Stack {
		if len(v) > 32 {
			return true
		}
	}

	return false

}

func criaPessoa(w http.ResponseWriter, r *http.Request) {
	var pessoa Pessoa
	
	uuidV,_ := uuid.NewUUID()
	pessoa.ID = uuidV.String()
	err := json.NewDecoder(r.Body).Decode(&pessoa)
	

	if err != nil || isInvalidBody(pessoa) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}


	_, err = pool.Exec(context.Background(),
		`
		INSERT INTO Pessoa(id, apelido, nome, nascimento, stack)
		values ($1, $2, $3, $4, $5)`,
		pessoa.ID, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento, pessoa.Stack,
	)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Location", "/pessoas/" + pessoa.ID)
	w.WriteHeader(http.StatusCreated)
}

func getPessoa(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	var pessoa Pessoa

	err := pool.QueryRow(context.Background(),
		"select id, nome, apelido, to_char(nascimento, 'yyyy-mm-dd'), stack from pessoa where pessoa.id = $1", 
		params["id"],
	).Scan(&pessoa.ID, &pessoa.Nome, &pessoa.Apelido, &pessoa.Nascimento, &pessoa.Stack)
	

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(&pessoa)
}

func getPessoaByTerm(w http.ResponseWriter, r *http.Request){
	
	term := r.URL.Query().Get("t")

	if len(term) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var pessoas []Pessoa
	
	rows, err := pool.Query(context.Background(),
		`select id, nome, apelido, to_char(nascimento, 'yyyy-mm-dd'), stack 
		from pessoa where pessoa.termo_busca ILIKE '%` + term +`%'
		`, 
	)
	w.WriteHeader(http.StatusOK)

	if err != nil {
		_ = json.NewEncoder(w).Encode([]Pessoa{})
		return
	}
	
	defer rows.Close()
	
	for rows.Next() {
		var pessoa Pessoa

		err = rows.Scan(&pessoa.ID, &pessoa.Nome, &pessoa.Apelido, &pessoa.Nascimento, &pessoa.Stack)

		if err != nil {
			log.Fatal(err)
		}

		pessoas = append(pessoas, pessoa)
	}

	_ = json.NewEncoder(w).Encode(&pessoas)
	
}


func countPessoas(w http.ResponseWriter, r *http.Request){

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

func connect() {
	
	var err error
	pool, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Println("Unable to connecto to database")
	}

	_, err = pool.Exec(context.Background(), `
        CREATE EXTENSION IF NOT EXISTS pg_trgm;

        CREATE OR REPLACE FUNCTION generate_termo_busca(_nome VARCHAR, _apelido VARCHAR, _stack TEXT[])
            RETURNS TEXT AS $$
            BEGIN
            RETURN _nome || _apelido || _stack;
            END;
        $$ LANGUAGE plpgsql IMMUTABLE;

        CREATE TABLE IF NOT EXISTS pessoa (
            id uuid DEFAULT gen_random_uuid() UNIQUE NOT NULL,
            apelido TEXT UNIQUE NOT NULL,
            nome TEXT NOT NULL,
            nascimento DATE NOT NULL,
            stack TEXT[],
            termo_busca text GENERATED ALWAYS AS (generate_termo_busca(nome, apelido, stack)) STORED
        );

        CREATE INDEX IF NOT EXISTS idx_pessoa_termo_busca ON public.pessoa USING gist (termo_busca public.gist_trgm_ops (siglen='64'));

        CREATE UNIQUE INDEX IF NOT EXISTS pessoa_apelido_index ON public.pessoa USING btree (apelido);
	`)

	if err != nil {
		log.Fatal(err)
	}
		fmt.Println("Connect with success")
}


func main(){


	r := mux.NewRouter()

	r.HandleFunc("/pessoas", criaPessoa).Methods("POST")
	r.HandleFunc("/pessoas/{id}", getPessoa).Methods("GET")
	r.HandleFunc("/pessoas", getPessoaByTerm).Methods("GET")
	r.HandleFunc("/contagem-pessoas", countPessoas).Methods("GET")
	
	db = connect()
	
    srv := &http.Server{
        Addr:         "0.0.0.0:8080",
        // Good practice to set timeouts to avoid Slowloris attacks.
        WriteTimeout: time.Second * 15,
        ReadTimeout:  time.Second * 15,
        Handler: r, // Pass our instance of gorilla/mux in.
    }
	
	log.Fatal(srv.ListenAndServe())

}