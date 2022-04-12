package data

import "gorm.io/gorm"

type TagsDAO struct {
	db *gorm.DB
}

func NewTagsDAO(db *gorm.DB) *TagsDAO {
	return &TagsDAO{db}
}

func (d *TagsDAO) GetAll() ([]string, error) {
	tags := make([]Tag, 0)
	err := d.db.Find(&tags).Error

	values := make([]string, len(tags))
	for i := range tags {
		values[i] = tags[i].Name
	}

	return values, err
}
