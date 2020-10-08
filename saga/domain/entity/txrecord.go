package entity

import (
	"time"
)

type Txrecord struct {
	ClassName string `gorm:"column:class_name"`

	CompensateName string `gorm:"column:compensate_name"`

	CreateTime time.Time `gorm:"column:create_time"`

	GenericParamTypes string `gorm:"column:generic_param_types"`

	ID int `gorm:"column:id;primary_key"`

	IsDeleted int `gorm:"column:is_deleted;default:0"`

	Lookup string `gorm:"column:lookup"`

	MannerName string `gorm:"column:manner_name"`

	MethodName string `gorm:"column:method_name"`

	ParamTypes string `gorm:"column:param_types"`

	Params string `gorm:"column:params"`

	RegAddress string `gorm:"column:reg_address"`

	ServiceName string `gorm:"column:service_name"`

	Step int `gorm:"column:step"`

	Txid uint64 `gorm:"column:txid"`

	UpdateTime time.Time `gorm:"column:update_time"`

	Version string `gorm:"column:version"`

	// 补偿通讯方式 1 HTTP 2 GRPC 3 dubbo
	TransportType int `gorm:"column:transport_type"`

	Host string `gorm:"column:host"`

	Path string `gorm:"column:path"`
}

// delete field
func (t Txrecord) DelKey() string {
	return "is_deleted"
}

func (t Txrecord) IsEmpty() bool {
	return t.ID == 0
}
