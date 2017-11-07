package request

import (
	"fmt"
	"net/http"
	"os"
	"time"
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
	fmt.Println("Channel(", r.Channel, ") Total(", r.Total, ") Average(", r.Average, ")")
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
	fmt.Println("200x: ", status_200)
	fmt.Println("400x: ", status_400)
	fmt.Println("500x: ", status_500)
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
}

/*
   Run the input parameters defined in Input struct.
   channel 'done' expects a Result object
*/
func (input Input) Run(done chan Result) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", input.Url, nil)
	req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Mobile Safari/537.36")

	start := time.Now()
	responses := make([]*http.Response, 0, input.Iterations)

	for i := 0; i < input.Iterations; i++ {
		resp, err := client.Do(req)
		responses = append(responses, resp)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(0)
		}
		resp.Body.Close()

		// This is how you would read the body
		//body, err := ioutil.ReadAll(resp.Body)
		//fmt.Println(string(body))

		if i%input.Output == 0 {
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
