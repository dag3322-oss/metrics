package metrics

type Repository interface {
	UpdateMetric(name string, value any) error
	GetMetrics() string
}
