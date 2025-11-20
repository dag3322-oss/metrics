package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
	"github.com/dag3322-oss/metrics/internal/model"
	"github.com/dag3322-oss/metrics/internal/service"
)

func TestMetricMemo(t *testing.T) {
	test(t, NewMemRepository())
}

func test(t *testing.T, repo Metric) {
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
	assert.True(t, *m == *m2, "float64 saved")

	m, err = service.NameValueToModel("float64", float64(100.600))
	assert.NoError(t, err, "float64-2 to model")
	mp := make(map[string]model.Metric)
	mp[m.ID] = *m
	err = repo.SetList(mp)
	assert.NoError(t, err, "float64-2 save all")
	mp, err2 = repo.GetAll()
	assert.NoError(t, err2, "float64-2 get all")
	assert.True(t, *m == mp[m.ID], "float64-2 saved")

	m, err = service.NameValueToModel("int64", int64(1))
	assert.NoError(t, err, "int64 to model")
	err = repo.SetOne(*m)
	assert.NoError(t, err, "int64 save one")
	m2, err2 = repo.Get(m.ID)
	assert.NoError(t, err2, "int64 get")
	assert.True(t, *m == *m2 && *m2.Delta == int64(1), "int64 saved")

	m, err = service.NameValueToModel("int64", int64(2))
	assert.NoError(t, err, "int642 to model")
	err = repo.SetOne(*m)
	assert.NoError(t, err, "int64-2 save one")
	m2, err2 = repo.Get(m.ID)
	assert.NoError(t, err2, "int64-2 get")
	assert.True(t, *m == *m2 && *m2.Delta == int64(2), "int64-2 saved")
}
