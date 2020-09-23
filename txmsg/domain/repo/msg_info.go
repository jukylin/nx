package repo

import (
	"context"
	"study-go/txmsg/domain/entity"
	"gorm.io/gorm"
	"github.com/jukylin/esim/log"
	"study-go/txmsg/domain/dao"
	"time"
)

type MsgInfoRepo interface {
	FindById(ctx context.Context, id int64) entity.MsgInfo

	CreateByTx(ctx context.Context, tx *gorm.DB, msgInfo *entity.MsgInfo) int

	Create(ctx context.Context, msgInfo *entity.MsgInfo) int

	UpdateByIds(ctx context.Context, tx *gorm.DB, update map[string]interface{}, ids []int64) int64

	UpdateStatusById(ctx context.Context, id int64) int64

	DelSendedMsg(ctx context.Context, num int) int64

	GetWaitingMsg(ctx context.Context, num int) entity.MsgInfos
}

type msgInfoRepo struct {
	logger log.Logger

	msgInfoDao *dao.MsgInfoDao
}

type MsgInfoOption func(c *msgInfoRepo)

func NewDBMsgInfoRepo(options ...MsgInfoOption) MsgInfoRepo {
	repo := &msgInfoRepo{}

	for _, option := range options {
		option(repo)
	}

	if repo.msgInfoDao == nil {
		repo.msgInfoDao = dao.NewMsgInfoDao()
	}

	return repo
}

func WithMsgInfoLogger(logger log.Logger) MsgInfoOption {
	return func(mir *msgInfoRepo) {
		mir.logger = logger
	}
}

func (mir *msgInfoRepo) FindById(ctx context.Context, id int64) entity.MsgInfo {
	var msgInfo entity.MsgInfo
	var err error

	msgInfo, err = mir.msgInfoDao.Find(ctx, "*", "id = ? and status = 0", id)
	if err != nil {
		mir.logger.Errorc(ctx, err.Error())
	}

	return msgInfo
}

func (mir *msgInfoRepo) CreateByTx(ctx context.Context, tx *gorm.DB, msgInfo *entity.MsgInfo) int {
	insertId, err := mir.msgInfoDao.Create(ctx, msgInfo)
	if err != nil {
		mir.logger.Errorc(ctx, err.Error())
	}

	return insertId
}

func (mir *msgInfoRepo) Create(ctx context.Context, msgInfo *entity.MsgInfo) int {
	insertId, err := mir.msgInfoDao.Create(ctx, msgInfo)
	if err != nil {
		mir.logger.Errorc(ctx, err.Error())
	}

	return insertId
}

func (mir *msgInfoRepo) UpdateByIds(ctx context.Context, tx *gorm.DB, update map[string]interface{}, ids []int64) int64 {
	insertId, err := mir.msgInfoDao.UpdateWithTx(ctx, tx, update, ids)
	if err != nil {
		mir.logger.Errorc(ctx, err.Error())
	}

	return insertId
}

func (mir *msgInfoRepo) UpdateStatusById(ctx context.Context, id int64) int64 {
	insertId, err := mir.msgInfoDao.Update(ctx, map[string]interface{}{"status":1}, id)
	if err != nil {
		mir.logger.Errorc(ctx, err.Error())
	}

	return insertId
}

// 删除前三天的数据
func (mir *msgInfoRepo) DelSendedMsg(ctx context.Context, num int) int64 {
	insertId, err := mir.msgInfoDao.Delete(ctx, num, "create_time < ? and status = 1", time.Now().Add(- 3 * time.Minute * 60 * 24))
	if err != nil {
		mir.logger.Errorc(ctx, err.Error())
	}

	return insertId
}

func (mir *msgInfoRepo) GetWaitingMsg(ctx context.Context, num int) entity.MsgInfos {
	msgInfos, err := mir.msgInfoDao.List(ctx, num, "id, content, topic, tag", "status = 0")
	if err != nil {
		mir.logger.Errorc(ctx, err.Error())
	}

	return msgInfos
}
