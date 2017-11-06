package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

// Results for a perf thread
type Result struct {
	total   time.Duration
	average time.Duration
	channel int
}

//display method for Results
func (r *Result) display() {
	fmt.Println("Channel(", r.channel, ") Total(", r.total, ") Average(", r.average, ")")
}

// This is the input structure for a perf thread
type Input struct {
	url        string
	threads    int
	iterations int
	output     int
	index      int // Also the channel number
}

// This is the main url perf testing method
func (input Input) run(done chan Result) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", input.url, nil)
	req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Mobile Safari/537.36")
	start := time.Now()
	for i := 0; i < input.iterations; i++ {
		client.Do(req)
		if i%input.output == 0 {
			fmt.Println("Thread: ", input.index, " iteration: ", i)
		}
	}
	end := time.Now()
	total := end.Sub(start)
	average := total / time.Duration(input.iterations)

	// Send results on done channel
	done <- Result{
		total:   total,
		average: average,
		channel: input.index,
	}
}

func main() {
	// Setup comment line parameters
	threads := flag.Int("threads", 10, "Number of Threads")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	flag.Parse()

	fmt.Println("Running again url:", *url)
	fmt.Println("threads: ", *threads)

	//Define the channel that will syncronize and wait before exiting
	done := make(chan Result, 1)

	//You want to make a copy when you pass this into the method so the url can change
	input := Input{
		iterations: *iterations,
		threads:    *threads,
		url:        *url,
		output:     5,
	}

	for i := 0; i < *threads; i++ {
		input.index = i + 1
		go input.run(done)
	}

	//Wait on all the threads
	for i := 0; i < *threads; i++ {
		r := <-done
		r.display()
	}
}
