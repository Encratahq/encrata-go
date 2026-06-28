package encrata

import "time"

const DefaultBaseURL = "https://api.encrata.com"

var userAgent = "encrata-go/" + Version

const (
	defaultMaxRetries = 3                // max retry attempts for failed requests
	initialBackoff    = 1 * time.Second  // first backoff delay
	backoffFactor     = 2.0              // exponential growth factor
	maxBackoff        = 30 * time.Second // hard ceiling per sleep
	defaultTimeout    = 30 * time.Second // per-request HTTP timeout
)

// retryableStatus is the set of HTTP status codes that are safe to retry.
var retryableStatus = map[int]bool{
	429: true, // Too Many Requests
	500: true, // Internal Server Error
	502: true, // Bad Gateway
	503: true, // Service Unavailable
	504: true, // Gateway Timeout
}
