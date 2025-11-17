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
	s.SetHost(host)
	go s.Run([]string{})

	var c = Agent{}
	c.SetHost(host)
	c.SetReportInterval(5)
	c.SetPollInterval(2)
	go c.Run([]string{})

	time.Sleep(time.Duration(10) * time.Second)
	var httpc = http.Client{Timeout: time.Duration(1) * time.Second}
	var metrics = ""
	resp, err := httpc.Get(fmt.Sprintf("http://%s", host))
	log.Printf("status=%d, len=%d, error=%+v", resp.StatusCode, resp.ContentLength, err)
	if err == nil && resp != nil {
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Err(err).Msg("read response exception")
		}
		metrics = string(b)
		log.Printf("%s", metrics)
	} else {
		log.Err(err).Msg("Get exception")
	}
	assert.NoError(t, err, "integration")
	assert.True(t, len(strings.Split(metrics, "\n")) == 27)
}
