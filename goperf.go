package main

/*
goperf is a load testing tool.

** Use Case 1:  Fetch a url and report stats

This command will return all information for a given url.
./goperf -url http://qa.teaquinox.com -fetchall -printjson

When fetchall is provided the returned struct will contain
url, time, size, and data info.

You can do a simpler request that leaves the data and headers out like this
./goperf -url http://qa.teaquinox.com -fetchall -printjson


** Use Case 2: Load testing


*/

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/gnulnx/goperf/request"
	"os"
	"runtime/pprof"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	// I ❤️  the way go handles command line arguments
	fetch := flag.Bool("fetch", false,
		"Fetch -url and report it's stats. Does not return resources")

	fetchall := flag.Bool("fetchall", false,
		") Fetch -url and report stats  return all assets (js, css, img)")

	printjson := flag.Bool("printjson", false, "Print json output")

	threads := flag.Int("connections", 1, "Number of concurrent connections")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	seconds := flag.Int("sec", 2, "Number of seconds each concurrant thread/user should make requests")
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	output := flag.Int("output", 5, "Show thread output every {n} iterations")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	increment := flag.Int("incr", 2, "How fast to increment the number of concurrent connections")
	max_response := flag.Int("max-response", 1, "Maximun number of seconds to wait for a response")
	flag.Parse()

	if *fetch || *fetchall {
		f, _ := os.Create(*cpuprofile)
		pprof.StartCPUProfile(f)

		color.Green("~~ Fetching a single url and printing info ~~")
		resp := request.FetchAll(*url, *fetchall)

		if *printjson {
			tmp, _ := json.MarshalIndent(resp, "", "    ")
			fmt.Println(string(tmp))
		}

		request.PrintFetchAllResponse(resp)

		pprof.StopCPUProfile()
		os.Exit(1)
	}

	// TODO Declare an inline parameter struct...
	input := request.Input{
		Iterations: *iterations,
		Threads:    *threads,
		Url:        *url,
		Output:     *output,
		Verbose:    *verbose,
		Seconds:    *seconds,
	}
	perf2(input)
	os.Exit(1)

	// Working on New Method:  Currently fetch url and assets.

	// Old Method Below Here.  Likely not relevant any longer
	fmt.Println("Running url again:", *url)
	fmt.Println("Concurrent Connections: ", *threads, "Sustained for: ", input.Threads)

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

func iterateRequest(url string, sec int) (time.Duration, int, time.Duration) {
	start := time.Now()
	maxTime := time.Duration(sec * 1000 * 1000 * 1000)
	elapsedTime := maxTime
	count := 1
	for {
		request.FetchAll(url, false)
		elapsedTime = time.Now().Sub(start)
		//fmt.Println(" - ", elapsedTime, " - ", resp.TotalTime)
		if elapsedTime > maxTime {
			break
		}
		count += 1
	}
	fmt.Println("----------------------------")
	fmt.Println(" - total: ", elapsedTime)
	fmt.Println(" - Num of Requests: ", int64(count))
	avg := time.Duration(int64(elapsedTime) / int64(count))
	fmt.Println(" - avg: ", avg)
	fmt.Println("----------------------------")

	return elapsedTime, count, avg
}

func perf2(input request.Input) time.Duration {
	// Print input params as JSON
	tmp, _ := json.MarshalIndent(input, "", "    ")
	fmt.Println(string(tmp))

	// Create some channels
	chanslice := []chan bool{}
	for i := 0; i< input.Threads; i++ {
		chanslice = append(chanslice, make(chan bool));
	}

	for i := 0; i < len(chanslice); i++ {
		go func(c chan bool) { 
			iterateRequest(input.Url, input.Seconds)
			//elapsed, numRequests, avg := iterateRequest(input.Url, input.Seconds)
			//fmt.Println(elapsed, numRequests, avg)
			c <- true
		}(chanslice[i])
	}


	// Wait on all the threads
	for i := 0; i < len(chanslice); i++ {
		<-chanslice[i]
	}


	return 0
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
