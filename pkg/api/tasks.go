package api

import (
	"net/http"

	applog "github.com/ljnsur/todosay/pkg/log"

	"github.com/ljnsur/todosay/pkg/db"
)

const DefaultLimit = 50

type TasksResp struct {
	Tasks []*db.DBTasks `json:"tasks"`
}

// showTasksHandler обрабатывает GET-запросы для получения списка задач.
// Поддерживается необязательный параметр "search" для фильтрации задач по ключевому слову в названии или описании.
func showTasksHandler(w http.ResponseWriter, r *http.Request) {
	applog.Printf("showTasks: начало обработки %s %s", r.Method, r.RemoteAddr)
	if r.Method != http.MethodGet {
		applog.Printf("showTasks: неверный метод %s", r.Method)
		http.Error(w, "разрешены только GET запросы", http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")

	tasks, err := db.ShowTasks(search, DefaultLimit)
	if err != nil {
		applog.Printf("showTasks: ошибка получения списка задач: %v", err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	applog.Printf("showTasks: найдено %d задач", len(tasks))
	writeJson(w, http.StatusOK, TasksResp{Tasks: tasks})
}
