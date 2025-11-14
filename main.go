package main

import (
	"log"
	"os"

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

	FullLog := log.New(file, "INFO: ", log.Lshortfile|log.LstdFlags)

	applog.SetLogger(FullLog)

	err = db.Init()
	if err != nil {
		FullLog.Fatalf("%s", err)
	}

	api.InitAuth()

	webDir := "web"
	server.Run(webDir, FullLog)

}
