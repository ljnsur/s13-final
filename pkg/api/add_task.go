package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ljnsur/todosay/pkg/constants"
	"github.com/ljnsur/todosay/pkg/db"
	applog "github.com/ljnsur/todosay/pkg/log"
)

// taskHandler маршрутизует HTTP-запросы для работы с задачами по HTTP-методу.
// Поддерживаемые методы:
//   - POST:  добавление задачи (вызывает addTaskHandler). Ожидает JSON в теле.
//   - GET:   получение задачи по id (вызывает getTaskHandler). Ожидается параметр id в query.
//   - PUT:   обновление задачи (вызывает updateTaskHandler). Ожидает JSON в теле.
//   - DELETE: удаление задачи по id (вызывает deleteTaskHandler). Ожидается параметр id в query.
//
// Для прочих методов возвращает ошибку 400 Bad Request.
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, http.ErrAbortHandler.Error(), http.StatusMethodNotAllowed)
		return
	}
}

// addTaskHandler обрабатывает POST-запросы для добавления новой задачи.
// Ожидается, что в теле запроса будет передан JSON с данными задачи.
func addTaskHandler(w http.ResponseWriter, r *http.Request) {

	var buf bytes.Buffer
	var task db.DBTasks

	if _, err := buf.ReadFrom(r.Body); err != nil {
		applog.Printf("addTask: ошибка чтения тела запроса: %v", err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "ошибка чтения тела запроса"})
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		applog.Printf("addTask: неверный JSON: %v (body=%q)", err, buf.String())
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "неверный формат JSON"})
		return
	}

	if len(task.Title) == 0 {
		applog.Printf("addTask: отсутствует название задачи (body=%q)", buf.String())
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "отсутствует название задачи"})
		return
	}

	if err := ValidateAndNormalizeTask(&task); err != nil {
		applog.Printf("addTask: некорректные данные задачи: %v (task=%+v)", err, task)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "некорректные данные задачи: " + err.Error()})
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		applog.Printf("addTask: ошибка сохранения в БД: %v (task=%+v)", err, task)
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "ошибка сервера при сохранении задачи"})
		return
	}

	idS := strconv.Itoa(int(id))
	applog.Printf("addTask: задача создана id=%s title=%q", idS, task.Title)
	writeJson(w, http.StatusOK, map[string]string{"id": idS})
}

// getTaskHandler обрабатывает GET-запросы для получения задачи по её идентификатору.
// Ожидается, что в query-параметрах будет указан параметр "id" с идентификатором задачи.
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		applog.Printf("getTask: отсутствует идентификатор")
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	if _, err := strconv.Atoi(id); err != nil {
		applog.Printf("getTask: id не является числом: %q (%v)", id, err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "id не является числом"})
		return
	}

	applog.Printf("getTask: запрошена задача id=%s", id)
	task, err := db.GetTask(id)
	if err != nil {
		applog.Printf("getTask: ошибка получения задачи id=%s: %v", id, err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, http.StatusOK, task)
}

// updateTaskHandler обрабатывает PUT-запросы для обновления задачи.
// Ожидается, что в теле запроса будет передан JSON с данными задачи,
// включая её идентификатор (ID).
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var input db.DBTasks
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		applog.Printf("updateTask: ошибка чтения тела запроса: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &input); err != nil {
		applog.Printf("updateTask: неверный JSON: %v (body=%q)", err, buf.String())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	applog.Printf("updateTask: входящие данные id=%q title=%q date=%q repeat=%q", input.ID, input.Title, input.Date, input.Repeat)

	if input.ID == "" {
		applog.Printf("updateTask: отсутствует идентификатор")
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	if input.Title == "" {
		applog.Printf("updateTask: отсутствует название задачи (id=%s)", input.ID)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указано название задачи"})
		return
	}

	if input.Date == "" {
		input.Date = time.Now().Format(constants.TimeFormat)
	}

	t, err := time.Parse(constants.TimeFormat, input.Date)
	if err != nil {
		applog.Printf("updateTask: неверный формат даты %q: %v", input.Date, err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if input.Repeat == "" {
		input.Repeat = ""
	} else if _, err := NextDate(time.Now(), t.Format(constants.TimeFormat), input.Repeat); err != nil {
		applog.Printf("updateTask: неверный формат повтора %q: %v", input.Repeat, err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неправильный формат повтора"})
		return
	}
	// Обновляем задачу в БД
	if err := db.UpdateTask(&input); err != nil {
		applog.Printf("updateTask: задача не найдена или ошибка обновления id=%s: %v", input.ID, err)
		writeJson(w, http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		return
	}

	writeJson(w, http.StatusOK, map[string]interface{}{})
}

// deleteTaskHandler обрабатывает DELETE-запросы для удаления задачи по её идентификатору.
// Ожидается, что в query-параметрах будет указан параметр "id" с идентификатором задачи.
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		applog.Printf("deleteTask: id не указан")
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "id не указан"})
		return
	}

	if _, err := strconv.Atoi(id); err != nil {
		applog.Printf("deleteTask: id не является числом: %q (%v)", id, err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "id не является числом"})
		return
	}

	applog.Printf("deleteTask: удаление задачи id=%s", id)
	if err := db.DeleteTask(id); err != nil {
		applog.Printf("deleteTask: ошибка удаления id=%s: %v", id, err)
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, http.StatusOK, map[string]any{})
}

// ValidateAndNormalizeTask валидирует и нормализует задачу:
// - если дата пустая → сегодня
// - если дата в прошлом → переносим на ближайшую по правилу (или сегодня, если одноразовая)
// - валидирует правило повторения (если есть)
func ValidateAndNormalizeTask(task *db.DBTasks) error {
	now := time.Now()
	normalizedNow := now.Format(constants.TimeFormat)

	// Если дата не указана — ставим сегодня
	if task.Date == "" {
		task.Date = normalizedNow
		if task.Repeat == "" {
			return nil
		}
		// Проверяем, валидно ли правило (даже если не будем переносить)
		_, err := NextDate(now, normalizedNow, task.Repeat)
		return err
	}

	// Парсим дату задачи
	t, err := time.Parse(constants.TimeFormat, task.Date)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}

	// Если дата уже прошла (или сегодня, но с повторением и нужно перепрыгнуть)
	if !t.After(now) {
		if task.Repeat == "" {
			// Одноразовая задача в прошлом → переносим на сегодня
			task.Date = normalizedNow
			return nil
		}
		// Если задача сегодняшняя с правилом (и добавляется впервые)
		if t.Format(constants.TimeFormat) == now.Format(constants.TimeFormat) {
			task.Date = now.Format(constants.TimeFormat)
			return nil
		}

		// Периодическая задача в прошлом или сегодня → переносим на следующую дату
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err // например, невалидное правило
		}
		task.Date = next
		return nil
	}

	// Дата в будущем — просто проверяем валидность правила (если есть)
	if task.Repeat != "" {
		_, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	return nil
}
