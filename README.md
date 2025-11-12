# Todosay (Scheduler)

Описание проекта

Проект "Todosay" — это простой планировщик задач с REST API и небольшим фронтендом. Позволяет создавать, просматривать, обновлять и удалять задачи, устанавливать повторяющиеся задачи и помечать задачи выполненными. Хранение данных реализовано в SQLite базе данных (`pkg/db/scheduler.db`).

- *Аутентификация* — реализован вход по токену JWT для защиты операций удалённого доступа.
- *Локальное логирование* — логируется активность сервера в файл `full.log`.

Задания со звездочкой

- Выполнены все, за исключением Docker`a, контейнерную сборку не настраивал.


Команды для запуска

1. Создать .env

    TODO_PORT=7540
    TODO_DBFILE=../pkg/db/scheduler.db
    TODO_DBSCRIPT=../pkg/db/scheduler.sql
    TODO_PASSWORD=12345 

2. Сервер и тесты:
    go run main.go
    go test ./tests

    TODO_PORT=7540 TODO_PASSWORD=12345 TODO_DBFILE="../server/scheduler.db" go run main.go 
    TODO_DBFILE="../server/scheduler.db" go test ./tests
   по умолчанию БД находится в "/pkg/db/scheduler.db"

3. Откройте в браузере: http://localhost:7540

Примеры использования

- Обращение к фронтенду: http://localhost:7540/
- API endpoints находятся под тем же доменом, например POST /api/addTask
