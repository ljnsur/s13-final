package api

import (
	"encoding/json"
	"net/http"
)

const (
	TIME_FORMAT = "20060102"
)

// Инициализация маршрутов API
func Init() {
	// Публичные маршруты (без авторизации)
	http.HandleFunc("/api/nextdate", nextDayHandler)
	http.HandleFunc("/api/signin", signInHandler)

	// Защищённые маршруты — оборачиваем в auth()
	http.HandleFunc("/api/task", auth(taskHandler))
	http.HandleFunc("/api/tasks", auth(showTasksHandler))
	http.HandleFunc("/api/task/done", auth(doneTaskHandler))
}

// Универсальная функция записи JSON-ответа
func writeJson(w http.ResponseWriter, status int, data any) {

	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	w.Write(resp)

}
