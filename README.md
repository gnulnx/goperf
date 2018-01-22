# goperf

## Go based Load Tester

Usage:
### Build the package
go build

### Basic perf test.
./goperf -url {url} -sec=3 -connections=3


### Fetch and return all stats as json
./goperf -url {url} fetch

### Fetch and return all assets bodys (js, css, html) as json
./goperf -url {url} fetchall
