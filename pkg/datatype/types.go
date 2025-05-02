package datatype

import (
	"gorm.io/gorm"
	"sync"
)

type VideoEntity struct {
	gorm.Model
	Name     string
	Path     string
	Size     int64
	Duration float64
	Complete bool
}

func (VideoEntity) TableName() string {
	return "videogo.video"
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
