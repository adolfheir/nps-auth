package sql

import (
	"time"

	"gorm.io/gorm"
)

type Channel struct {
	*ChannelEntity
}

type ChannelEntity struct {
	ChannelId int    `json:"channelId" gorm:"primary_key;auto_increment"`
	Desc      string `json:"desc"` // 备注

	NpsHost       string `json:"npsHost"`
	NpsClientId   string `json:"npsClientId"`
	NpsClientKey  string `json:"npsClientKey"`
	NpsTunnelId   string `json:"npsTunnelId"`
	NpsTunnelPort int    `json:"npsTunnelPort"`

	MachineId   string    `json:"machineId"`
	ExpiredTime time.Time `json:"expiredTime"`

	Cert string `json:"cert"`
	Csr  string `json:"csr"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
