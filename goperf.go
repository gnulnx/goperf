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
	"github.com/gnulnx/goperf/perf"
	"github.com/gnulnx/goperf/request"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	// I ❤️  the way go handles command line arguments
	fetch := flag.Bool("fetch", false, "Fetch -url and report it's stats. Does not return resources")
	fetchall := flag.Bool("fetchall", false, "Fetch -url and report stats  return all assets (js, css, img)")
	printjson := flag.Bool("printjson", false, "Print json output")
	threads := flag.Int("connections", 1, "Number of concurrent connections")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	seconds := flag.Int("sec", 2, "Number of seconds each concurrant thread/user should make requests")
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	output := flag.Int("output", 5, "Show thread output every {n} iterations")
	verbose := flag.Bool("verbose", false, "Show verbose output")
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
	input := perf.Input{
		Iterations: *iterations,
		Threads:    *threads,
		Url:        *url,
		Output:     *output,
		Verbose:    *verbose,
		Seconds:    *seconds,
	}
	f, _ := os.Create(*cpuprofile)
	pprof.StartCPUProfile(f)
	results := perf.Perf(input)
	defer pprof.StopCPUProfile()

	// Write json response to file.
	outfile, _ := os.Create("./results.json")
	tmp, _ := json.MarshalIndent(results, "", "    ")
	//tmp, _ = json.Marshal(results)
	outfile.WriteString(string(tmp))

	color.Magenta("Job Results: results.json")

}
