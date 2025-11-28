package repository

import (
	"os"
	"testing"

	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestMetricMemo(t *testing.T) {
	test(t, NewMemRepository())
}

func TestMetricFile(t *testing.T) {
	fileName := "./metrics.json"
	os.Remove(fileName)
	test(t, NewFileRepository(fileName))
}

func test(t *testing.T, repo Metric) {
	t.Logf("repository=%+v", repo)

	_, err := service.NameValueToModel("", int64(1))
	assert.Error(t, err, "empty metric name")

	_, err = service.NameValueToModel("ololo", nil)
	assert.Error(t, err, "empty metric value")

	_, err = service.NameValueToModel("ololo", "ololo")
	assert.Error(t, err, "invalid metric type")

	m, err := service.NameValueToModel("float64", float64(100.500))
	assert.NoError(t, err, "float64 to model")
	err = repo.SetOne(*m)
	assert.NoError(t, err, "float64 save one")
	m2, err2 := repo.Get(m.ID)
	assert.NoError(t, err2, "float64 get")
	assert.True(t, *m.Value == *m2.Value, "float64 saved")

	m, err = service.NameValueToModel("float64", float64(100.600))
	assert.NoError(t, err, "float64-2 to model")
	mm := make(map[string]model.Metric)
	mm[m.ID] = *m
	err = repo.SetList(mm)
	assert.NoError(t, err, "float64-2 save all")
	mm2, err2 := repo.GetAll()
	assert.NoError(t, err2, "float64-2 get all")
	m3, ok := mm2[m.ID]
	assert.True(t, ok, "float64-2 model in map")
	assert.True(t, *(m.Value) == *(m3.Value), "float64-2 saved")
	t.Logf("m=%+v,value=%v,m3=%+v,value=%v", m, *(m.Value), m3, *(m3.Value))

	m, err = service.NameValueToModel("int64", int64(1))
	assert.NoError(t, err, "int64 to model")
	err = repo.SetOne(*m)
	assert.NoError(t, err, "int64 save one")
	m2, err2 = repo.Get(m.ID)
	assert.NoError(t, err2, "int64 get")
	assert.True(t, m2 != nil && *(m.Delta) == *(m2.Delta), "int64 saved")
	t.Logf("m=%+v,delta=%v,m2=%+v,delta=%v", m, *m.Delta, m2, *m2.Delta)
	i := *m2.Delta

	j := int64(2)
	m, err = service.NameValueToModel("int64", j)
	assert.NoError(t, err, "int64-2 to model")
	err = repo.SetOne(*m)
	assert.NoError(t, err, "int64-2 save one")
	m2, err2 = repo.Get(m.ID)
	assert.NoError(t, err2, "int64-2 get")
	assert.True(t, (m2 != nil) && (*m2.Delta == (i+j)), "int64-2 saved")
	t.Logf("m=%+v,delta_old=%v,m2=%+v,delta_new=%v", m, *m.Delta, m2, *m2.Delta)
}
