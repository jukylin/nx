package txmsg

import (
	"github.com/prometheus/client_golang/prometheus"
)

var putTotimeWheelSuccessTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "put_to_time_wheel_success_total",
		Help: "放入时间轮数量",
	},
)

var taskSecond = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "task_second",
		Help: "任务耗时",
	},
	[]string{"name"},
)

var processorSuccessTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "processor_success_total",
		Help: "处理成功数量",
	},
)

var scanMsgTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "scan_msg_total",
		Help: "补漏数量",
	},
)

var sendTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "send_total",
		Help: "接收的消息数量",
	},
)

var overMaxDealTimes = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "over_max_deal_times",
		Help: "超出重试的消息数量",
	},
)

func init() {
	prometheus.MustRegister(putTotimeWheelSuccessTotal)
	prometheus.MustRegister(taskSecond)
	prometheus.MustRegister(processorSuccessTotal)
	prometheus.MustRegister(scanMsgTotal)
	prometheus.MustRegister(overMaxDealTimes)
}
