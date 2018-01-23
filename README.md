# goperf

## Go based Load Tester
This project is still in rabid development mode.  It is definitly not production ready code.
However, it does work and you may find it useful as is.  ;)

### Build the package
go build

## Usage:
### Basic perf test.

Fire a 3 second test with 3 simultaneous connections
```
./goperf -url {url} -sec=3 -connections=3
```

### Fetch

Fetch a page and its assets and display info.  
--printjson also pretty prints the json that is displayed.
```
./goperf -url {url} -fetch --printjson
```

Fetch a page and it's assets (js, css, img) and return the bodies for the assets.
--printjson also pretty prints the json that is displayed.
```
./goperf -url {url} -fetchall --printjson
```

### Run minimal unit and benchmarck tests
```
go test ./... -cover -bench
```
