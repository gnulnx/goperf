package main

import (
	"flag"
	"fmt"
	"github.com/gnulnx/goperf/request"
	"os"
	"time"
)

func main() {
	// Setup comment line parameters
	threads := flag.Int("connections", 10, "Number of concurrant connections")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	output := flag.Int("output", 5, "Show thread output every {n} iterations")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	increment := flag.Int("incr", 2, "How fast to increment the number of concurrant connections")
	max_response := flag.Int("max-response", 1, "Maximun number of seconds to wait for a response")
	flag.Parse()

	//You want to make a copy when you pass this into the method so the url can change
	input := request.Input{
		Iterations: *iterations,
		Threads:    *threads,
		Url:        *url,
		Output:     *output,
		Verbose:    *verbose,
	}

	fmt.Println("Running again url:", *url)
	fmt.Println("Concurrant Connections: ", *threads, "Sustained for: ", input.Threads)

	for i := 0; i < 1; i++ {
		input.Threads += *increment
		total, avg := perf(input)
		fmt.Println("Concurrant Connections:", *threads, "Sustained for:", input.Threads)
		fmt.Println("Total Time: ", total)
		fmt.Println("Average Request time: ", avg)
		if avg > time.Duration(1000*1000*1000**max_response) {
			fmt.Println(time.Duration(1000 * 1000 * 1000 * *max_response))
			os.Exit(1)
		}
	}
	os.Exit(1)
}

/*
	This method is the basic unit of perf testing.
	It takes an input object with run parameters and it returns
	the total time and the average time for all the requests to run
*/
func perf(input request.Input) (time.Duration, time.Duration) {
	//Define the channel that will syncronize and wait before exiting
	done := make(chan request.Result, 1)

	/* Start all concurrant request threads */
	for i := 0; i < input.Threads; i++ {
		input.Index = i + 1
		go input.Run(done)
	}

	/* Wait on all the threads to return and collect results */
	total := make([]time.Duration, 0, input.Threads)
	for i := 0; i < input.Threads; i++ {
		r := <-done

		total = append(total, r.Average)
		if input.Verbose {
			r.Display()
		}
	}

	sum := time.Duration(0)

	for i := range total {
		sum += total[i]
	}

	avg := sum / time.Duration(len(total))

	return sum, avg
}
