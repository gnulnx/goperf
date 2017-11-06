package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	url := "https://qa.teaquinox.com"
	//url := "https://teaquinox.com"
	//Create the basic request


	for i := 0; i < 100; i++ {
		go run(url, i)
	}
	run(url, 11)
}

func run(url string, index int) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Mobile Safari/537.36")
	//url := "https://www.google.com"


	iterations := 1000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		client.Do(req)
		fmt.Println(index)
	}
	end := time.Now()
	total := end.Sub(start)
	fmt.Println("Total: ", total)
	fmt.Println("Avg: ", total/time.Duration(iterations))
}
