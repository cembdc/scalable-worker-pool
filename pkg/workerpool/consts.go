package workerpool

import "time"

const (
	DefaultMinWorkers    = 3
	DefaultMaxWorkers    = 20
	DefaultLoadThreshold = 5
	DefaultWorkerTimeout = 10 * time.Millisecond
	DefaultBufferSize    = 50000
	DefaultRequestType   = 1
	DefaultTimeout       = 5 * time.Second
	DefaultMaxRetries    = 3

	LoadThresholdScaleDownRatio = 0.75
	LoadThresholdScaleUpRatio   = 0.25
)
