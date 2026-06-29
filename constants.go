package encrata

import "time"

const DefaultBaseURL = "https://api.encrata.com"

var userAgent = "encrata-go/" + Version

const (
	defaultMaxRetries = 3
	initialBackoff    = 1 * time.Second
	backoffFactor     = 2.0
	maxBackoff        = 30 * time.Second
	defaultTimeout    = 30 * time.Second
)

var retryableStatus = map[int]bool{
	429: true,
	500: true,
	502: true,
	503: true,
	504: true,
}
