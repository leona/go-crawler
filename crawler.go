package crawler

import (
	"time"
	"fmt"
	"net/http"
	"io/ioutil"
	"regexp"
	"os"
	"net/url"
	"strconv"
	"sort"
    "crypto/tls"
)

var urlPattern = regexp.MustCompile(`(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`)

type Crawler struct {
    outputFileFound *os.File
    outputFileCrawled *os.File
    failedHostCount int
    foundHosts map[string]int
    initialHosts []string
    successfullHostCount int
	requestPool chan string
	parsePool chan[]string
	outputChannel chan string
	counterChannel chan counterMsg
	countStore map[string]int
	httpClient http.Client
}

type counterMsg struct {
    Key string
    Value int
}

func Instance(hosts []string, found *os.File, crawled *os.File) *Crawler {
    defer func() { 
        if err := recover(); err != nil { 
            fmt.Fprintf(os.Stderr, "Exception caught: %v\n", err)
        }
    }()
    
    tlsConfig := &tls.Config{ InsecureSkipVerify: true }                        
    transport := &http.Transport{ TLSClientConfig: tlsConfig }                                


    return &Crawler{
        httpClient: http.Client{ 
            Timeout: time.Duration(4 * time.Second),
            Transport: transport,
        },
        countStore: make(map[string]int),
        foundHosts: make(map[string]int),
        initialHosts: hosts,
        outputFileFound: found,
        outputFileCrawled: crawled,
        failedHostCount: 0,
        successfullHostCount: 0,
    }
}

func (self *Crawler) Start(workerCount int) {
    networkWorkerCount := (workerCount / 4) * 3
    parserWorkerCount := workerCount / 4
    
    self.parsePool = make(chan []string, networkWorkerCount)
    self.outputChannel = make(chan string, networkWorkerCount * 150)
    self.requestPool = make(chan string, 9999999)
    self.counterChannel = make(chan counterMsg, networkWorkerCount * 150)
    
    for id := 0; id < networkWorkerCount; id++ {
        go self.networkWorker()
    }
    
    for _, value := range self.initialHosts {
        self.requestPool <- value
    }

    go self.store()
    go self.monitor()
    
    for id := 0; id < parserWorkerCount; id++ {
        go self.parser()
    }
}

func (self *Crawler) monitor() {
    for {
        count := <-self.counterChannel
        self.countStore[count.Key] += count.Value
        
        if self.countStore[count.Key] % 500 == 0 {
            var keys []string
            
            for k := range self.countStore {
                keys = append(keys, k)
            }
            
            sort.Strings(keys)
            fmt.Printf("\r")
            
            for index, key := range keys {
                fmt.Printf(key + ": " + strconv.Itoa(self.countStore[key]))
                
                if index < len(keys) - 1{
                    fmt.Printf(" - ")
                }
            }
        }
    }
}

func (self *Crawler) store() {
    for {
        result := <-self.outputChannel

		if value, status := self.foundHosts[result]; status {
		    self.foundHosts[result] = value + 1
			self.counterChannel <- counterMsg{ Key: "Duplicates", Value: 1 }
		} else {
			self.foundHosts[result] = 1
			self.writeResult(result, self.outputFileFound)
			self.counterChannel <- counterMsg{ Key: "Unique hosts", Value: 1 }
			self.requestPool <- result
		}
    }
}

func (self *Crawler) writeResult(result string, file *os.File) {
    if file != nil {
        file.WriteString(result + "\n")
    }
}

func (self *Crawler) parser() {
    for result := range self.parsePool {
    	for _, value := range result {
    		parsed, err := url.ParseRequestURI(value)
    
    		if err != nil {
    			self.counterChannel <- counterMsg{ Key: "Failed hosts", Value: 1 }
    			continue
    		}
    
    		self.outputChannel <- parsedUrl.Scheme + "://" + parsedUrl.Host
    	}
    }
}
    
func (self *Crawler) networkWorker() {
    for job := range self.requestPool {
        self.counterChannel <- counterMsg{ Key: "Active scrapers", Value: 1 }
        self.writeResult(job, self.outputFileCrawled)

        response, err := self.httpClient.Get(job)
        
        if err != nil {
        	self.counterChannel <- counterMsg{ Key: "Failed hosts", Value: 1 }
        	self.counterChannel <- counterMsg{ Key: "Active scrapers", Value: -1 }
        	continue
        }
    
        body, err := ioutil.ReadAll(response.Body)
    
        if err != nil {
        	self.counterChannel <- counterMsg{ Key: "Failed hosts", Value: 1 }
        	self.counterChannel <- counterMsg{ Key: "Active scrapers", Value: -1 }
        	continue
        }
        
        self.counterChannel <- counterMsg{ Key: "Crawled hosts", Value: 1 }
        
        if match := urlPattern.FindAllString(string(body), -1); len(match) > 0 {
            self.parsePool <- match
        }
        
        self.counterChannel <- counterMsg{ Key: "Active scrapers", Value: -1 }
        response.Body.Close()
    }
}

func getTime() int32 {
	return int32(time.Now().Unix())
}

	
func check(e error) {
    if e != nil {
        panic(e)
    }
}

func IsUrl(url string) bool {
    return urlPattern.MatchString(url)
}