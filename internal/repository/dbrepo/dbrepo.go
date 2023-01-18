package dbrepo

import (
	"database/sql"
	"github.com/454270186/Hotel-booking-web-application/internal/config"
	"github.com/454270186/Hotel-booking-web-application/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(a *config.AppConfig, conn *sql.DB) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}
