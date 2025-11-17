package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ljnsur/todosay/pkg/api"
	"github.com/ljnsur/todosay/pkg/db"
	"github.com/ljnsur/todosay/server"

	_ "github.com/joho/godotenv/autoload"
	applog "github.com/ljnsur/todosay/pkg/log"
)

func main() {
	file, err := os.OpenFile("full.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("ошибка при открытии лог-файла: %s", err)
	}
	defer file.Close()

	fullLog := log.New(file, "INFO: ", log.Lshortfile|log.LstdFlags)

	applog.SetLogger(fullLog)

	api.InitAuth()

	// Инициализация базы данных, открытие пула соединений
	err = db.Init()
	if err != nil {
		fullLog.Fatalf("Ошибка инициализации БД: %v", err)
	}
	// Запуск сервера
	webDir := "web"
	server.Run(webDir, fullLog)

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	fullLog.Println("Получен сигнал завершения.")

	// Закрытие базы данных
	if err := db.Close(); err != nil {
		fullLog.Printf("Ошибка закрытия БД: %v", err)
	} else {
		fullLog.Println("БД закрыта")
	}

	fullLog.Println("Приложение завершено")

}
