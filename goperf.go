package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

func main() {
	threads := flag.Int("threads", 10, "Number of Threads")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	flag.Parse()

	fmt.Println("Running again url:", *url)
	fmt.Println("threads: ", *threads)

	//Define the channel that will syncronize and wait before exiting
	done := make(chan time.Duration, 1)

	for i := 0; i < *threads; i++ {
		go run(*url, i, *iterations, done)

	}

	//Wait on all the threads
	for i := 0; i < *threads; i++ {
		fmt.Println(<-done)
	}
}

func run(url string, index int, iterations int, done chan time.Duration) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Mobile Safari/537.36")
	start := time.Now()
	for i := 0; i < iterations; i++ {
		client.Do(req)
		if i%1 == 0 {
			fmt.Println("Thread: ", index, " iteration: ", i)
		}
	}
	end := time.Now()
	total := end.Sub(start)
	fmt.Println("Thread(", index, ") Total: ", total)
	average := total / time.Duration(iterations)
	fmt.Println("Avg: ", total/time.Duration(iterations))

	done <- average
}
