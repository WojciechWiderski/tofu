package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	CurrentExp   uint
	CurrentLevel uint
}

type Level struct {
	gorm.Model
	MinExperience uint
	MaxExperience uint
	Award         string
}

type Day struct {
	gorm.Model
	Name   string
	TaskID uint
}

type Date struct {
	gorm.Model
	Value  uint64
	TaskID uint
}

type Task struct {
	gorm.Model
	IsRepeatable           bool
	DaysOfTheWeek          []Day
	Dates                  []Date
	Deadline               uint64
	Name                   string
	Description            string
	DefaultExperience      uint
	CurrentExperience      uint
	UpgradeExperienceValue uint
	Record                 uint
	Status                 uint
}
