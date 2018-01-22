# goperf

## Go based Load Tester

### Build the package
go build

## Usage:
### Basic perf test.
./goperf -url {url} -sec=3 -connections=3


### Fetch and return all stats as json
./goperf -url {url} -fetch --printjson

### Fetch and return all assets bodys (js, css, html) as json
./goperf -url {url} -fetchall --printjson
