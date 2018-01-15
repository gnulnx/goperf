package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

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
type FetchOutput struct {
	Url    string
	Body   string
	Time   time.Duration
	Status int
}

func Fetch(url string) *FetchOutput {
	/*
		Simple method that fetches a url and returns a FetchOutput structure
	*/

	// Set up the http request
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", "Chrome/61.0.3163.100 Mobile Safari/537.36")

	//Start Timer
	start := time.Now()

	//Fetch the url
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
	//End Timer
	end := time.Now()

	// Read the html 'body' content from the response object
	body, err := ioutil.ReadAll(resp.Body)
	responseBody := string(body)

	output := FetchOutput{
		Url:    url,
		Body:   responseBody,
		Time:   end.Sub(start),
		Status: resp.StatusCode,
	}
	//Close the response body and return the output
	resp.Body.Close()
	return &output
}

type FetchAllOutput struct {
	Url    string
	JS     *[]string
	IMG    *[]string
	CSS    *[]string
	Body   string
	Time   time.Duration
	Status int
}

func FetchAll(url string) *FetchAllOutput {
	// Fetch initial url
	output := Fetch(url)

	//Now parse it for javascript
	// Get a list of all script urls to download
	jsfiles := getjs(output.Body)
	imgfiles := getimg(output.Body)
	cssfiles := getcss(output.Body)

	fmt.Println(jsfiles)
	fmt.Println(imgfiles)
	fmt.Println(cssfiles)

	outputall := FetchAllOutput{
		Url: url,
		JS:  jsfiles,
		IMG: imgfiles,
		CSS: cssfiles,
	}
	return &outputall
}

func getjs(body string) *[]string {
	r, _ := regexp.Compile(`<script.*?src="(.*?)"`)
	match := r.FindAllStringSubmatch(body, -10)
	jsfiles := make([]string, 0)
	for j := 0; j < len(match); j++ {
		jsfiles = append(jsfiles, match[j][1])
	}
	return &jsfiles
}

func getimg(body string) *[]string {
	r, _ := regexp.Compile(`<img.*?src="(.*?)"`)
	match := r.FindAllStringSubmatch(body, -10)
	imgfiles := make([]string, 0)
	for j := 0; j < len(match); j++ {
		imgfiles = append(imgfiles, match[j][1])
	}
	return &imgfiles
}

func getcss(body string) *[]string {
	r, _ := regexp.Compile(`<link.*?href="(.*?)"`)
	match := r.FindAllStringSubmatch(body, -10)
	cssfiles := make([]string, 0)
	for j := 0; j < len(match); j++ {
		cssfiles = append(cssfiles, match[j][1])
	}
	return &cssfiles
}

func (input Input) Run(done chan Result) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", input.Url, nil)
	req.Header.Add("user-agent", "Chrome/61.0.3163.100 Mobile Safari/537.36")

	// Causes segfault if i try to assign baes_url to input.Url... not sure why
	//base_url := string(input.Url)
	//fmt.Println(&base_url, &input.Url)
	//base_url := *input.Url
	base_url := string(`https://teaquinox.com/`)

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
