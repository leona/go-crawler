package crawler

import (
	"fmt"
	"io/ioutil"
	"strings"
	"net/url"
	"strconv"
	"sort"
)

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