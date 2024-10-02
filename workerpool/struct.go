package workerpool

import "time"

// Request represents a request to be processed by a worker.
type Request struct {
	Handler    RequestHandler
	Type       int
	Data       interface{}
	Timeout    time.Duration // Timeout duration for the request
	Retries    int           // Number of retries
	MaxRetries int           // Max number of retries
}

// RequestHandler defines a function type for handling requests.
type RequestHandler func(interface{}) error
