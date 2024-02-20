package dao

import "github.com/simlecode/subspace-tool/models"

func (d *Dao) SaveSpace(s *models.Space) error {
	return d.db.Save(s).Error
}

func (d *Dao) ListSapce() ([]models.Space, error) {
	var s []models.Space
	err := d.db.Find(&s).Error
	return s, err
}
