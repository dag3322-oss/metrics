package helper

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func NewRetryableClient(hashKey *string) *RetryableClient {
	c := new(RetryableClient)
	c.hashKey = hashKey
	c.client = &http.Client{}
	c.client.Timeout = 30 * time.Second
	return c
}

type RetryableClient struct {
	client  *http.Client
	hashKey *string
}

func (c RetryableClient) Do(req *http.Request) (resp *http.Response, err error) {
	if c.hashKey != nil {
		r := req.Body
		defer r.Close()
		if r != nil {
			b, err := io.ReadAll(r)
			if err != nil {
				return nil, err
			}
			bodyHash := sha256.Sum256(append(b[:], []byte(*c.hashKey)...))
			s := base64.StdEncoding.EncodeToString(bodyHash[:])
			log.Debug().Str("header", "\""+s+"\"").Msg("sha256")
			req.Header.Add(HashHeaderName, s)
			req.Body = io.NopCloser(bytes.NewReader(b))
		}
	}

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
