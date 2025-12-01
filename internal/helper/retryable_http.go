package helper

import (
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func NewRetryableClient() *RetryableClient {
	c := new(RetryableClient)
	c.client = &http.Client{}
	c.client.Timeout = 30 * time.Second
	return c
}

type RetryableClient struct {
	client *http.Client
}

func (c RetryableClient) Do(req *http.Request) (resp *http.Response, err error) {
	i := 0
	for {
		time.Sleep(time.Duration(i) * time.Second)
		resp, err = c.client.Do(req)
		if err != nil {
			if _, ok := err.(net.Error); ok {
				switch i {
				case 0:
					i = 1
				default:
					i = i + 2
				}
				if i <= 5 {
					log.Debug().Int("delay", i).Msg("repeat after timeout")
					continue
				}
			} else {
				log.Debug().Msg("not network error")
			}
			return nil, err
		}
		return resp, nil
	}

}
