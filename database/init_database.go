package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/joho/godotenv"
)

var ReadWriteDb, ReadOnlyDb *sql.DB
var ReadWriteDbConnectionErr, ReadOnlyDbConnectionErr error

func init() {
	mode := os.Getenv("CHINESE_LEARNING_APP_MODE")
	if "" == mode {
		mode = "develop"
	}

	godotenv.Load(".env." + mode)
	godotenv.Load()

	dbDir := os.Getenv("DB_DIRECTORY")
	_, dirDoesntExistError := os.Stat(dbDir)
	if dirDoesntExistError != nil {
		log.Fatal("DB_DIRECTORY in config is not a valid directory")
	}

	dbFileNamePrefix := os.Getenv("DB_FILE_NAME_PREFIX")
	if dbFileNamePrefix == "" {
		log.Fatal("DB_FILE_NAME_PREFIX is not given in config")
	}

	dbFilePath := dbDir + dbFileNamePrefix

	switch mode {
	case "production":
		dbFilePath += ".production.db"
	case "develop":
		// make it so that everytime air refreshes we use a new database? this can be done by having a dbFilePath dynamically generated based on time?
		dbFilePath += ".development.db"
	case "test":
		dbFilePath += ".testing.db"
	default:
		log.Fatal("Error, mode was not specified in OS environment")
	}

	_, fileDoesntExistErr := os.Stat(dbFilePath)
	if fileDoesntExistErr != nil {
		// the database file doesn't exist so create it and initialise it
		_, err := os.Create(dbFilePath)
		if err != nil {
			log.Fatal("Error creating database file")
		}
	} else {
		if mode == "test" || mode == "develop" {
			removeDatabaseError := os.Remove(dbFilePath)
			removeSharedMemoryError := os.Remove(dbFilePath + "-shm")
			removeWalError := os.Remove(dbFilePath + "-wal")
			if removeDatabaseError != nil {
				log.Fatal("Could not remove old test database")
			}
			if removeSharedMemoryError != nil {
				log.Fatal("Could not remove old test shared memory")
			}
			if removeWalError != nil {
				log.Fatal("Could not remove old test wal")
			}
		}
		_, err := os.Create(dbFilePath)
		if err != nil {
			log.Fatal("Error creating database file")
		}
	}
	// at this point the file must exist but if fileDoesntExistError that means its just created without being intialised

	ReadWriteDb, ReadWriteDbConnectionErr = sql.Open("sqlite3", dbFilePath+"?_journal=wal&busy_timeout=5000&_foreign_keys=on")
	if fileDoesntExistErr != nil || mode == "test" || mode == "develop" {
		// if the database file didn't exist initialise it
		initScript1, err := os.ReadFile(dbDir + "001_initial_schema.sql")
		if err != nil {
			log.Fatal("Unable to read sql initialiser script 1")
		}
		initScript2, err := os.ReadFile(dbDir + "002_initialise_words.sql")
		if err != nil {
			log.Fatal("Unable to read sql initialiser script 2")
		}

		_, err = ReadWriteDb.Exec(string(initScript1))
		if err != nil {
			log.Fatal("Unable to execute initialiser script 1")
		}
		_, err = ReadWriteDb.Exec(string(initScript2))
		if err != nil {
			log.Fatal("Unable to execute initialiser script 2")
		}
	}
	ReadOnlyDb, ReadOnlyDbConnectionErr = sql.Open("sqlite3", dbFilePath+"?_journal=wal&busy_timeout=5000&mode=ro&_foreign_keys=on")
}
