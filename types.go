package crawler

import (
	"net/http"
	"regexp"
	"os"
)

type Crawler struct {
    urlPattern *regexp.Regexp
    outputFileFound *os.File
    outputFileCrawled *os.File
    foundHosts map[string]int32
    initialHosts []string
    requestPool chan string
    parsePool chan[]string
    outputChannel chan writeRequest
    counterChannel chan counterMsg
    countStore map[logType]int
    queues map[outputType][][]string
    httpClient http.Client
    state bool
}

type counterMsg struct {
    Key logType
    Value int
}

type Host struct {
    backlinks int
    timeFound int32
}

type writeRequest struct {
    file outputType
    host string
}

type outputType string

type logType string

