package transactionmanager

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/saga/domain/repo"
	"github.com/jukylin/esim/http"
	"github.com/jukylin/nx/saga/domain/entity"
	"github.com/jukylin/nx/saga/domain/value-object"
	"fmt"
	"github.com/jukylin/esim/mysql"
)

var (
	ErrDataIsEmpty = "数据为空 %+v"

	ErrHostIsEmpty = "Host为空"

	ErrParamsIsEmpty = "Params为空"

	ErrParamsMustJson = "Params必须为json"

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

	CompensateHook(ctx context.Context)
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

func (bc *backwardCompensate) ExeCompensate(ctx context.Context, txgroup entity.Txgroup) error {
	if txgroup.IsEmpty() {
		return fmt.Errorf(ErrDataIsEmpty, txgroup)
	}

	if txgroup.State != value_object.TranCompensate {
		return fmt.Errorf(ErrStateMustCompensate, txgroup.Txid, txgroup.State)
	}

	txcompensates := bc.txcompensateRepo.ListByTxID(ctx, txgroup.Txid)
	for _, txcompensate := range txcompensates {
		err := bc.CompensateRecord(ctx, txcompensate)
		if err != nil {
			bc.logger.Errorc(ctx, err.Error())
		}
	}

	return nil
}

func (bc *backwardCompensate) BuildCompensate(ctx context.Context, txgroup entity.Txgroup) error {
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

	err = bc.txgroupRepo.SetStateBytxID(ctx, tx, value_object.TranCompensate, txgroup.Txid)
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

func (bc *backwardCompensate) CompensateRecord(ctx context.Context, txcompensate entity.Txcompensate) error {
	bc.logger.Infoc(ctx, "processorCompensateRecord txID %d actionId %d", txcompensate.Txid, txcompensate.ID)

	if txcompensate.IsEmpty() {
		return fmt.Errorf(ErrDataIsEmpty, txcompensate)
	}

	txrecord := bc.txrecordRepo.FindById(ctx, int64(txcompensate.ID))
	if txrecord.IsEmpty() {
		return fmt.Errorf(ErrDataIsEmpty, txrecord)
	}

	if txrecord.TransportType == value_object.TranSportHTTP {
	} else {
		return fmt.Errorf(ErrUnSupportTranSportType, txrecord.TransportType)
	}
	
	

	return nil
}

func (bc *backwardCompensate) CompensateHook(ctx context.Context) {

}