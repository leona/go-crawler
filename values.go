package crawler

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

const (
    FOUND outputType = "FOUND"
    CRAWLED = "CRAWLED"
    UPDATE = "UPDATE"
)

type logType string

const (
    ACTIVE_WORKERS logType = "Active workers" 
    FAILED = "Failed hosts"
    SUCCESSFULL = "Crawled hosts"
    UNIQUE = "Unique hosts"
    DUPLICATES = "Duplicates"
)