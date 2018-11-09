package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	. "github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"
	"postgres/models"
	"strconv"
)

func connectSQLBoiler() (*sql.DB, error) {
	config := loadConfiguration("./config.json")
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Db,
		config.Database.Password)

	db, err := sql.Open("postgres", dbInfo)
	boil.SetDB(db)
	return db, err
}

//5.1.1
/*
	db, err := connectSQLBoiler()
	if err != nil {
		fmt.Println(err)
		return
	}

	services, err := servicesAbove(float64(500),db)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, val := range services {
		fmt.Println(val)
	}
	return
*/

func servicesAbove(n float64, db *sql.DB) (services []models.Service, err error) {

	type serviceSum struct {
		Id         int    `boil:"service_id"`
		Name       string `boil:"services.name"`
		Foreign    string `boil:"cost_foreign"`
		Our        string `boil:"cost_our"`
		models.Car `boil:",bind"`
	}
	var servicesJoined []serviceSum

	models.NewQuery(
		Select("*"),
		From("works"),
		InnerJoin("services on works.service_id = services.id"),
		InnerJoin("cars ON works.car_id = cars.id")).
		Bind(context.TODO(), db, &servicesJoined)
	serviceSums := make(map[models.Service]float64)

	for _, service := range servicesJoined {
		templateService := models.Service{ID: service.Id, Name: null.String{String: service.Name}}
		var sum float64
		if service.IsForeign.Int16 == 1 {
			sum, _ = strconv.ParseFloat(service.Foreign, 64)
		} else {
			sum, _ = strconv.ParseFloat(service.Our, 64)
		}

		if val, ok := serviceSums[templateService]; ok {
			serviceSums[templateService] = val + sum
		} else {
			serviceSums[templateService] = sum
		}
	}

	services = make([]models.Service, 0)
	for key, sum := range serviceSums {
		if sum > n {
			key.CostForeign = types.NullDecimal{Big: new(decimal.Big).SetFloat64(sum)}
			services = append(services, key)
		}
	}
	return
}
