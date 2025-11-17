CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(64) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(64) NOT NULL DEFAULT ""
);
CREATE INDEX scheduler_date ON scheduler (date);