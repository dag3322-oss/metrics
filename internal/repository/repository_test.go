package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

var repo MemRepository = NewMemRepository()

func TestMetrics(t *testing.T) {
	var err = repo.UpdateMetric("", int64(1))
	assert.True(t, err != nil, "empty metric name")

	err = repo.UpdateMetric("ololo", nil)
	assert.True(t, err != nil, "empty metric value")

	err = repo.UpdateMetric("ololo", "ololo")
	assert.True(t, err != nil, "invalid metric type")

	err = repo.UpdateMetric("float64", float64(100.500))
	assert.True(t, err == nil && repo.metrics["float64"] == float64(100.500), "float64 1")

	err = repo.UpdateMetric("float64", float64(100.600))
	assert.True(t, err == nil && repo.metrics["float64"] == float64(100.600), "float64 1")

	err = repo.UpdateMetric("int64", int64(1))
	assert.True(t, err == nil && repo.metrics["int64"] == int64(1), "int64 1")

	err = repo.UpdateMetric("int64", int64(2))
	assert.True(t, err == nil && repo.metrics["int64"] == int64(3), "int64 1")
}
