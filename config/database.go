package config

import (
    "database/sql"
    _ "github.com/lib/pq"
    "log"
)

func InitDB() *sql.DB {
    connStr := "user=postgres password=Erasyl2007erasyl dbname=banking_db sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    return db
} 