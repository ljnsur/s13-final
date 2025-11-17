package api

import (
	"net/http"
	"strings"
	"time"

	applog "github.com/ljnsur/todosay/pkg/log"

	"github.com/ljnsur/todosay/pkg/db"
)

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	applog.Printf("doneTask: начало обработки %s %s", r.Method, r.RemoteAddr)
	if r.Method != http.MethodPost {
		applog.Printf("doneTask: неверный метод %s", r.Method)
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "разрешены только POST-запросы"})
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		applog.Printf("doneTask: id не указан")
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "задача не найдена"})
		return
	}

	applog.Printf("doneTask: обработка задачи id=%s", id)
	task, err := db.GetTask(id)
	if err != nil {
		applog.Printf("doneTask: ошибка получения задачи id=%s: %v", id, err)
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Одноразовая задача → удаляем
	if strings.TrimSpace(task.Repeat) == "" {
		applog.Printf("doneTask: одноразовая задача, удаление id=%s", id)
		if err := db.DeleteTask(id); err != nil {
			applog.Printf("doneTask: ошибка удаления id=%s: %v", id, err)
			writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJson(w, http.StatusOK, map[string]any{})
		return
	}

	// Периодическая → обновляем дату
	nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		applog.Printf("doneTask: ошибка вычисления следующей даты для id=%s: %v", id, err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if nextDate == "" {
		applog.Printf("doneTask: следующая дата не назначена для id=%s: %v", id, err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "следующая дата не назначена"})
		return
	}
	if err := db.UpdateDate(nextDate, id); err != nil {
		applog.Printf("doneTask: ошибка обновления даты id=%s next=%s: %v", id, nextDate, err)
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, http.StatusOK, map[string]any{})
}
