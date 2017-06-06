package crawler

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"regexp"
	"strings"
	"os"
	"os/signal"
	"net/url"
	"strconv"
	"sort"
    "syscall"
)

// TODO
// Write Stop method
// Pick up from the History file
// Database/redis support

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

func Instance(hosts []string, found *os.File, crawled *os.File) *Crawler {
    urlPattern, err := regexp.Compile(`(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`)

    if err != nil {
        panic("urlPattern will not compile")
    }

    return &Crawler{
        httpClient: createRequestClient(),
        countStore: make(map[logType]int),
        foundHosts: make(map[string]int32),
        queues: make(map[outputType][][]string),
        initialHosts: hosts,
        outputFileFound: found,
        outputFileCrawled: crawled,
        urlPattern: urlPattern,
        state: false,
    }
}

func (self *Crawler) Start(workerCount int) {
    self.state = true
    
    networkWorkerCount := (workerCount / 4) * 3
    parserWorkerCount := workerCount / 4
    
    self.parsePool = make(chan []string, networkWorkerCount)
    self.outputChannel = make(chan writeRequest, networkWorkerCount * 150)
    self.requestPool = make(chan string, 9999999)
    self.counterChannel = make(chan counterMsg, networkWorkerCount * 150)
    
    for id := 0; id < networkWorkerCount; id++ {
        go self.networkWorker()
    }
    
    self.loadHistory()
    go self.outputWorker()
    go self.monitor()
    
    for id := 0; id < parserWorkerCount; id++ {
        go self.parser()
    }
    
    signalChannel := make(chan os.Signal, 100000)
    signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
    
    go func(){
        for sig := range signalChannel {  
            self.Stop()
            fmt.Println("\n\nSignal received:", sig)
            os.Exit(1)
        }
    }()
    
    for _, value := range self.initialHosts {
        self.requestPool <- value
    }
}

func (self *Crawler) Stop() {
    self.state = false
    close(self.requestPool)
    queue := self.mergeQueue(FOUND)
    self.writeQueue(FOUND, queue)
    queue = self.mergeQueue(CRAWLED)
    self.writeQueue(CRAWLED, queue)
    
    self.outputFileCrawled.Close()
    self.outputFileFound.Close()
    workers := self.countStore[ACTIVE_WORKERS]
    
    fmt.Printf("\nRemaining workers: " + strconv.Itoa(workers) + "\n")
    fmt.Println("Cleaned up queues and file handlers")
    fmt.Println("Final results -", "Crawled:", self.countStore[SUCCESSFULL], "Found:", self.countStore[UNIQUE])
}

func (self *Crawler) loadHistory() {
    fmt.Println("Loading history...")
    foundHosts := loadCsv(self.outputFileFound)
    crawledHosts := loadCsv(self.outputFileCrawled)
    
    for _, value := range foundHosts {
        integer, err := strconv.Atoi(value[1])
        
        if err != nil {
            fmt.Println("Error parsing string to Int:", value[1])
            continue
        }
        
        self.foundHosts[value[0]] = int32(integer)
    }
    
    queueHosts := self.foundHosts
    
    for _, value := range crawledHosts {
        delete(queueHosts, value[0])
    }
    
    if len(queueHosts) > 0 {
        self.initialHosts = []string{}
    }
    
    for index, _ := range queueHosts {
        self.initialHosts = append(self.initialHosts, index)
    }
    
    fmt.Println("Finished loading results from last crawl")
}

func (self *Crawler) monitor() {
    // TODO: Print req/s
    for {
        count := <-self.counterChannel
        self.countStore[count.Key] += count.Value
        
        if self.countStore[count.Key] % 1000 == 0 {
            var keys []string
            
            for key := range self.countStore {
                keys = append(keys, string(key))
            }
            
            sort.Strings(keys)
            fmt.Printf("\r")
            
            for index, key := range keys {
                fmt.Printf(string(key) + ": " + strconv.Itoa(self.countStore[logType(key)]))
                
                if index < len(keys) - 1{
                    fmt.Printf(" - ")
                }
            }
        }
    }
}

func (self *Crawler) outputWorker() {
    for {
        request := <- self.outputChannel
        
        switch request.file {
            case FOUND:
        		if _, status := self.foundHosts[request.host]; status {
        		    self.foundHosts[request.host] += 1
        			self.counterChannel <- counterMsg{ Key: DUPLICATES, Value: 1 }
        		} else {
        			self.foundHosts[request.host] = 1
        			self.counterChannel <- counterMsg{ Key: UNIQUE, Value: 1 }
        			
        			if self.state {
        			    self.requestPool <- request.host
        			}
        			self.requestWrite(request)
        		}
        	case CRAWLED:
        	    self.requestWrite(request)
        }
		
    }
}

func (self *Crawler) requestWrite(request writeRequest) {
    if _, status := self.queues[request.file]; !status {
        self.queues[request.file] = [][]string{}
    }
    
    if queue := self.queues[request.file]; len(queue) <= 500 {
        backlinks := self.foundHosts[request.host]
        self.queues[request.file] = append(queue, []string{request.host, String(backlinks), String(getTime())})
    } else {
        self.writeQueue(request.file, self.mergeQueue(request.file))
    }
}

func (self *Crawler) mergeQueue(queueType outputType) string {
    output := ""
    queue := self.queues[queueType]
    self.queues[queueType] = [][]string{}
    
    for _, value := range queue {
        output += strings.Join(value[:], ",") + "\n"
    }
    
    return output
}

func (self *Crawler) writeQueue(queueType outputType, queue string) {
    switch queueType {
        case UPDATE:
            // TODO
        case FOUND:
            self.writeResultGroup(self.outputFileFound, queue)
        case CRAWLED:
            self.writeResultGroup(self.outputFileCrawled, queue)
    }
}

func (self *Crawler) writeResultGroup(file *os.File, results string, ) {
    if file != nil {
        file.WriteString(results)
    }
}

func (self *Crawler) writeResult(file *os.File, results...string, ) {
    if file != nil {
        return
        output := strings.Join(results[:],",")
        file.WriteString(output + "\n")
    }
}

func (self *Crawler) parser() {
    for result := range self.parsePool {
    	for _, value := range result {
    		parsed, err := url.ParseRequestURI(value)
    
    		if err != nil {
    			self.counterChannel <- counterMsg{ Key: FAILED, Value: 1 }
    			continue
    		}
    		
            host := parsed.Scheme + "://" + strings.ToLower(parsed.Host)
    		self.outputChannel <- writeRequest{ file: FOUND, host: host}
    	}
    }
}
    
func (self *Crawler) networkWorker() {
    for job := range self.requestPool {
        self.counterChannel <- counterMsg{ Key: ACTIVE_WORKERS, Value: 1 }
		self.outputChannel <- writeRequest{ file: CRAWLED, host: job }
			
        response, err := self.httpClient.Get(job)
        
        if err != nil {
            self.counterChannel <- counterMsg{ Key: ACTIVE_WORKERS, Value: -1 }
        	self.counterChannel <- counterMsg{ Key: FAILED, Value: 1 }
        	continue
        }
    
        body, err := ioutil.ReadAll(response.Body)
    
        if err != nil {
        	self.counterChannel <- counterMsg{ Key: ACTIVE_WORKERS, Value: -1 }
        	self.counterChannel <- counterMsg{ Key: FAILED, Value: 1 }
        	continue
        }
        
        self.counterChannel <- counterMsg{ Key: SUCCESSFULL, Value: 1 }
        
        if match := self.urlPattern.FindAllString(string(body), -1); len(match) > 0 {
            self.parsePool <- removeDuplicates(match)
        }
        
        response.Body.Close()
        self.counterChannel <- counterMsg{ Key: ACTIVE_WORKERS, Value: -1 }
    }
}

func (self *Crawler) IsUrl(url string) bool {
    return self.urlPattern.MatchString(url)
}




