package db

import (
	"database/sql"
	"fmt"
	"os"

	applog "github.com/ljnsur/todosay/pkg/log"

	_ "modernc.org/sqlite"
)

var DBPath string

const (
	DefaultDBSQL  = "../pkg/db/scheduler.sql"
	DefaultDBPath = "../pkg/db/scheduler.db"
)

func Init() error {
	applog.Printf("Init: проверка наличия файла БД %s", DBPath)
	if env := os.Getenv("TODO_DBFILE"); env != "" {
		DBPath = env
	} else {
		DBPath = DefaultDBPath
	}
	_, err := os.Stat(DBPath)
	if os.IsNotExist(err) {
		applog.Printf("Init: база не найдена, создаём из %s", DefaultDBSQL)
		createDB(DBPath, DefaultDBSQL)
		return nil
	}
	if err != nil {
		applog.Printf("Init: ошибка при проверке файла БД %s: %v", DBPath, err)
	} else {
		applog.Printf("Init: файл БД найден: %s", DBPath)
	}
	return err

}

func createDB(dbFile, dbSQL string) {
	applog.Printf("createDB: создание БД %s из %s", dbFile, dbSQL)
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		applog.Printf("createDB: ошибка при открытии базы данных: %v", err)
		panic(fmt.Errorf("ошибка при открытии базы данных: %w", err))
	}
	defer db.Close()

	sqlBytes, err := os.ReadFile(dbSQL)
	if err != nil {
		applog.Printf("createDB: ошибка чтения sql-файла %s: %v", dbSQL, err)
		panic(fmt.Errorf("ошибка чтения sql-файла: %w", err))
	}

	sqlStatments := string(sqlBytes)

	_, err = db.Exec(sqlStatments)
	if err != nil {
		applog.Printf("createDB: ошибка выполенения sql-команд: %v", err)
		panic(fmt.Errorf("ошибка выполенения sql-команд: %w", err))
	}

	applog.Printf("createDB: создана база данных %s", dbFile)
}
