package workerpool

import "time"

const (
	DefaultMinWorkers    = 3
	DefaultMaxWorkers    = 20
	DefaultLoadThreshold = 40000
	DefaultWorkerTimeout = 10 * time.Millisecond

	LoadThresholdScaleDownRatio = 0.75
	LoadThresholdScaleUpRatio   = 0.25
)
