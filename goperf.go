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
	"github.com/gnulnx/vestigo"
	"io"
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
	users := flag.Int("users", 1, "Number of concurrent users/connections")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	seconds := flag.Int("sec", 2, "Number of seconds each concurrant user/connection should make consequitive requests")
	web := flag.Bool("web", false, "Run as a webserver -web {port}")
	port := flag.Int("port", 8080, "used with -web to specif which port to bind")
	cookies := flag.String("cookies", "{}", "Set up cookies for the request")
	useragent := flag.String("useragent", "goperf", "Set the user agent string")

	// Not currently used, but could be
	iterations := flag.Int("iter", 1000, "Iterations per user/connection")
	output := flag.Int("output", 5, "Show user output every {n} iterations")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	flag.Parse()

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100

	if *web {
		router := vestigo.NewRouter()
		router.SetGlobalCors(&vestigo.CorsAccessControl{
			AllowOrigin: []string{"*", "http://138.197.97.39:8080"},
		})

		router.Post("/api/", handler)
		router.SetCors("/api/", &vestigo.CorsAccessControl{
			AllowMethods: []string{"POST"}, // only allow cors for this resource on POST calls
		})
		sPort := ":" + strconv.Itoa(*port)
		color.Green("Your website is available at 127.0.0.1:%s", sPort)
		http.ListenAndServe(sPort, router)
	}

	if *fetch || *fetchall {
		color.Green("~~ Fetching a single url and printing info ~~")

		// This section will make an initial GET request and try to set any cookies we find
		if *cookies == "" {
			resp1, _ := http.Get(*url)
			if len(resp1.Header["Set-Cookie"]) > 0 {
				cookies = &resp1.Header["Set-Cookie"][0]
				fmt.Println("cookies: ", *cookies)
			}
		}
		fmt.Println("cookies: ", *cookies)

		resp := request.FetchAll(
			request.FetchInput{
				BaseUrl:   *url,
				Retdat:    *fetchall,
				Cookies:   *cookies,
				UserAgent: *useragent,
			},
		)

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
		Threads:    *users,
		Url:        *url,
		Output:     *output,
		Verbose:    *verbose,
		Seconds:    *seconds,
		Cookies:    *cookies,
		UserAgent:  *useragent,
	}
	f, _ := os.Create(*cpuprofile)
	pprof.StartCPUProfile(f)
	results := perfJob.Basic()
	defer pprof.StopCPUProfile()

	// Write json response to file.
	outfile, _ := os.Create("./output.json")

	if *printjson {
		perfJob.JsonResults()
	} else {
		perfJob.Print()
	}

	tmp, _ := json.MarshalIndent(results, "", "    ")
	outfile.WriteString(string(tmp))
	color.Magenta("Job Results Saved: ./output.json")
}

func handler(w http.ResponseWriter, r *http.Request) {
	/*
	 */
	r.ParseForm()
	fmt.Println(r.PostForm)
	fmt.Println(r.Form)
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
	seconds, _ := strconv.Atoi(strSeconds[0])

	strConnections, ok := r.PostForm["conn"]
	if !ok {
		w.Write([]byte("conn is required"))
		return
	}
	conn, _ := strconv.Atoi(strConnections[0])

	perfJob := &perf.Init{
		Threads: conn,
		Url:     url[0],
		Seconds: seconds,
	}
	perfJob.Basic()
	json_results := perfJob.JsonResults()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, json_results)
}
