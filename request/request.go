package request

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gnulnx/goperf/httputils"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
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
    NOTE: Input is intentionally passed by value.
*/
func Fetch(url string, retdat bool) *FetchResponse {
	/*
		Simple method that fetches a url and returns a FetchOutput structure
		retdata if True then we return the Body and the Headers
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

	output := FetchResponse{
		Url:     url,
		Body:    responseBody,
		Headers: resp.Header,
		Bytes:   len(responseBody),
		Runes:   utf8.RuneCountInString(responseBody),
		Time:    end.Sub(start),
		Status:  resp.StatusCode,
	}

	if !retdat {
		output.Body = ``
		output.Headers = make(map[string][]string)
	}
	//Close the response body and return the output
	resp.Body.Close()
	return &output
}

type FetchResponse struct {
	Url     string
	Body    string
	Headers map[string][]string
	Bytes   int
	Runes   int
	Time    time.Duration
	Status  int
}

func FetchAll(baseurl string, retdat bool) *FetchAllResponse {
	/*
		Try to simulate a request.
		1) Fetch base_url
		2) Parse for script, css, and img tags
		3) Fetch other resources with go threads
		4) compile results and return as json
	*/
	// Fetch initial url
	output := Fetch(baseurl, true)
	color.Red("Fetching: " + output.Url)
	if output.Status == 200 {
		color.Green(" - Status: " + strconv.Itoa(output.Status))
	} else {
		color.Red(" - Status: " + strconv.Itoa(output.Status))
	}
	color.Yellow(" - Time to first byte: " + output.Time.String())
	color.Yellow(" - Bytes: " + strconv.Itoa(output.Bytes))
	color.Yellow(" - Runes: " + strconv.Itoa(output.Runes))

	// Now parse for js, css, img urls
	jsfiles, imgfiles, cssfiles, bundle := httputils.Resources(output.Body)

	// TODO:  This needs to be done as Go routines to simulate a real browser
	jsResponses := []FetchResponse{}
	files := *jsfiles
	for i := 0; i < len(files); i++ {
		asset_url := (files)[i]
		jsResponses = append(jsResponses, *FetchAsset(baseurl, asset_url, retdat))
		color.Magenta(asset_url)
	}

	imgResponses := []FetchResponse{}
	files = *imgfiles
	for i := 0; i < len(files); i++ {
		asset_url := (files)[i]
		imgResponses = append(imgResponses, *FetchAsset(baseurl, asset_url, retdat))
		color.Magenta(asset_url)
	}

	cssResponses := []FetchResponse{}
	files = *cssfiles
	for i := 0; i < len(files); i++ {
		asset_url := (files)[i]
		cssResponses = append(cssResponses, *FetchAsset(baseurl, asset_url, retdat))
		color.Magenta(asset_url)
	}
	if !retdat {
		output.Body = ``
		output.Headers = make(map[string][]string)
	}
	outputall := FetchAllResponse{
		//Url:          baseurl,
		BaseUrl:      output,
		JSReponses:   jsResponses,
		IMGResponses: imgResponses,
		CSSResponses: cssResponses,
	}
	//tmp, _ := json.MarshalIndent(outputall, "", "    ")
	//fmt.Println(string(tmp))
	log("Javascript files", jsfiles)
	log("CSS files", cssfiles)
	log("IMG files", imgfiles)
	log("Full Bundle", bundle)

	return &outputall
}

type FetchAllResponse struct {
	//Url          string
	BaseUrl      *FetchResponse
	JSReponses   []FetchResponse
	IMGResponses []FetchResponse
	CSSResponses []FetchResponse

	Body   string
	Time   time.Duration
	Status int
}

func DefineAssetUrl(baseurl string, asseturl string) string {
	if asseturl[0] == '/' {
		asseturl = baseurl + asseturl
	}
	return asseturl
}

func FetchAsset(baseurl string, asseturl string, retdat bool) *FetchResponse {
	asset_url := DefineAssetUrl(baseurl, asseturl)
	return Fetch(asset_url, retdat)
}

func log(header string, files *[]string) {
	color.Red(header)
	for i := 0; i < len(*files); i++ {
		color.Cyan(" - " + (*files)[i])
	}
}

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
