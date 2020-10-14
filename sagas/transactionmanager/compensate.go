package transactionmanager

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/repo"
	"github.com/jukylin/esim/http"
	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/nx/sagas/domain/value-object"
	"fmt"
	"github.com/jukylin/esim/mysql"
)

var (
	ErrDataIsEmpty = "数据为空 %+v"

	ErrHostIsEmpty = "Host为空"

	ErrParamsIsEmpty = "Params为空"

	ErrHTTPStatus = "http响应状态码错误 %d"

	ErrStateMustStart = "状态必须为：开始 %d : %d"

	ErrStateMustCompensate = "状态必须为：补偿 %d : %d"

	ErrUnSupportTranSportType = "不支持的通讯协议 : %d"
)

type Compensate interface {
	ExeCompensate(ctx context.Context, txgroup entity.Txgroup) error

	// 修改事物状态 0 = 》 2
	BuildCompensate(ctx context.Context, txgroup entity.Txgroup) error

	CompensateRecord(ctx context.Context, txcompensate entity.Txcompensate) error

	CompensateHook(ctx context.Context, txgroup entity.Txgroup) error
}

type BcOption func(*backwardCompensate)

// 逆向补偿
type backwardCompensate struct {
	logger log.Logger

	txgroupRepo repo.TxgroupRepo

	txrecordRepo repo.TxrecordRepo

	txcompensateRepo repo.TxcompensateRepo

	httpClient *http.Client

	mysqlClient *mysql.Client

	tf *TransportFactory
}

func NewBackwardCompensate(options ...BcOption) Compensate {
	bc := &backwardCompensate{}
	for _, option := range options {
		option(bc)
	}

	return bc
}


func WithBcTxgroupRepo(txgroupRepo repo.TxgroupRepo) BcOption {
	return func(bc *backwardCompensate) {
		bc.txgroupRepo = txgroupRepo
	}
}

func WithBcTxrecordRepo(txrecordRepo repo.TxrecordRepo) BcOption {
	return func(bc *backwardCompensate) {
		bc.txrecordRepo = txrecordRepo
	}
}

func WithBcTxcompensateRepo(txcompensateRepo repo.TxcompensateRepo) BcOption {
	return func(bc *backwardCompensate) {
		bc.txcompensateRepo = txcompensateRepo
	}
}

func WithBcMysqlClient(mysqlClient *mysql.Client) BcOption {
	return func(bc *backwardCompensate) {
		bc.mysqlClient = mysqlClient
	}
}

func WithBcLogger(logger log.Logger) BcOption {
	return func(bc *backwardCompensate) {
		bc.logger = logger
	}
}

// 获取事物组里需要补偿的事物，进行补偿
func (bc *backwardCompensate) ExeCompensate(ctx context.Context, txgroup entity.Txgroup) error {
	if txgroup.IsEmpty() {
		return fmt.Errorf(ErrDataIsEmpty, txgroup)
	}

	if txgroup.State != value_object.TranCompensate {
		return fmt.Errorf(ErrStateMustCompensate, txgroup.Txid, txgroup.State)
	}

	txcompensates := bc.txcompensateRepo.GetCompensateListByTxID(ctx, txgroup.Txid)
	for _, txcompensate := range txcompensates {
		err := bc.CompensateRecord(ctx, txcompensate)
		if err != nil {
			bc.logger.Errorc(ctx, err.Error())
		}
	}

	err := bc.CompensateHook(ctx, txgroup)
	if err != nil {
		bc.logger.Errorc(ctx, err.Error())
	}

	return nil
}

// 生产补偿状态数据，和修改事物组状态
func (bc *backwardCompensate) BuildCompensate(ctx context.Context, txgroup entity.Txgroup) error {
	bc.logger.Infoc(ctx, "processorBuildCompensate txID %d actionId %d", txgroup.Txid, txgroup.ID)

	var err error
	if txgroup.State != value_object.TranStart {
		return  fmt.Errorf(ErrStateMustStart, txgroup.Txid, txgroup.State)
	}

	tx := bc.mysqlClient.GetCtxDb(ctx, "sagas").Begin()
	if tx.Error != nil {
		return tx.Error
	}
	err = bc.txcompensateRepo.InsertUpdateFromRecord(ctx, tx, txgroup.Txid)
	if err != nil {
		tx.Rollback()
		bc.logger.Errorc(ctx, tx.Error.Error())
		return err
	}

	err = bc.txgroupRepo.SetStateWithTx(ctx, tx, value_object.TranCompensate, txgroup.Txid)
	if err != nil {
		tx.Rollback()
		bc.logger.Errorc(ctx, tx.Error.Error())
		return err
	}

	tx.Commit()
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

// 补偿服务
func (bc *backwardCompensate) CompensateRecord(ctx context.Context, txcompensate entity.Txcompensate) error {
	bc.logger.Infoc(ctx, "processorCompensateRecord txID %d actionId %d", txcompensate.Txid, txcompensate.ID)

	if txcompensate.IsEmpty() {
		return fmt.Errorf(ErrDataIsEmpty, txcompensate)
	}

	txrecord := bc.txrecordRepo.FindById(ctx, int64(txcompensate.ID))
	if txrecord.IsEmpty() {
		return fmt.Errorf(ErrDataIsEmpty, txrecord)
	}

	ts, err := bc.tf.GetTransport(txrecord.TransportType)
	if err != nil {
		bc.logger.Errorc(ctx, "GetTransport %s", err.Error())
	}

	err = ts.Invoke(ctx, txrecord)
	if err != nil {
		bc.logger.Errorc(ctx, "GetTransport %s", err.Error())
	} else {
		// 状态更新失败，也作为成功返回
		// 重新补偿业务服务，需要业务保证幂等性
		err = bc.txcompensateRepo.CompensateSuccess(ctx, txcompensate.ID)
		if err != nil {
			bc.logger.Errorc(ctx, err.Error())
		}
	}

	return nil
}

// 完全补偿完毕，修改事务组状态
func (bc *backwardCompensate) CompensateHook(ctx context.Context, txgroup entity.Txgroup) error {
	bc.logger.Infoc(ctx, "CompensateHook txID %d", txgroup.Txid)

	have, err := bc.txcompensateRepo.StillHaveUnfinshedCompensationInTransactionGroup(ctx, txgroup.Txid)
	if err != nil {
		return err
	}

	if !have {
		err = bc.txgroupRepo.SetStateBytxID(ctx, value_object.TranCompensateFinish, txgroup.Txid)
		if err != nil {
			return err
		}

		bc.logger.Infoc(ctx, "compensateFinishTransaction 事物补偿完成 txID : %d", txgroup.Txid)
	} else {
		bc.logger.Infoc(ctx, "CompensateHook 存在未补偿完事物 txID : %d", txgroup.Txid)
	}

	return nil
}