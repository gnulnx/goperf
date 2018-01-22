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
	"github.com/gnulnx/color"
	"github.com/gnulnx/goperf/perf"
	"github.com/gnulnx/goperf/request"
	"net/http"
	"os"
	"runtime/pprof"
	"strconv"
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
	web := flag.Bool("web", false, "Run as a webserver")

	// Not currently used, but could be
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	output := flag.Int("output", 5, "Show thread output every {n} iterations")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	flag.Parse()

	if *web {
		http.HandleFunc("/api/", handler)
		http.ListenAndServe(":8080", nil)
	}

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
	perfJob := &perf.Init{
		Iterations: *iterations,
		Threads:    *threads,
		Url:        *url,
		Output:     *output,
		Verbose:    *verbose,
		Seconds:    *seconds,
	}
	f, _ := os.Create(*cpuprofile)
	pprof.StartCPUProfile(f)
	results := perfJob.Basic()
	defer pprof.StopCPUProfile()

	// Write json response to file.
	outfile, _ := os.Create("./results.json")

	if *printjson {
		perfJob.Json()
	} else {
		perfJob.Print()
	}

	tmp, _ := json.MarshalIndent(results, "", "    ")
	outfile.WriteString(string(tmp))
	color.Magenta("Job Results Saved: ./results.json")
}

func handler(w http.ResponseWriter, r *http.Request) {
	/*
	 */
	r.ParseForm()
	url, ok := r.PostForm["url"]
	if !ok {
		w.Write([]byte("url is required"))
		return
	}

	strSeconds, ok := r.PostForm["seconds"]
	if !ok {
		w.Write([]byte("seconds is required"))
		return
	}
	s := strSeconds[0]
	seconds, _ := strconv.Atoi(s)

	strConnections, ok := r.PostForm["conn"]
	if !ok {
		w.Write([]byte("conn is required"))
		return
	}
	//c := strConnections[0]
	conn, _ := strconv.Atoi(strConnections[0])

	if 1 == 0 {
		strconv.Itoa(2)
	}

	fmt.Fprintf(w, "PostForm: %s\n", r.PostForm)

	perfJob := &perf.Init{
		//Iterations: 10,
		Threads: conn,
		Url:     url[0],
		Seconds: seconds,
		//Output:     *output,
		//Verbose:    *verbose,
	}
	results := perfJob.Basic()
	tmp, _ := json.Marshal(results)
	//tmp, _ := json.MarshalIndent(results, "", "   ")
	_ = string(tmp)
	w.Header().Set("Server", "A Go Web Server")
	w.Header().Set("Content-Type", "application/json")
	w.Write(tmp)
}
