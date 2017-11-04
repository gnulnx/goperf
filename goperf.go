package main

import (
    "fmt"
    "net/http"
)

func main() {
    url := "https://qa.teaquinox.com"
    //url := "https://www.google.com"

    client := &http.Client{
        //CheckRedirect: redirectPolicyFunc,
    }

    //Create the basic request
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Mobile Safari/537.36")
    
    for i:=0; i<10; i++ {
        resp, err := client.Do(req)
        fmt.Println("resp: ", resp)
        fmt.Println("err: ", err)
    }
}


func get_url(count int) (time float64) {
    
    return (1.234)
}
