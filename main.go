package main

import (
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kzon/technopark-sem2-db/pkg/component/forum"
	"github.com/kzon/technopark-sem2-db/pkg/component/user"
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

	userRepo := user.NewRepository(db)
	userUsecase := user.NewUsecase(userRepo)
	user.NewHandler(e, userUsecase)

	forumRepo := forum.NewRepository(db)
	forumUsecase := forum.NewUsecase(forumRepo, userRepo)
	forum.NewHandler(e, forumUsecase)

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
