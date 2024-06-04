package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	// "sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db  *sql.DB
	err error

	numWorkers int = 10
	workersCurrentTask = make(map[int]string)
	workersCrawlTime = make(map[int]time.Duration)
)

func init() {
	db, err = sql.Open("sqlite3", "./foo.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS domains (
            id INTEGER PRIMARY KEY, 
            domain TEXT UNIQUE
        )
    `); err != nil {
		fmt.Println(err)
		return
	}

	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS paths (
			id INTEGER PRIMARY KEY, 
			domain INTEGER, 
			path TEXT,
			secure BOOLEAN,
			scanned BOOLEAN,
			onHold BOOLEAN,
			FOREIGN KEY(domain) REFERENCES domains(id),
			UNIQUE(domain, path)
		)
    `); err != nil {
		fmt.Println(err)
		return
	}

	if _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS links (
            id INTEGER PRIMARY KEY,  
            parent INTEGER,
			child INTEGER,
            FOREIGN KEY(parent) REFERENCES paths(id)
			FOREIGN KEY(child) REFERENCES paths(id)
        )
    `); err != nil {
		fmt.Println(err)
		return
	}

	var domains int

	db.QueryRow(`
	SELECT COUNT(*) from domains
	`).Scan(&domains)

	if (domains == 0) {
		insert(true, "www.codegalaxy.co.uk", "/", -1)
	}

	fmt.Println("Successfully connected to the database.")
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

	ticker := time.NewTicker(1 * time.Second)
	first := true

	go func() {
        for range ticker.C {

			total, scanned, sites := stats()
			var totalCrawlsPerMinute float64 = 0
			var average time.Duration =  time.Now().Sub(time.Now())

			if !first {
				fmt.Printf("\033[%dA", 6)
				fmt.Print("\033[2K")

				var sum time.Duration
				for _, duration := range workersCrawlTime {
					sum += duration
				}

				if len(workersCrawlTime) != 0 {
					average = time.Duration(int64(sum) / int64(len(workersCrawlTime)))

					if average != 0 {
						averageInMinutes := float64(average) / float64(time.Minute)
						crawlsPerMinutePerWorker := 1 / averageInMinutes
						totalCrawlsPerMinute = crawlsPerMinutePerWorker * float64(numWorkers)
					}
				}

				


			} else {
				fmt.Println("")
			}

			fmt.Println("Workers:\t\t", numWorkers)
			fmt.Println("Pages crawled:\t\t", scanned)
			fmt.Println("Pages total:\t\t", total)
			fmt.Println("Domains total:\t\t", sites)
			fmt.Println("Average crawl time:\t", average)
			fmt.Println("Crawls p/m:\t\t", totalCrawlsPerMinute)


			first = false

        }
    }()

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
                //fmt.Println(err)
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