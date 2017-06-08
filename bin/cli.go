package main


import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"runtime"
	"time"
	"strings"
	"strconv"
	"bufio"
	"net/http"
	".."
)

import _ "net/http/pprof"

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    if len(os.Args) > 1 {
        go http.ListenAndServe(":8080", http.DefaultServeMux)

        fmt.Println("Running from cli")
        var workerAmount int
        var outputFilePath string
        var outputFileCrawledPath string
        var outputFile *os.File
        var outputFileCrawled *os.File
        
        if len(os.Args) > 2 {
            outputFilePath = os.Args[2] + "_output.txt"
            outputFileCrawledPath = os.Args[2] + "_history.txt"
        } else {
            outputFilePath = "./crawl_output.txt"
            outputFileCrawledPath = "./crawl_history.txt"
        }
        
        
        if len(os.Args) > 3 {
            workerAmount, _ = strconv.Atoi(os.Args[3])
        } else {
            workerAmount = runtime.NumCPU() * 10
        }

        if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
            outputFile, err = os.Create(outputFilePath)
            check(err)
        } else {
            outputFile, err = os.OpenFile(outputFilePath, os.O_APPEND, os.ModeAppend)
            check(err)
        }
        
        if _, err := os.Stat(outputFileCrawledPath); os.IsNotExist(err) {
            outputFileCrawled, err = os.Create(outputFileCrawledPath)
            check(err)
        } else {
            outputFileCrawled, err = os.OpenFile(outputFileCrawledPath, os.O_APPEND, os.ModeAppend)
            check(err)
        }
        
    	signals := make(chan os.Signal, 1)
        done := make(chan bool, 1)
        
        signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
    
        go func() {
            <-signals
            fmt.Println()
            done <- true
        }()
    
        startTime := getTime()
        fmt.Println("Starting", workerAmount, "threads")
        
        crawler := *crawler.Instance([]string{os.Args[1]}, outputFile, outputFileCrawled)
        crawler.Start(workerAmount)
        
        fmt.Println("Press F + \\r to finish crawling gracefully")
        scanner := bufio.NewScanner(os.Stdin)
        
        for scanner.Scan() {
            text := scanner.Text()

            if strings.ToLower(text) == "f" {
                crawler.Stop()
                os.Exit(1)
                done <- true
            }
        }
        
        <-done
    	fmt.Println("Finished in:", getTime() - startTime, "second(s)")
    	os.Exit(1)
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