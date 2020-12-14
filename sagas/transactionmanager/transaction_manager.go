package transactionmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/jukylin/esim/http"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/nxlock"
	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/nx/sagas/domain/repo"
	"github.com/jukylin/esim/pkg/tracer-id"
)

const (
	SagasCompensate = "sagas:compensate"

	SagasCompensateQueue = "sagas:compensate:queue"

	SagasCompensateExec = "sagas:compensate:exec"

	SagasCompensateBuild = "sagas:compensate:build"

	// 没有获取锁
	CompensateLockStateZore = 0

	// 获取失败
	CompensateLockStateFail = 0

	// 获取成功
	CompensateLockStateSucc = 0
)

// 补偿任务锁状态信息
type CompensateLockState struct {
	state int

	// 成功才有创建时间，释放清0
	ctime time.Time
}

type TmOption func(*TransactionManager)

type TransactionManager struct {
	logger log.Logger

	txrecordRepo repo.TxrecordRepo

	txgroupRepo repo.TxgroupRepo

	txcompensateRepo repo.TxcompensateRepo

	httpClient *http.Client

	nxlock *nxlock.Nxlock

	queue *queue.BoundedQueue

	compensate Compensate

	cls *CompensateLockState

	// 消费txgroup
	consumer func(item interface{})
}

func NewTransactionManager(options ...TmOption) *TransactionManager {
	tm := &TransactionManager{}

	for _, option := range options {
		option(tm)
	}

	// 被丢弃的数据会重新从数据库获取
	tm.queue = queue.NewBoundedQueue(1000, nil)

	if tm.consumer == nil {
		tm.consumer = tm.execCompensate
	}
	tm.queue.StartConsumers(200, tm.consumer)

	tm.cls = &CompensateLockState{}

	return tm
}

func WithTmLogger(logger log.Logger) TmOption {
	return func(tm *TransactionManager) {
		tm.logger = logger
	}
}

func WithTmTxgroupRepo(txgroupRepo repo.TxgroupRepo) TmOption {
	return func(tm *TransactionManager) {
		tm.txgroupRepo = txgroupRepo
	}
}

func WithTmTxrecordRepo(txrecordRepo repo.TxrecordRepo) TmOption {
	return func(tm *TransactionManager) {
		tm.txrecordRepo = txrecordRepo
	}
}

func WithTmTxcompensateRepo(txcompensateRepo repo.TxcompensateRepo) TmOption {
	return func(tm *TransactionManager) {
		tm.txcompensateRepo = txcompensateRepo
	}
}

func WithTmNxLock(nxLock *nxlock.Nxlock) TmOption {
	return func(tm *TransactionManager) {
		tm.nxlock = nxLock
	}
}

func WithTmTxgroupConsumer(consumer func(item interface{})) TmOption {
	return func(tm *TransactionManager) {
		tm.consumer = consumer
	}
}

func WithTmCompensate(compensate Compensate) TmOption {
	return func(tm *TransactionManager) {
		tm.compensate = compensate
	}
}

func (tm *TransactionManager) Start(ctx context.Context) error {
	go tm.getCompensateToQueue(ctx)

	go tm.buildCompensate(ctx)

	return nil
}

// 获取200个随机事务组(需要补偿) 添加到事务队列中
func (tm *TransactionManager) getCompensateToQueue(ctx context.Context) {
	tm.logger.Debugc(ctx, "开始 getCompensateToQueue")
	ticker := time.NewTicker(1 * time.Second)
	var err error
	for {
		select {
		case <-ticker.C:
			err = tm.nxlock.Lock(ctx, SagasCompensate, 3)
			if err != nil {
				tm.logger.Errorc(ctx, err.Error())
				tm.cls.state = CompensateLockStateFail
			} else {
				tm.cls.state = CompensateLockStateSucc
				tm.cls.ctime = time.Now()

				txgroups, err := tm.txgroupRepo.GetCompensateList(context.Background(), 200)
				if err != nil {
					tm.logger.Errorc(ctx, err.Error())
				} else {
					tm.logger.Debugc(ctx, "%+v", txgroups)
					for _, txgroup := range txgroups {
						// err = tm.nxlock.Lock(ctx, fmt.Sprintf("%s:%d", SagasCompensateQueue, txgroup.Txid), 10)
						if err == nil {
							tm.queue.Produce(txgroup)
						} else {
							tm.logger.Errorc(ctx, err.Error())
						}
						// tm.nxlock.Release(ctx, fmt.Sprintf("%s:%d", SagasCompensateQueue, txgroup.Txid))
					}
				}
			}
			tm.nxlock.Release(ctx, SagasCompensate)
		case <-ctx.Done():
			tm.logger.Infoc(ctx, "补偿任务退出.")
			ticker.Stop()
			return
		}
	}
}

// 每隔1秒 随机获取 1000个 1小时未释放的 事务组列表
func (tm *TransactionManager) buildCompensate(ctx context.Context) {
	tm.logger.Debugc(ctx, "开始 buildCompensate")
	ticker := time.NewTicker(1 * time.Second)
	var err error
	for {
		select {
		case <-ticker.C:
			err = tm.nxlock.Lock(ctx, SagasCompensateBuild, 3)
			if err != nil {
				tm.logger.Errorc(ctx, err.Error())
				continue
			}

			txgroups, err := tm.txgroupRepo.GetUnfishedTransactionGroup(context.Background(), 3600)
			if err != nil {
				tm.logger.Errorc(ctx, err.Error())
			} else {
				for _, txgroup := range txgroups {
					tm.compensate.BuildCompensate(ctx, txgroup)
				}
			}

			tm.nxlock.Release(ctx, SagasCompensateBuild)
		case <-ctx.Done():
			tm.logger.Infoc(ctx, "建立补偿任务退出.")
			ticker.Stop()
			return
		}
	}
}

// 执行补偿
func (tm *TransactionManager) execCompensate(item interface{}) {
	txgroup, ok := item.(entity.Txgroup)
	ctx := context.Background()
	if !ok {
		tm.logger.Errorc(ctx, "类型错误 ： %T，期望 entity.Txgroup", item)
		return
	}

	context.WithValue(ctx, tracerid.ActiveEsimKey, txgroup.Txid)

	err := tm.nxlock.Lock(ctx, fmt.Sprintf("%s:%d", SagasCompensateExec, txgroup.Txid), 10)
	if err != nil {
		tm.logger.Errorc(ctx, err.Error())
		return
	}

	tm.logger.Infoc(ctx, "Start execCompensate")

	err = tm.compensate.ExeCompensate(ctx, txgroup)
	if err != nil {
		tm.logger.Errorc(ctx, err.Error())
		return
	}

	tm.logger.Infoc(ctx, "End execCompensate")

	err = tm.nxlock.Release(ctx, fmt.Sprintf("%s:%d", SagasCompensateExec, txgroup.Txid))
	if err != nil {
		tm.logger.Errorc(ctx, err.Error())
	}
}

func (tm *TransactionManager) Close() {
	tm.nxlock.Close()
	tm.queue.Stop()
}
