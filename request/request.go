package request

import (
	"fmt"
	"time"
	"net/http"
)

/* Results for a perf thread */
type Result struct {
    Total   time.Duration
    Average time.Duration
    Channel int
}

//display method for Results
func (r *Result) Display() {
    fmt.Println("Channel(", r.Channel, ") Total(", r.Total, ") Average(", r.Average, ")")
}

// This is the input structure for a perf thread
type Input struct {
    Url        string
    Threads    int
    Iterations int
    Output     int
    Index      int // Also the channel number
}

// This is the main url perf testing method
func (input Input) Run(done chan Result) {
    client := &http.Client{}

    req, _ := http.NewRequest("GET", input.Url, nil)
    req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Mobile Safari/537.36")
    start := time.Now()
    for i := 0; i < input.Iterations; i++ {
        client.Do(req)
        if i % input.Output == 0 {
            fmt.Println("Thread: ", input.Index, " iteration: ", i)
        }
    }
    end := time.Now()
    total := end.Sub(start)
    average := total / time.Duration(input.Iterations)

    // Send results on done channel
    done <- Result{
        Total:   total,
        Average: average,
        Channel: input.Index,
    }
}
