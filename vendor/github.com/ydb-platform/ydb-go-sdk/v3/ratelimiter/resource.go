package ratelimiter

type Resource struct {
	ResourcePath    string
	HierarchicalDrr HierarchicalDrrSettings
}

type HierarchicalDrrSettings struct {
	MaxUnitsPerSecond       float64
	MaxBurstSizeCoefficient float64
	PrefetchCoefficient     float64
	PrefetchWatermark       float64
}
