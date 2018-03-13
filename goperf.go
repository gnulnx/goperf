/*
Package goper is a highly concurrant website load tester with a simple intuitive command line syntax.

* Fetch a url and report stats

This command will return all information for a given url.
 ./goperf -url http://qa.teaquinox.com -fetchall -printjson

When fetchall is provided the returned struct will contain
url, time, size, and data info.

You can do a simpler request that leaves the data and headers out like this
 ./goperf -url http://qa.teaquinox.com -fetchall -printjson


* Load testing
 ./goperf -url http://qa.teaquinox.com -sec 5 -users 5
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/pprof"
	"strconv"

	"github.com/gnulnx/color"
	"github.com/gnulnx/goperf/perf"
	"github.com/gnulnx/goperf/request"
	"github.com/gnulnx/vestigo"
)

func main() {
	// I ❤️  the way go handles command line arguments
	fetch := flag.Bool("fetch", false, "Fetch -url and report it's stats. Does not return resources")
	fetchall := flag.Bool("fetchall", false, "Fetch -url and report stats  return all assets (js, css, img)")
	printjson := flag.Bool("printjson", false, "Print json output")
	perftest := flag.Bool("perftest", false, "Run the goland perf suite")
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
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
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
		color.Green("Your website is available at 127.0.0.1%s", sPort)
		http.ListenAndServe(sPort, router)
	}

	if *fetch || *fetchall {
		// TODO This method treats these command line arguments exactly the same... no good
		// -fetch -printjson should ONLY return the body of the primary request and not the other assets

		// This section will make an initial GET request and try to set any cookies we find
		if *cookies == "" {
			resp1, _ := http.Get(*url)
			if len(resp1.Header["Set-Cookie"]) > 0 {
				cookies = &resp1.Header["Set-Cookie"][0]
			}
		}
		resp := request.FetchAll(
			request.FetchInput{
				BaseURL:   *url,
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
	results := perfJob.Basic()

	if *perftest {
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

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

/*
Check that the request parameters are correct and return them.
Also return an array of error string if the parameters were not right
*/
func checkParams(r *http.Request) ([]string, string, int, int) {
	errors := []string{}
	seconds := 0
	users := 0
	var err error

	// Check that url has been supplied
	url, ok := r.PostForm["url"]
	if !ok {
		errors = append(errors, " - url (string) is a required field")
		url = []string{""}
	}

	// Check that seconds is supplied
	strSeconds, ok := r.PostForm["sec"]
	if !ok {
		errors = append(errors, " - sec (int) is a required field")
		strSeconds = []string{}
	}
	if len(strSeconds) > 0 {
		seconds, err = strconv.Atoi(strSeconds[0])
		if err != nil {
			errors = append(errors, " - sec (int) is a required field")
			seconds = 0
		}
	}

	// Check user field has been supplied
	strUsers, ok := r.PostForm["users"]
	if !ok {
		errors = append(errors, " - users (int) is a required field")
		strUsers = []string{}
	}
	if len(strUsers) > 0 {
		users, err = strconv.Atoi(strUsers[0])
		if err != nil {
			errors = append(errors, " - users (int) is a required field")
			users = 0
		}
	}

	return errors, url[0], seconds, users
}

func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	errors, url, seconds, users := checkParams(r)
	if len(errors) > 0 {
		for i := 0; i < len(errors); i++ {
			e := errors[i] + "\n"
			w.Write([]byte(e))
		}
		return
	}

	perfJob := &perf.Init{
		Url:     url,
		Threads: users,
		Seconds: seconds,
	}
	perfJob.Basic()
	jsonResults := perfJob.JsonResults()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, jsonResults)
}
