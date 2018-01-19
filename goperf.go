package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/gnulnx/goperf/request"
	"os"
	"runtime/pprof"
	"strconv"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	// I ❤️  the way go handles command line arguments
	fetch := flag.Bool("fetch", false, "Fetch only the stats from url.  Do not return resources")
	fetchall := flag.Bool("fetchall", false, "Fetch all the resources and stats from -url")
	printjson := flag.Bool("printjson", false, "Print results as json")

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
		Iterations: *iterations,
		Threads:    *threads,
		Url:        *url,
		Output:     *output,
		Verbose:    *verbose,
	}

	// Working on New Method:  Currently fetch url and assets.
	if *fetch || *fetchall {
		f, _ := os.Create(*cpuprofile)
		pprof.StartCPUProfile(f)

		color.Green("~~ Fetching a single url and printing info ~~")
		resp := request.FetchAll(*url, *fetchall)

		if *printjson {
			tmp, _ := json.MarshalIndent(resp, "", "    ")
			fmt.Println(string(tmp))
		}

		printFetchAllResponse(resp)

		pprof.StopCPUProfile()
		os.Exit(1)
	}

	// Old Method Below Here.  Likely not relevant any longer
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
	color.Red("Base Url Results")

	if resp.BaseUrl.Status == 200 {
		color.Green(" - Status: " + strconv.Itoa(resp.BaseUrl.Status))
	} else {
		color.Red(" - Status: " + strconv.Itoa(resp.BaseUrl.Status))
	}

	total := resp.BaseUrl.Time
	color.Yellow(" - Url: " + resp.BaseUrl.Url)
	color.Yellow(" - Time to first byte: " + total.String())
	color.Yellow(" - Bytes: " + strconv.Itoa(resp.BaseUrl.Bytes))
	color.Yellow(" - Runes: " + strconv.Itoa(resp.BaseUrl.Runes))

	// This part will work for a single response
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// JSResponses   []FetchResponse
	calcTotal := func(resp []request.FetchResponse) time.Duration {
		total := time.Duration(0)
		for _, val := range resp {
			total += val.Time
		}
		return total
	}

	total += calcTotal(resp.JSResponses)
	total += calcTotal(resp.CSSResponses)
	total += calcTotal(resp.IMGResponses)

	color.Magenta(" - Total Time: %s", total.String())

	color.Red("JS Responses")
	for _, val := range resp.JSResponses {
		total += val.Time
		fmt.Printf(" - %-22s %-15s %-50s \n", green(val.Time.String()), yellow(strconv.Itoa(val.Bytes)), val.Url)
	}

	color.Red("CSS Responses")
	for _, val := range resp.CSSResponses {
		total += val.Time
		fmt.Printf(" - %-22s %-15s %-50s \n", green(val.Time.String()), yellow(strconv.Itoa(val.Bytes)), val.Url)
	}

	color.Red("IMG Responses")
	for _, val := range resp.IMGResponses {
		total += val.Time
		fmt.Printf(" - %-22s %-15s %-50s \n", green(val.Time.String()), yellow(strconv.Itoa(val.Bytes)), val.Url)
	}

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
