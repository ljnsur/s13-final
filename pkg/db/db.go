package db

import (
	"database/sql"
	"fmt"
	"os"

	applog "github.com/ljnsur/todosay/pkg/log"

	_ "modernc.org/sqlite"
)

var (
	DB     *sql.DB
	DBPath string
)

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

	// Открываем пул соединений
	DB, err = sql.Open("sqlite", DBPath)
	if err != nil {
		applog.Printf("Init: база не найдена, %s", err.Error())
		return err
	}

	DB.SetMaxOpenConns(2)
	DB.SetMaxIdleConns(2)
	DB.SetConnMaxLifetime(0)

	if err = DB.Ping(); err != nil {
		DB.Close()
		DB = nil
		applog.Printf("Init: соединение с БД не установлено %s", err.Error())
		return err
	}

	applog.Printf("Init: пул соединений успешно открыт")
	return nil

}

func Close() error {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			applog.Printf("Close: ошибка закрытия пула соединений: %v", err)
			return err
		}
		DB = nil
		applog.Printf("Close: пул соединений закрыт")
	}
	return nil
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
