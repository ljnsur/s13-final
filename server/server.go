package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ljnsur/todosay/pkg/api"

	_ "github.com/joho/godotenv/autoload"
)

// Run запускает HTTP-сервер, обслуживающий статические файлы из webDir и API.
func Run(webDir string, loggerInfo *log.Logger) {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	config := &http.Server{
		Addr:         ":" + port,
		ErrorLog:     loggerInfo,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	api.Init()

	fmt.Printf("сервер запущен %s\n", fmt.Sprintf("http://localhost%s", config.Addr))

	err := config.ListenAndServe()
	if err != nil {
		loggerInfo.Fatalf("ошибка запуска сервера: %s", err)
	}

}
