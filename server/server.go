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
	srv := &http.Server{
		Addr:         ":" + port,
		ErrorLog:     loggerInfo,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	api.Init()

	go func() {
		fmt.Printf("сервер запущен на http://localhost:%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerInfo.Fatalf("Run: сервер упал: %v", err)
		}
	}()

}
