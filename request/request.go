package request

import (
	"fmt"
	"net/http"
	"os"
	"time"
    "io/ioutil"
    "regexp"
)

import s "strings"

/*
	Structure used to create web request Channel.  This is how we get the results
	back from the 'go Run(...) method call
*/
type Result struct {
	Total     time.Duration
	Average   time.Duration
	Channel   int
	Responses []*http.Response
}

//display method for Results
func (r *Result) Display() {
	fmt.Println("---------------------------------------------------------------------")
	fmt.Println("Channel(", r.Channel, ") Total(", r.Total, ") Average(", r.Average, ")")
	r.Status()
	fmt.Println("---------------------------------------------------------------------")
}

func (r *Result) Status() {
	status_200 := make([]string, 0, len(r.Responses))
	status_400 := make([]string, 0, len(r.Responses))
	status_500 := make([]string, 0, len(r.Responses))

	for resp := 0; resp < len(r.Responses); resp++ {
		status_code := r.Responses[resp].Status

		if s.Contains(status_code, "200") {
			status_200 = append(status_200, r.Responses[resp].Status)
		} else if s.Contains(status_code, "400") {
			status_400 = append(status_400, r.Responses[resp].Status)
		} else if s.Contains(status_code, "500") {
			status_500 = append(status_500, r.Responses[resp].Status)
		} else {
			fmt.Println("Invalid status code", status_code)
		}

	}
	fmt.Println("200x: ", float32(len(status_200))/float32(len(r.Responses))*100.0, "%")
	fmt.Println("400x: ", float32(len(status_400))/float32(len(r.Responses))*100.0, "%")
	fmt.Println("500x: ", float32(len(status_500))/float32(len(r.Responses))*100.0, "%")
}

/*
   Structure to hold the input variables to 'go Run(...)' method call
*/
type Input struct {
	Url        string
	Threads    int
	Iterations int
	Output     int
	Index      int // Also the channel number
	Verbose    bool
}

/*
   Run the input parameters defined in Input struct.
   Channel 'done' expects a Result object
    NOTE: Input is intentioanlly passed by value.
*/
func (input Input) Run(done chan Result) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", input.Url, nil)
	req.Header.Add("user-agent", "Chrome/61.0.3163.100 Mobile Safari/537.36")

	start := time.Now()
	responses := make([]*http.Response, 0, input.Iterations)

	for i := 0; i < input.Iterations; i++ {
		resp, err := client.Do(req)
		responses = append(responses, resp)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(0)
		}

		// This is how you would read the body
		body, err := ioutil.ReadAll(resp.Body)
        responseBody := string(body)

        // Get a list of all script urls to download
        r, _ := regexp.Compile(`<script.*?src="(.*?)"`)
        match := r.FindAllStringSubmatch(responseBody, -10)
        script_urls := make([]string, 0)
        for j := 0; j < len(match); j++ {
            script_urls = append(script_urls, match[j][1])
            //fmt.Println(match[j][1])
        }
        fmt.Println(script_urls)

        //Get a list of all image urls to download
        r, _ = regexp.Compile(`<img.*?src="(.*?)"`)
        match = r.FindAllStringSubmatch(responseBody, -10)
        img_urls := make([]string, 0)
        for j := 0; j < len(match); j++ {
            img_urls = append(img_urls, match[j][1])
            fmt.Println(match[j][1])
        }
        fmt.Println(img_urls)

        // TODO You need to loop over the script and img tags and download them as part of the perf test
        // TODO you also need to include css tags in here

        // Now close the response body
		resp.Body.Close()

		if i != 0 && i%input.Output == 0 {
			fmt.Println("Thread: ", input.Index, " iteration: ", i)
		}
	}
	end := time.Now()
	total := end.Sub(start)
	average := total / time.Duration(input.Iterations)

	// Send results on done channel
	done <- Result{
		Total:     total,
		Average:   average,
		Channel:   input.Index,
		Responses: responses,
	}
}
