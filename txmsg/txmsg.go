package txmsg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/nxlock"
	"github.com/jukylin/nx/txmsg/domain/entity"
	"github.com/jukylin/nx/txmsg/domain/repo"
	"github.com/jukylin/nx/txmsg/queue"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	TxMsgLock = "tx_msg_lock"

	HoldLockTime = 60
)

const (
	Err001 = "%d 超出最大处理次数"

	Err002 = "放到时间轮队列失败 %v"

	Err003 = "数据类型错误 %T, 期望 entity.Msg"

	Err004 = "超出重试次数 %d"

	Err005 = "没有获取到数据 %d"

	Err006 = "投递失败 %s"

	Err007 = "已过期 %v"

	Err008 = "%d 重试次数超出边界"
)

type TxMsg struct {
	logger log.Logger

	msgInfoRepo repo.MsgInfoRepo

	// 事物操作队列
	msgQueue queue.LocalQueue

	// 重试时间队列
	timeWheel queue.LocalQueue

	// 远程队列
	remoteQueue queue.RemoteQueue

	// 是否获得分布式事物锁
	holdLock bool

	// 最大重试次数
	maxDealTimes int

	nxlock *nxlock.Nxlock

	workingTaskWg *sync.WaitGroup

	scanMsgChan chan struct{}

	cleanMsgChan chan struct{}

	timeOutDate []int

	// 每次删除多少条数据
	deleteMsgOneTimeNum int

	maxDealNumOneTime int64

	// 每次补漏多少条
	limitNum int
}

type Option func(c *TxMsg)

// 在内存的数据和在数据库的数据是否会被重复消费
func NewTxMsg(options ...Option) *TxMsg {
	txMsg := &TxMsg{}

	for _, option := range options {
		option(txMsg)
	}

	txMsg.maxDealTimes = 6
	// 处理间隔数组(6次): 0s, 5s, 10s, 25s, 50s, 100s
	txMsg.timeOutDate = []int{0, 5, 10, 25, 50, 100}
	txMsg.deleteMsgOneTimeNum = 200
	txMsg.maxDealNumOneTime = 2000
	txMsg.limitNum = 50

	txMsg.scanMsgChan = make(chan struct{}, 1)
	txMsg.cleanMsgChan = make(chan struct{}, 1)

	// 投递工作
	if txMsg.msgQueue == nil {
		txMsg.msgQueue = queue.NewBoundedQueue(
			queue.WithLogger(txMsg.logger),
			queue.WithCapacity(2000),
			queue.WithConsumersNum(200),
			queue.WithName("msg_queue"),
		)
	}
	txMsg.msgQueue.Consumer(func(i interface{}) {
		err := txMsg.processor(i)
		if err != nil {
			txMsg.logger.Errorf(err.Error())
		}
	})

	// 时间轮
	if txMsg.timeWheel == nil {
		txMsg.timeWheel = queue.NewBoundedQueue(
			queue.WithLogger(txMsg.logger),
			queue.WithCapacity(1000),
			queue.WithConsumersNum(100),
			queue.WithName("time_wheel_queue"),
		)
	}
	txMsg.timeWheel.Consumer(func(i interface{}) {
		err := txMsg.processorTimeWheel(i)
		if err != nil {
			txMsg.logger.Errorf(err.Error())
		}
	})

	txMsg.workingTaskWg = &sync.WaitGroup{}

	return txMsg
}

func WithLogger(logger log.Logger) Option {
	return func(tm *TxMsg) {
		tm.logger = logger
	}
}

func WithMsgInfoRepo(msgInfoRepo repo.MsgInfoRepo) Option {
	return func(tm *TxMsg) {
		tm.msgInfoRepo = msgInfoRepo
	}
}

func WithMsgQueue(queue queue.LocalQueue) Option {
	return func(tm *TxMsg) {
		tm.msgQueue = queue
	}
}

func WithTimeWheel(queue queue.LocalQueue) Option {
	return func(tm *TxMsg) {
		tm.timeWheel = queue
	}
}

func WithRemoteQueue(queue queue.RemoteQueue) Option {
	return func(tm *TxMsg) {
		tm.remoteQueue = queue
	}
}

func (tm *TxMsg) Start(ctx context.Context) error {

	go tm.keepLockTask(ctx)

	go tm.scanMsgTask(ctx)

	go tm.cleanMsgTask(ctx)

	return nil
}

func (tm *TxMsg) Send(ctx context.Context, msg entity.Msg) error {
	tm.msgQueue.Produce(msg)
	sendTotal.Inc()
	return nil
}

func (tm *TxMsg) processor(item interface{}) error {
	msg, ok := item.(entity.Msg)
	if !ok {
		return fmt.Errorf(Err003, item)
	}

	// 往ctx 追加 tracerId
	ctx := context.Background()

	msg.HaveDealedTimes += 1
	if msg.HaveDealedTimes > tm.maxDealTimes {
		overMaxDealTimes.Inc()
		return fmt.Errorf(Err004, msg.ID)
	}

	// 是否存在数据
	msgInfo := tm.msgInfoRepo.FindById(ctx, msg.ID)
	if msgInfo.IsEmpty() {
		// 记录没有数据
		tm.putTotimeWheel(ctx, msg)
		return fmt.Errorf(Err005, msg.ID)
	}

	tm.logger.Debugc(ctx, "msgInfo from db %+v", msgInfo)

	message := tm.buildMessage(ctx, msgInfo)
	err := tm.remoteQueue.Send(ctx, message)
	if err != nil {
		tm.putTotimeWheel(ctx, msg)
		return fmt.Errorf(Err006, err.Error())
	}

	tm.logger.Debugc(ctx, "投递成功")
	processorSuccessTotal.Inc()

	// 投递生成，修改事物消息状态
	res := tm.msgInfoRepo.UpdateStatusById(ctx, msg.ID)
	if res == 0 {
		tm.logger.Warnc(ctx, "修改状态失败 %d", msg.ID)
	}

	return nil
}

// 放入到时间轮
func (tm *TxMsg) putTotimeWheel(ctx context.Context, msg entity.Msg) error {
	if msg.HaveDealedTimes < tm.maxDealTimes {
		if len(tm.timeOutDate) <= msg.HaveDealedTimes {
			return fmt.Errorf(Err008, msg.ID)
		}

		msg.NextExpireTime = time.Now().Unix() + int64(tm.timeOutDate[msg.HaveDealedTimes])
		if !tm.timeWheel.Produce(msg) {
			return fmt.Errorf(Err002, msg)
		}
		putTotimeWheelSuccessTotal.Inc()
		tm.logger.Infoc(ctx, "%+v 放入时间轮成功", msg)
		return nil
	}

	return fmt.Errorf(Err001, msg.ID)
}

// 重试投递失败的事务操作(未超时)
func (tm *TxMsg) processorTimeWheel(item interface{}) error {
	msg, ok := item.(entity.Msg)
	if !ok {
		return fmt.Errorf(Err003, item)
	}

	if msg.IsExpire() {
		return fmt.Errorf(Err007, item)
	}

	tm.logger.Debugf("%+v 重新放入事物消息队列")
	tm.msgQueue.Produce(msg)

	return nil
}

func (tm *TxMsg) buildMessage(ctx context.Context, msgInfo entity.MsgInfo) queue.Message {
	msg := queue.NewMessage()
	msg.Content = msgInfo.Content
	msg.Topic = msgInfo.Topic
	msg.Tag = msgInfo.Tag

	return msg
}

// 抢锁
func (tm *TxMsg) keepLockTask(ctx context.Context) {
	for {
		err := tm.nxlock.Lock(ctx, TxMsgLock, "1", HoldLockTime)
		if err != nil {
			tm.logger.Errorc(ctx, "获取锁失败 %s", err.Error())
			time.Sleep(10 * time.Second)
		} else {
			start := time.Now()

			tm.logger.Infoc(ctx, "成功获取锁")
			tm.workingTaskWg.Add(2)

			tm.cleanMsgChan <- struct{}{}
			tm.scanMsgChan <- struct{}{}
			// 等待任务完成
			// 使用同步方式，不适合执行长时间任务
			tm.workingTaskWg.Wait()

			// 任务完成，释放锁
			tm.nxlock.Release(ctx, TxMsgLock)

			end := time.Now()
			lab := prometheus.Labels{}
			lab["name"] = "keep_lock"
			taskSecond.With(lab).Set(end.Sub(start).Seconds())
		}
	}
}

// 补漏 扫描最近10分钟未提交事务消息,防止各种场景的消息丢失
func (tm *TxMsg) scanMsgTask(ctx context.Context) {
	for {
		select {
		case _, ok := <-tm.scanMsgChan:
			if ok {
				start := time.Now()

				tm.logger.Infoc(ctx, "补漏开始")
				var count int64
				var num int64

				for {
					msgInfos := tm.msgInfoRepo.GetWaitingMsg(ctx, tm.limitNum)
					count = int64(len(msgInfos))
					num += count

					scanMsgTotal.Add(float64(num))
					tm.logger.Debugc(ctx, "补漏数量 %d", num)

					for _, msgInfo := range msgInfos {
						message := tm.buildMessage(ctx, msgInfo)
						err := tm.remoteQueue.Send(ctx, message)
						if err != nil {
							tm.logger.Errorc(ctx, "补漏发送失败 %+v", err.Error(), msgInfo)
						} else {
							res := tm.msgInfoRepo.UpdateStatusById(ctx, msgInfo.ID)
							if res == 0 {
								tm.logger.Errorc(ctx, "补漏更新失败 %+v", msgInfo)
								tm.msgInfoRepo.UpdateStatusById(ctx, msgInfo.ID)
							}
						}
					}

					if count == 0 || num > tm.maxDealNumOneTime {
						goto WorkDone
					}
				}
			WorkDone:
				tm.workingTaskWg.Done()
				tm.logger.Infoc(ctx, "补漏完成")

				end := time.Now()
				lab := prometheus.Labels{}
				lab["name"] = "scan_msg"
				taskSecond.With(lab).Set(end.Sub(start).Seconds())
			}
		case <-ctx.Done():
			tm.logger.Infoc(ctx, "补漏退出")
			return
		}
	}
}

// 删除三天之前的发送成功的消息
func (tm *TxMsg) cleanMsgTask(ctx context.Context) {
	for {
		select {
		case _, ok := <-tm.cleanMsgChan:
			if ok {
				start := time.Now()

				tm.logger.Infoc(ctx, "清除开始")
				var count int64
				var num int64
				for {
					count = tm.msgInfoRepo.DelSendedMsg(ctx, tm.deleteMsgOneTimeNum)
					num += count
					tm.logger.Debugc(ctx, "删除数量 %d", num)
					if count == 0 || num > tm.maxDealNumOneTime {
						goto WorkDone
					}
				}
			WorkDone:
				tm.workingTaskWg.Done()
				tm.logger.Infoc(ctx, "清除完成")

				end := time.Now()
				lab := prometheus.Labels{}
				lab["name"] = "clean_msg"
				taskSecond.With(lab).Set(end.Sub(start).Seconds())
			}
		case <-ctx.Done():
			tm.logger.Infoc(ctx, "清除退出")
			return
		}
	}
}

func (tm *TxMsg) Close() {
	close(tm.scanMsgChan)
	close(tm.cleanMsgChan)
	err := tm.nxlock.Close()
	if err != nil {
		tm.logger.Errorf(err.Error())
	}
}
