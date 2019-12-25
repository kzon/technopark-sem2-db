package main

import (
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kzon/technopark-sem2-db/pkg/api"
	"github.com/kzon/technopark-sem2-db/pkg/api/repository"
	"github.com/valyala/fasthttp"
	"log"
	"os"
)

const PORT = "5000"

func main() {
	db, err := NewDB()
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	usecase := api.NewUsecase(repo)
	handler := api.NewHandler(usecase)

	fmt.Println("listening port " + PORT)
	log.Fatal(fasthttp.ListenAndServe(":"+PORT, handler.GetHandleFunc()))
}

func NewDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", os.Getenv("POSTGRES_DSN"))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
