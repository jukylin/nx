package transactionmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/jukylin/esim/http"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/nxlock"
	"github.com/jukylin/nx/saga/domain/entity"
	"github.com/jukylin/nx/saga/domain/repo"
)

const (
	SagasCompensate = "sagas:compensate"

	SagasCompensateQueue = "sagas:compensate:queue"

	SagasCompensateExec = "sagas:compensate:exec"
)

type TmOption func(*TransactionManager)

type TransactionManager struct {
	logger log.Logger

	txrecordRepo repo.TxrecordRepo

	txgroupRepo repo.TxgroupRepo

	txcompensateRepo repo.TxcompensateRepo

	httpClient *http.Client

	nxlock *nxlock.Nxlock

	queue *queue.BoundedQueue
}

func NewTransactionManager(options ...TmOption) *TransactionManager {
	tm := &TransactionManager{}

	for _, option := range options {
		option(tm)
	}

	// 被丢弃的数据会重新从数据库获取
	tm.queue = queue.NewBoundedQueue(1000, nil)

	tm.queue.StartConsumers(200, tm.execCompensate)

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

func WithTmHttpClient(httpClient *http.Client) TmOption {
	return func(tm *TransactionManager) {
		tm.httpClient = httpClient
	}
}

func WithTmNxLock(nxLock *nxlock.Nxlock) TmOption {
	return func(tm *TransactionManager) {
		tm.nxlock = nxLock
	}
}

func (tm *TransactionManager) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():

		}
	}
}

// 获取200个随机事务组(需要补偿) 添加到事务队列中
func (tm *TransactionManager) getCompensateToQueue(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	var err error
	for {
		select {
		case <-ticker.C:
			err = tm.nxlock.Lock(ctx, SagasCompensate, 3)
			if err != nil {
				tm.logger.Infoc(ctx, "补偿任务抢锁失败 ：%s", err.Error())
			} else {
				txgroups, err := tm.txgroupRepo.GetCompensateList(context.Background(), 200)
				if err != nil {
					tm.logger.Errorf(err.Error())
				} else {
					for _, txgroup := range txgroups {
						err = tm.nxlock.Lock(ctx, fmt.Sprintf("%s:%d", SagasCompensateQueue, txgroup.Txid), 10)
						if err == nil {
							tm.queue.Produce(txgroup)
						}
					}
				}
			}
			tm.nxlock.Release(ctx, SagasCompensate)
		case ctx.Done():
			ticker.Stop()
		}
	}
}

// 执行补偿
func (tm *TransactionManager) execCompensate(item interface{}) {
	txgroup := item.(entity.Txgroup)
	tm.logger.Infof("txID %d", txgroup.Txid)

	ctx := context.Background()
	err := tm.nxlock.Lock(ctx, fmt.Sprintf("%s:%d", SagasCompensateExec, txgroup.Txid), 10)
	if err == nil {
		tm.logger.Infof("Exec compensate txID %d", txgroup.Txid)

	}

}
