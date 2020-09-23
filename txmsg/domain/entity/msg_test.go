package entity

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestMsg_SetHaveDealedTimes(t *testing.T) {
	tests := []struct {
		name   string
	}{
		{"incr处理次数"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMsg(1, "")
			m.SetHaveDealedTimes()
			assert.Equal(t, 2, m.HaveDealedTimes)
		})
	}
}
