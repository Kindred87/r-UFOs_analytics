package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"ufo_collector/dropbox"

	_ "github.com/mattn/go-sqlite3"
)

const (
	analyticsFile = "analytics.db"
)

type SQLCommand struct {
	Query string
	Args  []interface{}
}

var (
	DB         *sql.DB
	lock       sync.Mutex
	sqlChannel = make(chan SQLCommand, 200)
	ctx        context.Context
	cancel     context.CancelFunc
)

func InitDB() error {
	lock.Lock()
	defer lock.Unlock()

	ctx, cancel = context.WithCancel(context.Background())

	_, err := os.Stat(analyticsFile)
	if os.IsNotExist(err) {
		err = createDBFile()
		if err != nil {
			return err
		}
	}

	db, err := sql.Open("sqlite3", analyticsFile)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if db == nil {
		return sql.ErrConnDone
	}

	DB = db
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	go sqlWorker(ctx)

	return nil
}

func sqlWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case cmd := <-sqlChannel:
			lock.Lock()
			stmt, err := DB.Prepare(cmd.Query)
			if err != nil {
				fmt.Printf("error preparing statement: %v\n", err)
				lock.Unlock()
				continue
			}

			_, err = stmt.Exec(cmd.Args...)
			if err != nil {
				fmt.Printf("error executing statement: %v\n", err)
			}

			stmt.Close()
			lock.Unlock()
		case <-ticker.C:
			log.Printf("Number of write tasks in the channel: %v\n", len(sqlChannel))
		case <-ctx.Done():
			return
		}
	}
}

func createDBFile() error {
	file, err := os.Create(analyticsFile) // Create SQLite file if it does not exist
	if err != nil {
		return fmt.Errorf("error creating database file: %w", err)
	}
	file.Close()

	db, err := sql.Open("sqlite3", analyticsFile)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if db == nil {
		return sql.ErrConnDone
	}

	DB = db
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	// Create table if the database file was just created
	return createTable()
}

func CloseDB() error {
	lock.Lock()
	defer lock.Unlock()

	cancel()

	if DB != nil {
		return DB.Close()
	}
	return nil
}

func SyncDB() error {
	lock.Lock()
	defer lock.Unlock()

	err := dropbox.Upload(analyticsFile, "/"+analyticsFile)
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}

	return nil
}

func createTable() error {
	lock.Lock()
	defer lock.Unlock()

	createTableSQL := `CREATE TABLE post_history (
		"id" TEXT NOT NULL PRIMARY KEY,		
		"post_time" DATETIME,
		"flair" TEXT,
		"url" TEXT,
		"author" TEXT  // Add this line
		"num_comments" INTEGER
	);`

	stmt, err := DB.Prepare(createTableSQL)
	if err != nil {
		return fmt.Errorf("error preparing create table statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	return err
}

type PostHistory struct {
	ID       string
	PostTime string
	Flair    string
	URL      string
	Author   string
}

func AddPostHistory(id, postTime, flair, url, author string, numComments int) error {
	sqlChannel <- SQLCommand{
		Query: `INSERT INTO post_history(id, post_time, flair, url, author, num_comments) VALUES(?, ?, ?, ?, ?, ?)`,
		Args:  []interface{}{id, postTime, flair, url, author, numComments},
	}

	return nil
}
