# goperf - Performance tester
A highly concurrant website load tester with simple intuitive command line syntax.

### Running on a 32 CPU machine
![Alt text](readme_imgs/GoPerf.png?raw=true "GoPerf")

Goperf fetches the html document as well as all the img, css, and js assets in an effort to realistically simulate an initial browser request to your site.

Goperf also supports simple request settings like user-agent and cookies strings.


## Usage:

### Fetch

Fetch a page and its assets and display info.  
```
./goperf -url {url} -fetch
```
This will print output like:

![Alt text](readme_imgs/Fetch.png?raw=true "Fetch")

To Fetch a page and display all it's assets use:
```
./goperf -url {url} -fetch --printjson
```
NOTE this will print the content of the body in each of the fetched assets. If you have large minified JS bundles it will be pretty messy.


Fetch a page and it's assets (js, css, img) and return the bodies for the assets.
--printjson also pretty prints the json that is returned (above other output).
```
./goperf -url {url} -fetchall --printjson
```

Fetch a page that requires a session id (such as a django login)

```
./goperf -url http://192.168.33.11/student/ -fetchall -cookies "sessionid_vagrant=0xkfeev882n0i9efkiq7vmd2i6efufz9;" --printjson
```

## Load testing

Tell goperf the number of users you want simulate and the number of seconds you want the simulation to run.

```
./goperf -url {url} -users {int}  -sec {int}
```

Goperf will give off a seperate go routine for each user.  Each user will then continiously fetch the url along with all it's page assets (also in seperate go routines).  The light weight nature of goroutines allows this high concurancy to simulate many users with very litte memory.  You will most likely overhewlm the test url servers or consume all of the available network bandwidth before memory becomes an issue.

Load testing results: 

![Alt text](readme_imgs/GoPerfOutput.png?raw=true "Output")

## Setup
#### Ensure gopath is correctly setup

Make sure you have your GOPATH setup to point to the go/bin directory.
If you have a default go install on ubuntu it would be ~/go/bin.
If so you would add this to your path.
```
export PATH=$PATH:~/go/bin
```
#### Install

```
go get github.com/gnulnx/goperf
```

#### Build
```
go install github.com/gnulnx/goperf
```


### Basic perf test.

Fire a 3 second test with 3 simultaneous connections
```
./goperf -url {url} -sec=3 -connections=3
```

### Run minimal unit and benchmark tests
```
go test ./... -cover -bench
```


## Road map and future plans.

Currently goperf is currently quite good at simulating browser requests that include the body, css, img, and js assets.  

However goper has no concept of an ajax request.  

The next phase of golang will be adding in support for additional requests after intial page load.  For example say you wanted to time how long it took for 10 users to hit your website and also request a specific api.  This approach will allow us to have much better simulation for javacsript heavy sites.  

Longer term support for a chaos mode where the performance "users" move through the site randomly selecting a new url after each request. 
