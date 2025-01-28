package main

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/valkyraycho/bank_project/utils"
)

const migrationURL = "file://db/migration"

func main() {
	cfg, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load environment variables")
	}
	runDBMigrations(migrationURL, cfg.DBSource)
}

func runDBMigrations(migrationURL, dbsource string) {
	migration, err := migrate.New(migrationURL, dbsource)
	if err != nil {
		log.Fatal("cannot create new migration instance: ", err)
	}
	if err := migration.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("database already migrated to the latest version")
			return
		}
		log.Fatal("failed to migrate")
		return
	}
	log.Println("database migrated successfully")
}
