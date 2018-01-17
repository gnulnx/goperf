package request

import (
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// display method for Results
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

		if strings.Contains(status_code, "200") {
			status_200 = append(status_200, r.Responses[resp].Status)
		} else if strings.Contains(status_code, "400") {
			status_400 = append(status_400, r.Responses[resp].Status)
		} else if strings.Contains(status_code, "500") {
			status_500 = append(status_500, r.Responses[resp].Status)
		} else {
			fmt.Println("Invalid status code", status_code)
		}

	}
	fmt.Println("200x: ", float32(len(status_200))/float32(len(r.Responses))*100.0, "%")
	fmt.Println("400x: ", float32(len(status_400))/float32(len(r.Responses))*100.0, "%")
	fmt.Println("500x: ", float32(len(status_500))/float32(len(r.Responses))*100.0, "%")
}

func log(header string, files *[]string) {
	color.Red(header)
	for i := 0; i < len(*files); i++ {
		color.Cyan(" - " + (*files)[i])
	}
}

/*
   Run the input parameters defined in Input struct.
   Channel 'done' expects a Result object
    NOTE: Input is intentionally passed by value.
*/
func (input Input) Run(done chan Result) {
	base_url := input.Url

	client := &http.Client{}
	req, _ := http.NewRequest("GET", base_url, nil)
	req.Header.Add("user-agent", "Chrome/61.0.3163.100 Mobile Safari/537.36")

	start := time.Now()
	responses := make([]*http.Response, 0, input.Iterations)

	for i := 0; i < input.Iterations; i++ {
		// Requet the base page
		resp, err := client.Do(req)
		responses = append(responses, resp)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(0)
		}

		// This is how you would read the body
		body, err := ioutil.ReadAll(resp.Body)
		responseBody := string(body)

		done := make(chan bool) // Channel to keep track of extra assets downloads
		assets := 0             // Counter for all the static and img resources

		// Get a list of all script urls to download
		r, _ := regexp.Compile(`<script.*?src="(.*?)"`)
		match := r.FindAllStringSubmatch(responseBody, -10)
		for j := 0; j < len(match); j++ {
			assets += 1
			url := base_url + match[j][1]
			go func(url string) {
				req, _ := http.NewRequest("GET", url, nil)
				req.Header.Add("user-agent", "Chrome/61.0.3163.100 Mobile Safari/537.36")
				resp, err = client.Do(req)
				//status := strconv.Itoa(resp.StatusCode)
				//fmt.Println(status + "  " + url)
				done <- true
			}(url)
		}

		//Get a list of all image urls to download
		r, _ = regexp.Compile(`<img.*?src="(.*?)"`)
		match = r.FindAllStringSubmatch(responseBody, -10)
		for j := 0; j < len(match); j++ {
			assets += 1
			url := base_url + match[j][1]
			go func(url string) {
				req, _ := http.NewRequest("GET", url, nil)
				req.Header.Add("user-agent", "Chrome/61.0.3163.100 Mobile Safari/537.36")
				client.Do(req)
				//status := strconv.Itoa(resp.StatusCode)
				//fmt.Println(status + "  " + url)
				done <- true
			}(url)
		}

		// Now lets wait on all the extra requests to get called
		for wait := 0; wait < assets; wait++ {
			<-done
		}
		//fmt.Println(img_urls)

		// TODO You need to loop over the script and img tags and download them as part of the perf test
		// TODO you also need to include css tags in here

		// Now close the response body
		resp.Body.Close()

		//if i != 0 && i%input.Output == 0 {
		//	fmt.Println("Thread: ", input.Index, " iteration: ", i)
		//}
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
