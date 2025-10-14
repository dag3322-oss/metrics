package application

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerAndAgent(t *testing.T) {
	var s = Server{}
	go s.Run()

	var c = Agent{}
	go c.Run()

	time.Sleep(time.Duration(15) * time.Second)
	var httpc = http.Client{Timeout: time.Duration(1) * time.Second}
	var metrics = ""
	resp, err := httpc.Get("http://localhost:8080/list")
	log.SetOutput(os.Stdout)
	log.Printf("status=%d, len=%d, error=%+v", resp.StatusCode, resp.ContentLength, err)
	if err == nil && resp != nil {
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		metrics = string(b)
		log.Printf("%s", metrics)
	}

	assert.True(t, len(strings.Split(metrics, "\n")) == 27)
}
