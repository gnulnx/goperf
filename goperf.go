package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/gnulnx/goperf/request"
	"os"
	"strconv"
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
	// Is thre any reason this 'Input' struc has to live in requests?  I'm going with no
	input := request.Input{
		//var Input struct {
		Iterations: *iterations,
		Threads:    *threads,
		Url:        *url,
		Output:     *output,
		Verbose:    *verbose,
	}

	// Working on New Method:  Currently fetch url and assets.
	color.Green("~~ Fetching a single url and printing info ~~")
	resp := request.FetchAll(*url, false)
	printFetchAllResponse(resp)
	tmp, _ := json.MarshalIndent(resp, "", "    ")
	fmt.Println(string(tmp))

	// Old Method Below Here.  Likely not relevant any longer
	os.Exit(1)

	fmt.Println("Running again url:", *url)
	fmt.Println("Concurrant Connections: ", *threads, "Sustained for: ", input.Threads)

	for i := 0; i < 100; i++ {
		input.Threads += *increment
		total, avg := perf(input)
		fmt.Println("Concurrant Connections:", input.Threads, "Sustained for:", input.Iterations, " iterations")
		fmt.Println("Total Time: ", total)
		fmt.Println("Average Request time: ", avg)
		if avg > time.Duration(1000*1000*1000**max_response) {
			fmt.Println("Exitiing because we reached max response")
			fmt.Println(time.Duration(1000 * 1000 * 1000 * *max_response))
			os.Exit(1)
		}
	}
	os.Exit(1)
}

func printFetchAllResponse(resp *request.FetchAllResponse) {
	color.Red("Fetching: " + resp.BaseUrl.Url)

	//if output.Status == 200 {
	if resp.BaseUrl.Status == 200 {
		color.Green(" - Status: " + strconv.Itoa(resp.BaseUrl.Status))
	} else {
		color.Red(" - Status: " + strconv.Itoa(resp.BaseUrl.Status))
	}

	color.Yellow(" - Time to first byte: " + resp.BaseUrl.Time.String())
	color.Yellow(" - Bytes: " + strconv.Itoa(resp.BaseUrl.Bytes))
	color.Yellow(" - Runes: " + strconv.Itoa(resp.BaseUrl.Runes))

}

/*
	This method is the basic unit of perf testing.
	It takes an input object with run parameters and it returns
	the total time and the average time for all the requests to run
*/
func perf(input request.Input) (time.Duration, time.Duration) {
	// Define the channel that will syncronize and wait before exiting
	done := make(chan request.Result, 1)

	// Start all concurrant request threads
	for i := 0; i < input.Threads; i++ {
		input.Index = i + 1
		go input.Run(done)
	}

	// Wait on all the threads to return and collect results
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
