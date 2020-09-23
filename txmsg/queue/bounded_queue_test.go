package queue

import (
	"study-go/txmsg/domain/entity"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/log"
	"time"
	"os"
)

var logger log.Logger

func TestMain(m *testing.M) {
	logger = log.NewLogger(
		log.WithDebug(true),
	)

	code := m.Run()

	os.Exit(code)
}

func TestBoundedQueue_Produce(t *testing.T) {
	type args struct {
		msg entity.Msg
	}
	tests := []struct {
		name   string
		args   args
		want   bool
	}{
		{"生产1条数据", args{entity.Msg{ID:1}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bq := NewBoundedQueue(
				WithLogger(logger),
				WithName("Produce"),
				WithReportPeriod(100 * time.Millisecond),
			)
			bq.Consumer(func(i interface{}) {
				assert.Equal(t, int64(1), i.(entity.Msg).ID)
			})
			if got := bq.Produce(tt.args.msg); got != tt.want {
				t.Errorf("BoundedQueue.Produce() = %v, want %v", got, tt.want)
			}
			bq.Close()
			time.Sleep(1 * time.Second)
		})
	}
}
