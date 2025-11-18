package application

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	const osAddress = "osAddress:8080"
	const flagAddress = "flagAddress:8080"

	var server = Server{}
	var msg = "Server host assigned to os environment variable ADDRESS"
	os.Setenv("ADDRESS", osAddress)
	var err = server.setParams([]string{})
	assert.NoError(t, err, msg)
	assert.True(t, server.Host == osAddress, msg)

	server = Server{}
	msg = "Server host assigned to command line argument -a"
	os.Setenv("ADDRESS", "")
	err = server.setParams([]string{"-a", flagAddress})
	assert.NoError(t, err, msg)
	assert.True(t, server.Host == flagAddress, msg)

	server = Server{}
	os.Setenv("ADDRESS", osAddress)
	msg = "Server host assigned to os environment variable ADDRESS, but command line argument -a is not empty"
	err = server.setParams([]string{"-a", flagAddress})
	assert.NoError(t, err, msg)
	assert.True(t, server.Host == osAddress, msg)
}
