package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db  *sql.DB
	numWorkers int = 25
	workersCurrentTask = make(map[int]string)
	workersCrawlTime = make(map[int]time.Duration)
)

func init() {
	database, err := initDB()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	db = database
}

func worker(id int, jobs <-chan string, done chan<- bool) {
    for url := range jobs {
		workersCurrentTask[id] = url
		startTime := time.Now()
        crawl(url, -1)
		endTime := time.Now()
		totalTime := endTime.Sub(startTime)
		workersCrawlTime[id] = totalTime
    }
    done <- true
}

func main() {
    defer cleanup()

	ticker := ticker()
	go ticker()

    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    jobs := make(chan string, numWorkers)
    done := make(chan bool, numWorkers)

    for w := 1; w <= numWorkers; w++ {
        go worker(w, jobs, done)
    }

    go func() {
        for {
            url, err := next()
            if err != nil {
                fmt.Println(err)
                close(jobs)
                return
            }
            jobs <- url
        }
    }()

    <-c
    fmt.Println("\nReceived an interrupt, stopping services...")
    close(jobs)

    for a := 1; a <= numWorkers; a++ {
        <-done
    }

    fmt.Println("Finished all jobs")
}