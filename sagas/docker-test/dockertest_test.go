package docker_t

import (
	"testing"

	"github.com/jukylin/esim/log"
)

func TestDockerTest_RunReids(t *testing.T) {
	logger := log.NewLogger()
	tests := []struct {
		name string
	}{
		{"启动redis"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := NewDockerTest(logger)
			dt.RunReids()
			dt.RunMysql()
			dt.Close()
		})
	}
}
