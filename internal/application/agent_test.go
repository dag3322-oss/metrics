package application

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgent(t *testing.T) {
	const osAddress = "osAddress:8080"
	const flagAddress = "flagAddress:8080"
	const osReportInterval = 10
	const flagReportInterval = 20
	const osPollInterval = 30
	const flagPollInterval = 40

	var agent = Agent{}
	var msg = "Server host assigned to os environment variable ADDRESS"
	os.Setenv("ADDRESS", osAddress)
	var err = agent.setParams([]string{})
	assert.NoError(t, err, msg)
	assert.True(t, agent.host == osAddress, msg)

	agent = Agent{}
	msg = "Server host assigned to command line argument -a"
	os.Setenv("ADDRESS", "")
	err = agent.setParams([]string{"-a", flagAddress})
	assert.NoError(t, err, msg)
	assert.True(t, agent.host == flagAddress, msg)

	agent = Agent{}
	os.Setenv("ADDRESS", osAddress)
	msg = "Server host assigned to os environment variable ADDRESS, but command line argument -a is not empty"
	err = agent.setParams([]string{"-a", flagAddress})
	assert.NoError(t, err, msg)
	assert.True(t, agent.host == osAddress, msg)

	agent = Agent{}
	msg = "ReportInterval not assigned from os environment variable REPORT_INTERVAL"
	os.Setenv("REPORT_INTERVAL", "string")
	err = agent.setParams([]string{})
	assert.Error(t, err, msg)

	agent = Agent{}
	msg = "ReportInterval assigned from os environment variable REPORT_INTERVAL"
	os.Setenv("REPORT_INTERVAL", fmt.Sprintf("%d", osReportInterval))
	err = agent.setParams([]string{})
	assert.NoError(t, err, msg)
	assert.True(t, agent.reportInterval == osReportInterval, msg)

	agent = Agent{}
	msg = "ReportInterval assigned from flag -r"
	os.Setenv("REPORT_INTERVAL", "")
	err = agent.setParams([]string{"-r", fmt.Sprintf("%d", flagReportInterval)})
	assert.NoError(t, err, msg)
	assert.True(t, agent.reportInterval == flagReportInterval, msg)

	agent = Agent{}
	msg = "ReportInterval assigned from os environment variable REPORT_INTERVAL but flag -r defined"
	os.Setenv("REPORT_INTERVAL", fmt.Sprintf("%d", osReportInterval))
	err = agent.setParams([]string{"-r", fmt.Sprintf("%d", flagReportInterval)})
	assert.NoError(t, err, msg)
	assert.True(t, agent.reportInterval == osReportInterval, msg)

	agent = Agent{}
	msg = "PollInterval not assigned from os environment variable POLL_INTERVAL"
	os.Setenv("POLL_INTERVAL", "string")
	err = agent.setParams([]string{})
	assert.Error(t, err, msg)

	agent = Agent{}
	msg = "PollInterval assigned from os environment variable POLL_INTERVAL"
	os.Setenv("POLL_INTERVAL", fmt.Sprintf("%d", osPollInterval))
	err = agent.setParams([]string{})
	assert.NoError(t, err, msg)
	assert.True(t, agent.pollInterval == osPollInterval, msg)

	agent = Agent{}
	msg = "PollInterval assigned from flag -p"
	os.Setenv("POLL_INTERVAL", "")
	err = agent.setParams([]string{"-p", fmt.Sprintf("%d", flagPollInterval)})
	assert.NoError(t, err, msg)
	assert.True(t, agent.pollInterval == flagPollInterval, msg)

	agent = Agent{}
	msg = "PollInterval assigned from os environment variable POLL_INTERVAL but flag -p defined"
	os.Setenv("POLL_INTERVAL", fmt.Sprintf("%d", osPollInterval))
	err = agent.setParams([]string{"-p", fmt.Sprintf("%d", flagPollInterval)})
	assert.NoError(t, err, msg)
	assert.True(t, agent.pollInterval == osPollInterval, msg)

}
