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
	threads := flag.Int("threads", 10, "Number of Threads")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	output := flag.Int("output", 5, "Show thread output every {n} iterations")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	flag.Parse()

	fmt.Println("Running again url:", *url)
	fmt.Println("threads: ", *threads)

	//You want to make a copy when you pass this into the method so the url can change
	input := request.Input{
		Iterations: *iterations,
		Threads:    *threads,
		Url:        *url,
		Output:     *output,
		Verbose:    *verbose,
	}

	total, avg := perf(input)
	fmt.Println("Threads: ", input.Threads)
	fmt.Println("Total Time: ", total)
	fmt.Println("Average Request time: ", avg)

	for i := 0; i < 1000; i++ {
		input.Threads += 10
		total, avg = perf(input)
		fmt.Println("\nThreads: ", input.Threads)
		fmt.Println("Total Time: ", total)
		fmt.Println("Average Request time: ", avg)
		if avg > time.Duration(1000*1000*1000) {
			fmt.Println(time.Duration(1000 * 1000 * 1000))
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
