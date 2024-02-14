package database

import (
	"api/model"
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	db   *pgxpool.Pool
	once sync.Once
)

func warmUp() {

	var ids []string

	_, err := db.Exec(
		context.Background(),
		"DELETE FROM Pessoa",
	)

	fmt.Println("Start warmup")
	for i := 0; i < 10; i++ {
		pessoa := model.NewPessoa(
			uuid.NewString(),
			fmt.Sprintf("apelido-%d", i),
			fmt.Sprintf("Nome-%d", i),
			"1970-01-01",
			[]string{"python", "java"},
		)
		ids = append(ids, pessoa.ID)
		_, err = db.Exec(
			context.Background(),
			`INSERT INTO Pessoa(id, apelido, nome, nascimento, stack)
			values ($1, $2, $3, $4, $5)`,
			pessoa.ID, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento, pessoa.Stack,
		)

		if err != nil {
			fmt.Println(err)
		}
	}

	for _, id := range ids {
		_, err = db.Exec(
			context.Background(),
			"DELETE FROM Pessoa WHERE id = $1",
			id,
		)

		if err != nil {
			fmt.Println(err)
		}
	}

	if err := db.Ping(context.Background()); err != nil {
		fmt.Println(err)
	}

	fmt.Println("finishing warmup")
}

func Connect() *pgxpool.Pool {

	once.Do(func() {

		poolConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
		if err != nil {
			fmt.Println("Unable to parse connection url:", err)
		}

		db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			fmt.Println("Unable to create connection pool:", err)
		}

		fmt.Println("Connect with success")

		warmUp()

	})

	return db
}

func GetPessoaByTermInPostgres(term string, pool *pgxpool.Pool) []model.Pessoa {

	var pessoas []model.Pessoa

	rows, err := pool.Query(context.Background(),
		`
		select id, nome, apelido, to_char(nascimento, 'yyyy-mm-dd'), stack 
		from pessoa where pessoa.termo_busca ILIKE '%`+term+`% LIMIT 50;'
		`,
	)

	if err != nil {
		return []model.Pessoa{}
	}

	defer rows.Close()

	for rows.Next() {
		var pessoa model.Pessoa

		err = rows.Scan(&pessoa.ID, &pessoa.Nome, &pessoa.Apelido, &pessoa.Nascimento, &pessoa.Stack)

		if err != nil {
			log.Fatal(err)
		}

		pessoas = append(pessoas, pessoa)
	}

	return pessoas

}
