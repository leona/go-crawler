package crawler

import (
	"time"
    "crypto/tls"
	"net/http"
	"os"
    "encoding/csv"
    "bufio"
    "io"
    //"fmt"
)
// https://groups.google.com/forum/#!topic/golang-nuts/-pqkICuokio
func removeDuplicates(a []string) []string {
        result := []string{}
        seen := map[string]string{}
        
        for _, val := range a {
            if _, ok := seen[val]; !ok {
                result = append(result, val)
                seen[val] = val
            }
        }
        return result
} 

func createRequestClient() http.Client {
    tlsConfig := &tls.Config{ InsecureSkipVerify: true }                        
    transport := &http.Transport{ TLSClientConfig: tlsConfig }   
    
    return http.Client{ 
        Timeout: time.Duration(4 * time.Second),
        Transport: transport,
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

//https://stackoverflow.com/questions/39442167/convert-int32-to-string-in-golang
func String(n int32) string {
    buf := [11]byte{}
    pos := len(buf)
    i := int64(n)
    signed := i < 0
    if signed {
        i = -i
    }
    for {
        pos--
        buf[pos], i = '0'+byte(i%10), i/10
        if i == 0 {
            if signed {
                pos--
                buf[pos] = '-'
            }
            return string(buf[pos:])
        }
    }
}

func loadCsv(file *os.File) [][]string {
    reader := csv.NewReader(bufio.NewReader(file))
    var output [][]string
    
    for {
        record, err := reader.Read()

        if err == io.EOF {
            break
        }

        var line []string
        
        for value := range record {
            line = append(line, record[value])
        }
        
        output = append(output, line)
    }

    return output
}