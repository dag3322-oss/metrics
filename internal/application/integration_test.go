package application

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	var host = "localhost:8080"

	var s = Server{}
	s.Host = host
	go s.Run([]string{})

	var a = Agent{}
	a.Host = host
	var i int64 = 5
	a.ReportInterval = &i
	var j int64 = 2
	a.PollInterval = &j
	go a.Run([]string{})

	time.Sleep(time.Duration(10) * time.Second)
	var httpc = http.Client{Timeout: time.Duration(1) * time.Second}
	var metrics = ""
	resp, err := httpc.Get(fmt.Sprintf("http://%s", host))
	log.Debug().Msg(fmt.Sprintf("status=%d, len=%d, error=%+v", resp.StatusCode, resp.ContentLength, err))
	if err == nil && resp != nil {
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Err(err).Msg("read response exception")
		}
		metrics = string(b)
		log.Debug().Msg(fmt.Sprintf("%s", metrics))
	} else {
		log.Err(err).Msg("Get exception")
	}
	assert.NoError(t, err, "integration")
	ii := len(strings.Split(metrics, "\n"))
	assert.True(t, ii == 29, fmt.Sprintf("metrics collection size=%d", ii))
}
