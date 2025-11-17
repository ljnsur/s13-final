package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ljnsur/todosay/pkg/constants"
	applog "github.com/ljnsur/todosay/pkg/log"
	_ "modernc.org/sqlite"
)

const Limit = 50

type DBTasks struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *DBTasks) (int64, error) {
	applog.Printf("AddTask: попытка добавить задачу title=%q date=%s repeat=%q", task.Title, task.Date, task.Repeat)
	if DB == nil {
		return 0, fmt.Errorf("база данных не инициализирована")
	}

	var id int64

	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES ($data, $title, $comment, $repeat)"
	res, err := DB.Exec(query, sql.Named("data", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		applog.Printf("AddTask: ошибка выполнения INSERT: %v", err)
		return 0, err
	}

	id, err = res.LastInsertId()
	if err != nil {
		applog.Printf("AddTask: ошибка получения LastInsertId: %v", err)
		return 0, err
	}
	applog.Printf("AddTask: задача сохранена id=%d title=%q", id, task.Title)
	return id, err
}

func ShowTasks(search string, limit int) ([]*DBTasks, error) {
	applog.Printf("ShowTasks: поиск=%q limit=%d", search, limit)

	if limit <= 0 || limit > 50 {
		limit = Limit
	}

	var query string
	var args []interface{}

	base := "SELECT id, date, title, comment, repeat FROM scheduler"

	if search == "" {
		query = base + " ORDER BY date ASC LIMIT $limit"
		args = append(args, sql.Named("limit", limit))

	}
	if isDate(search) {
		parsed, err := time.Parse("02.01.2006", search)
		if err != nil {
			applog.Printf("ShowTasks: неверная дата поиска %q: %v", search, err)
			return nil, err
		}
		date := parsed.Format(constants.TimeFormat)

		query = base + " WHERE date = $date ORDER BY date ASC LIMIT $limit"
		args = append(args, sql.Named("date", date), sql.Named("limit", limit))

	}
	if (search != "") && (!isDate(search)) {
		like := "%" + search + "%"
		query = base + " WHERE title LIKE $like1 OR comment LIKE $like2 ORDER BY date ASC LIMIT $limit"
		args = append(args, sql.Named("like1", like), sql.Named("like2", like), sql.Named("limit", limit))
	}

	if DB == nil {
		return nil, fmt.Errorf("база данных не инициализирована")
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		applog.Printf("ShowTasks: ошибка выполнения запроса: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []*DBTasks
	for rows.Next() {
		var t DBTasks
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			applog.Printf("ShowTasks: ошибка сканирования строки: %v", err)
			return nil, err
		}
		tasks = append(tasks, &t)
	}

	if err := rows.Err(); err != nil {
		applog.Printf("ShowTasks: ошибка итерации rows: %v", err)
		return nil, err
	}

	if tasks == nil {
		tasks = []*DBTasks{}
	}

	applog.Printf("ShowTasks: найдено %d задач", len(tasks))
	return tasks, nil
}

func isDate(s string) bool {
	_, err := time.Parse("02.01.2006", s)
	return err == nil
}

func GetTask(id string) (*DBTasks, error) {
	if id == "" {
		return nil, fmt.Errorf("пустой id")
	}
	applog.Printf("GetTask: получение задачи id=%s", id)

	if DB == nil {
		return nil, fmt.Errorf("база данных не инициализирована")
	}

	var t DBTasks

	err := DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = $id", sql.Named("id", id)).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)

	if err == sql.ErrNoRows {
		applog.Printf("GetTask: задача не найдена id=%s", id)
		return nil, fmt.Errorf("задача не найдена")
	}
	if err != nil {
		applog.Printf("GetTask: ошибка запроса id=%s: %v", id, err)
		return nil, err
	}

	applog.Printf("GetTask: задача получена id=%s title=%q", t.ID, t.Title)
	return &t, nil

}

func UpdateTask(task *DBTasks) error {
	if task.ID == "" {
		return fmt.Errorf("id отсутвует")
	}
	applog.Printf("UpdateTask: обновление задачи id=%s title=%q", task.ID, task.Title)
	if DB == nil {
		return fmt.Errorf("база данных не инициализирована")
	}

	query := "UPDATE scheduler SET date = $date, title = $title, comment = $comment, repeat = $repeat WHERE id = $id"

	res, err := DB.Exec(query, sql.Named("date", task.Date), sql.Named("title", task.Title), sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat), sql.Named("id", task.ID))
	if err != nil {
		applog.Printf("UpdateTask: ошибка выполнения UPDATE id=%s: %v", task.ID, err)
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		applog.Printf("UpdateTask: ошибка получения RowsAffected: %v", err)
		return err
	}

	if count == 0 {
		applog.Printf("UpdateTask: задача не найдена id=%s", task.ID)
		return fmt.Errorf("задача не найдена")
	}

	applog.Printf("updateTask: задача обновлена id=%s title=%q", task.ID, task.Title)
	return nil
}

func DeleteTask(id string) error {
	if id == "" {
		return fmt.Errorf("id отсутвует")
	}
	applog.Printf("DeleteTask: удаление задачи id=%s", id)
	if DB == nil {
		return fmt.Errorf("база данных не инициализирована")
	}

	_, err := DB.Exec("DELETE FROM scheduler WHERE id = $id", sql.Named("id", id))
	if err != nil {
		applog.Printf("DeleteTask: ошибка выполнения DELETE id=%s: %v", id, err)
		return err
	}

	applog.Printf("DeleteTask: задача удалена id=%s", id)
	return nil
}

func UpdateDate(date, id string) error {

	if date == "" {
		return fmt.Errorf("date отсутвует")
	}
	applog.Printf("UpdateDate: обновление даты id=%s date=%s", id, date)
	if DB == nil {
		return fmt.Errorf("база данных не инициализирована")
	}

	query := "UPDATE scheduler SET date = $date WHERE id = $id"

	res, err := DB.Exec(query, sql.Named("date", date), sql.Named("id", id))
	if err != nil {
		applog.Printf("UpdateDate: ошибка выполнения UPDATE id=%s: %v", id, err)
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		applog.Printf("UpdateDate: ошибка получения RowsAffected: %v", err)
		return err
	}

	if count == 0 {
		applog.Printf("UpdateDate: задача не найдена id=%s", id)
		return fmt.Errorf("задача не найдена")
	}

	applog.Printf("UpdateDate: успешно обновлена дата id=%s date=%s", id, date)
	return nil
}
