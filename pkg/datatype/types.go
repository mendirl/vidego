package datatype

import (
	"sync"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ConfigEntity struct {
	gorm.Model
	Name   string         `gorm:"uniqueIndex"`
	Values pq.StringArray `gorm:"type:text[]"`
}

func (ConfigEntity) TableName() string {
	return "vidego.config"
}

type VideoEntity struct {
	gorm.Model
	Name     string
	Path     string
	Size     int64
	Duration float64
	Complete bool
}

func (VideoEntity) TableName() string {
	return "vidego.video"
}

type Tabler interface {
	TableName() string
}
type CMap = struct {
	sync.RWMutex
	Value map[uint][]string
}
type CVideoEntityMap = struct {
	sync.RWMutex
	Value map[float64][]VideoEntity
}

type CSet = struct {
	sync.RWMutex
	value map[uint]void
}
type void struct{}

var member void

type CStringList = struct {
	sync.RWMutex
	Value []string
}

type CVideoEntityList = struct {
	sync.RWMutex
	Value []VideoEntity
}

type Video = struct {
	Name     string
	Path     string
	Size     int64
	Duration float64
	Complete bool
}
