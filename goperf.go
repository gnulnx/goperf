package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

type Result struct {
	total time.Duration
	average time.Duration
	channel int
}

func main() {
	threads := flag.Int("threads", 10, "Number of Threads")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	flag.Parse()

	fmt.Println("Running again url:", *url)
	fmt.Println("threads: ", *threads)

	//Define the channel that will syncronize and wait before exiting
	//done := make(chan time.Duration, 1)
	done := make(chan Result, 1)

	for i := 0; i < *threads; i++ {
		go run(*url, i, *iterations, done)
	}

	//Wait on all the threads
	for i := 0; i < *threads; i++ {
		r := <-done
		fmt.Println("Channel(" , r.channel , ") Total(", r.total, ") Average(", r.average, ")")
	}
}

func run(url string, index int, iterations int, done chan Result) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Mobile Safari/537.36")
	start := time.Now()
	for i := 0; i < iterations; i++ {
		client.Do(req)
		if i % 5 == 0 {
			fmt.Println("Thread: ", index, " iteration: ", i)
		}
	}
	end := time.Now()
	total := end.Sub(start)
	average := total / time.Duration(iterations)

	// Send results on done channel
	done <- Result {
		total: total,
		average: average,
		channel: index,
	}
}
