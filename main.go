package main

import (
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kzon/technopark-sem2-db/pkg/component/forum"
	forumRepository "github.com/kzon/technopark-sem2-db/pkg/component/forum/repository"
	"github.com/kzon/technopark-sem2-db/pkg/component/service"
	"github.com/labstack/echo"
	"log"
	"os"
)

const PORT = "5000"

func main() {
	e := echo.New()
	db, err := NewDB()
	if err != nil {
		log.Fatal(err)
	}

	forumRepo := forumRepository.NewRepository(db)
	forumUsecase := forum.NewUsecase(forumRepo)
	forum.NewHandler(e, forumUsecase)

	serviceRepo := service.NewRepository(db)
	serviceUsecase := service.NewUsecase(serviceRepo)
	service.NewHandler(e, serviceUsecase)

	log.Fatal(e.Start(":" + PORT))
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
