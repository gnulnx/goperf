package main

import (
	"flag"
	"fmt"
	"github.com/gnulnx/goperf/request"
)

func main() {
	// Setup comment line parameters
	threads := flag.Int("threads", 10, "Number of Threads")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	flag.Parse()

	fmt.Println("Running again url:", *url)
	fmt.Println("threads: ", *threads)

	//Define the channel that will syncronize and wait before exiting
	done := make(chan request.Result, 1)

	//You want to make a copy when you pass this into the method so the url can change
	input := request.Input{
		Iterations: *iterations,
		Threads:    *threads,
		Url:        *url,
		Output:     5,
	}

	for i := 0; i < *threads; i++ {
		input.Index = i + 1
		go input.Run(done)
	}

	//Wait on all the threads
	for i := 0; i < *threads; i++ {
		r := <-done
		r.Display()
	}
}
