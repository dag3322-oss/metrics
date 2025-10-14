package metrics

type Repository interface {
	UpdateMetric(name string, value any) error
	GetAllAsString() string
	GetAll() map[string]any
}
