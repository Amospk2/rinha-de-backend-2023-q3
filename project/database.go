package database

import (
	"context"
	"fmt"
	"sync"
	"github.com/Amospk2/rinha-de-backend-2023-q3/project/model"
	"github.com/jackc/pgx/v5/pgxpool"
)


var (
	db *pgxpool.Pool
	once     sync.Once
)


funct warmUp() {

	var ids []string
	for i := 0; i < 10; i++ {
		person := &Person{
			ID: 	uuidV,_ := uuid.NewString()
			Nome: fmt.Sprintf("apelido-%d", i),
			Apelido: fmt.Sprintf("Nome-%d", i),
			Nascimento: "1970-01-01",
			Stack: []string{"python", "java"},
		}
		ids = append(ids, person.ID)
		_, err := db.Exec(
			context.Background(),
			`INSERT INTO Pessoa(id, apelido, nome, nascimento, stack)
			values ($1, $2, $3, $4, $5)`,
			pessoa.ID, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento, pessoa.Stack,
		)

		if err != nil {
			fmt.Println( err)
		}
	}

	for _, id := range ids {
		_, err := db.Exec(
			context.Background(),
			"DELETE FROM people WHERE id = $1",
			id,
		)

		if err != nil {
			fmt.Println( err)
		}
	}

	if err := db.Ping(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func connect(config *config.Config) *pgxpool.Pool {

	once.Do(func(){

		poolConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
		if err != nil {
			fmt.Println("Unable to parse connection url:", err)
		}

		db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			fmt.Println("Unable to create connection pool:", err)
		}

		warmUp()

	})

	return db
}