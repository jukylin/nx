package sagas

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/nx/sagas/domain/repo"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/sonyflake"
)

var (
	ErrTxIDNotFound = errors.New("获取txID失败")

	ErrTxIDIsZore = errors.New("txID为0")

	ErrEmptyTxIDStateString = errors.New("Cannot convert empty string to TxID")
)

type EsimSagas struct {
	logger log.Logger

	txgroupRepo repo.TxgroupRepo

	txrecordRepo repo.TxrecordRepo

	sf *sonyflake.Sonyflake
}

type EssOption func(*EsimSagas)

func NewEsimSagas(options ...EssOption) Sagas {
	ess := &EsimSagas{}

	for _, option := range options {
		option(ess)
	}

	ess.sf = sonyflake.NewSonyflake(sonyflake.Settings{})

	return ess
}

func WithEssLogger(logger log.Logger) EssOption {
	return func(ess *EsimSagas) {
		ess.logger = logger
	}
}

func WithEssTxgroupRepo(txgroupRepo repo.TxgroupRepo) EssOption {
	return func(ess *EsimSagas) {
		ess.txgroupRepo = txgroupRepo
	}
}

func WithEssTxrecordRepo(txrecordRepo repo.TxrecordRepo) EssOption {
	return func(ess *EsimSagas) {
		ess.txrecordRepo = txrecordRepo
	}
}

// 缺优先级.
func (ess *EsimSagas) StartTransaction(ctx context.Context) (Transaction, error) {
	var err error

	et := &esimTransaction{}
	et.logger = ess.logger
	et.txgroupRepo = ess.txgroupRepo

	tc := esimTransactionContext{}
	tc.txId, err = ess.sf.NextID()
	if err != nil {
		return nil, err
	}

	et.context = tc

	txGroup := entity.Txgroup{}
	txGroup.Txid = tc.txId
	err = ess.txgroupRepo.Create(ctx, &txGroup)
	if err != nil {
		return nil, err
	}

	et.txgroup = txGroup

	return et, nil
}

// 创建saga，如果txID为0 返回noopSaga
func (ess *EsimSagas) CreateSaga(ctx context.Context, txID uint64) (Saga, error) {
	ess.logger.Debugc(ctx, "创建 saga txID: %d", txID)
	if txID == 0 {
		return &noopSaga{}, ErrTxIDIsZore
	}

	es := &esimSaga{}
	es.logger = ess.logger
	es.txrecordRepo = ess.txrecordRepo
	es.txID = txID

	return es, nil
}

func (ess *EsimSagas) Inject(ctx context.Context, format interface{}, abstractCarrier interface{}) error {
	textMapWriter, ok := abstractCarrier.(opentracing.TextMapWriter)
	if !ok {
		return opentracing.ErrInvalidCarrier
	}

	txID := TxIDFromContext(ctx)
	if txID != 0 {
		textMapWriter.Set(TranContextHeaderName, strconv.FormatUint(txID, 10))
		return nil
	}

	return ErrTxIDNotFound
}

// 从 abstractCarrier 提取 TransactionContext
func (ess *EsimSagas) Extract(ctx context.Context, format interface{}, abstractCarrier interface{}) (TransactionContext, error) {
	var tc esimTransactionContext
	var err error
	var txID uint64

	textMapReader, ok := abstractCarrier.(opentracing.TextMapReader)
	if !ok {
		return tc, opentracing.ErrInvalidCarrier
	}

	textMapReader.ForeachKey(func(key, val string) error {
		key = strings.ToLower(key)
		if key == TranContextHeaderName {
			if val == "" {
				err = ErrEmptyTxIDStateString
				return err
			}
			ess.logger.Debugc(ctx, "Extract txID: %s", val)
			txID, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}
			tc.txId = txID
		}

		return nil
	})

	return tc, err
}
