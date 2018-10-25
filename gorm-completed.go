package main

import "github.com/jinzhu/gorm"

// 4.2.1

func (s *Service) add10Cost (db *gorm.DB) error {
	db.Model(s).
		Where("id = ?",
			db.Table("works").
				Select("service_id").
				Order("date_work DESC").
				Limit(1).
				SubQuery()).
		Updates(map[string]interface{}{
			"cost_our": gorm.Expr("cost_our + ?", 10),
			"cost_foreign": gorm.Expr("cost_foreign + ?", 10) })
	return nil
}
