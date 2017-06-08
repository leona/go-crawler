package crawler

const (
    FOUND outputType = "FOUND"
    CRAWLED = "CRAWLED"
    UPDATE = "UPDATE"
)

const (
    ACTIVE_WORKERS logType = "Active workers" 
    FAILED = "Failed hosts"
    SUCCESSFULL = "Crawled hosts"
    UNIQUE = "Unique hosts"
    DUPLICATES = "Duplicates"
)

const (
    URL_REGEX = `(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`
)