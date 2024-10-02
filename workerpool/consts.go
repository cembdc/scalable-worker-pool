package workerpool

const (
	DefaultMinWorkers    = 3
	DefaultMaxWorkers    = 20
	DefaultLoadThreshold = 40000

	LoadThresholdScaleDownRatio = 0.75
	LoadThresholdScaleUpRatio   = 0.25
)
