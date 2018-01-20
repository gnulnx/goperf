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
	//increment := flag.Int("incr", 2, "How fast to increment the number of concurrent connections")
	//max_response := flag.Int("max-response", 1, "Maximun number of seconds to wait for a response")
	flag.Parse()

	if *fetch || *fetchall {

		color.Green("~~ Fetching a single url and printing info ~~")
		resp := request.FetchAll(*url, *fetchall)

		if *printjson {
			tmp, _ := json.MarshalIndent(resp, "", "    ")
			fmt.Println(string(tmp))
		}

		request.PrintFetchAllResponse(resp)

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
	f, _ := os.Create(*cpuprofile)
	pprof.StartCPUProfile(f)
	perf2(input)

	defer pprof.StopCPUProfile()
}

//func iterateRequest(url string, sec int) (time.Duration, int, time.Duration, []request.FetchAllResponse) {
func iterateRequest(url string, sec int) []request.FetchAllResponse {
	start := time.Now()
	maxTime := time.Duration(sec * 1000 * 1000 * 1000)
	elapsedTime := maxTime
	count := 1
	resps := []request.FetchAllResponse{}
	for {
		resps = append(resps, *request.FetchAll(url, false))
		elapsedTime = time.Now().Sub(start)
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

	return resps
	//return elapsedTime, count, avg, resps
}

func perf2(input request.Input) time.Duration {
	// Print input params as JSON
	tmp, _ := json.MarshalIndent(input, "", "    ")
	fmt.Println(string(tmp))

	// Create some channels
	chanslice := []chan []request.FetchAllResponse{}
	for i := 0; i< input.Threads; i++ {
		chanslice = append(chanslice, make(chan []request.FetchAllResponse));
	}

	// Fire off a new go routine for each channel
	for i := 0; i < len(chanslice); i++ {
		go func(c chan []request.FetchAllResponse) { 
			//elapsed, numRequests, avg := iterateRequest(input.Url, input.Seconds)
			//fmt.Println(elapsed, numRequests, avg)
			//iterateRequest(input.Url, input.Seconds)
			c <- iterateRequest(input.Url, input.Seconds)
		}(chanslice[i])
	}

	// Wait on all the channels
	results := [][]request.FetchAllResponse{}
	for i := 0; i < len(chanslice); i++ {
		results = append(results, <-chanslice[i])
	}

	tmp, _ = json.MarshalIndent(results, "", "    ")
    //fmt.Println(string(tmp))

	f, _ := os.Create("./results.json")
	f.WriteString(string(tmp))

	return 0
}
