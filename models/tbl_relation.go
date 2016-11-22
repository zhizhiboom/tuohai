package models

import (
	"tuohai/internal/convert"
)

type Relation struct {
	Id        int    `gorm:"column:id"`
	Rid       string `gorm:"column:rid"`
	SmallId   string `gorm:"column:small_id"`
	BigId     string `gorm:"column:big_id"`
	OriginId  string `gorm:"column:origin_id"`
	Status    int    `gorm:"column:status"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpatedAt  int64  `gorm:"column:upated_at"`
}

func (r *Relation) TableName() string {
	return "tbl_relation"
}

func Friends(uuid string) ([]Relation, error) {
	var r []Relation
	err := db.Raw("SELECT * FROM newim.tbl_relation where status = 0 and (small_id = ? or big_id = ?)", uuid, uuid).Find(&r).Error
	return r, err
}

func Friend(token, fuuid string) (*Relation, error) {
	var rel Relation
	small, big := "", ""
	if convert.RuneAccumulation(token) > convert.RuneAccumulation(fuuid) {
		small, big = fuuid, token
	} else {
		small, big = token, fuuid
	}

	err := db.Find(&rel, "status = 0 and small_id = ? and big_id = ?", small, big).Error
	return &rel, err
}