package request

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
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

func log(header string, files []string) {
	color.Red(header)
	for i := 0; i < len(files); i++ {
		color.Cyan(" - " + (files)[i])
	}
}
