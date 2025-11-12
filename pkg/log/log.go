package applog

import "log"

// Logger хранит глобальный логгер приложения. Устанавливается из cmd/main.go.
var Logger *log.Logger

// SetLogger задаёт глобальный логгер.
func SetLogger(l *log.Logger) {
	Logger = l
}

// Printf записывает сообщение через глобальный логгер, если он задан,
// иначе использует стандартный лог-пакет.
func Printf(format string, v ...interface{}) {
	if Logger != nil {
		Logger.Printf(format, v...)
		return
	}
	log.Printf(format, v...)
}
