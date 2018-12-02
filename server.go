package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"strconv"
	"time"
)


func ginServer() {
	db := connectGORMDatabase()
	autoMigrate(db)
	r := gin.Default()

	r.GET("/view/:num", func(c *gin.Context) {
		num := c.Param("num")
		switch num {
		case "6.3.2":
			sum, car := masterRecords(Master{},db)
			type answer struct {
				Car Car
				Sum float64
			}
			c.JSON(200, answer{Sum: sum, Car: car})
			break
		case "6.3.1":
			m1, _ := strconv.Atoi(c.Query("master1"))
			m2, _ := strconv.Atoi(c.Query("master2"))
			count := mastersServiceCount(Master{Id: m1},Master{Id: m2}, db)
			c.JSON(200, map[string]int{"Count":count})
			break
		case "6.2":
			cId, _ := strconv.Atoi(c.Query("car"))
			sId, _ := strconv.Atoi(c.Query("service"))
			sum, count := averageServiceForCar(Car{Id:cId},Service{Id:sId},db)
			type answer struct {
				Count int
				Sum float64
			}
			c.JSON(200, answer{Sum: sum, Count: count})
			break
		case "6.1":
			cars := averageServiceForCars(db)
			c.JSON(200, cars)
			break
		case "5.1.2":
			masters := mastersAboveGORMPreload(db)
			c.JSON(200, masters)
			break
		case "4.2.1":
			sId, _ := strconv.Atoi(c.Query("service"))
			service := Service{Id:sId}
			err := service.add10Cost(db)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, true)
			break
		case "1.3.2":
			masters := topFiveMasters(db)
			c.JSON(200, masters)
			break
		case "1.3.1":
			sum := sumCarsServicesTogether(db)
			c.JSON(200, map[string]float64{"Sum":sum})
			break
		case "1.2.2":
			type aCar struct {
				Car Car
				Sum float64
			}
			aCars := make([]aCar,0)
			cars := everyCarLastYear(db)
			for car := range cars {
				aCars = append(aCars, aCar{Car:car, Sum:cars[car]})
			}
			c.JSON(200, aCars)
			break
		case "1.2.1":
			cars := foreignCarsLastMonth(db)
			c.JSON(200, cars)
			break
		case "1.1.2":
			works := worksLastMonth(db)
			c.JSON(200, works)
			break
		case "1.1.1":
			our, foreign := sumCarsServices(db)
			c.JSON(200, map[string]float64{"Our":our, "Foreign": foreign})
			break
		}
	})

	r.GET("/list/:table", func(c *gin.Context) {
		var err error

		table := c.Param("table")
		switch table {
		case "masters":
			dbData := make([]Master, 0)
			err = db.Find(&dbData).Error
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			c.JSON(200, dbData)
			break
		case "services":
			dbData := make([]Service, 0)
			err = db.Find(&dbData).Error
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			c.JSON(200, dbData)
			break
		case "cars":
			dbData := make([]Car, 0)
			err = db.Find(&dbData).Error
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			c.JSON(200, dbData)
			break
		case "works":
			dbData := make([]Work, 0)
			err = db.Find(&dbData).Error
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			c.JSON(200, dbData)
			break
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

	})

	type Request struct {
		Action 	string 				`json:"action" binding:"required"`
		Table  	string 				`json:"table" binding:"required"`
		Id		string				`json:"id" binding:"-"`
		Update	[]struct {
			Field 	string 				`json:"field" binding:"-"`
			Value 	string 				`json:"value" binding:"-"`
		}							`json:"update" binding:"-"`
		Insert	[]struct {
			Field 	string 				`json:"field" binding:"-"`
			Value 	string 				`json:"value" binding:"-"`
		}							`json:"insert" binding:"-"`
	}

	r.PUT("/edit", func(c *gin.Context) {
		req := make([]Request,0)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		answers := make([]string, 0)
		for _, action := range req {

			id, _ := strconv.Atoi(action.Id)
			switch action.Action {
			case "UPDATE":

				var err *gorm.DB
				columns := make(map[string]interface{},0)
				for _, update := range action.Update {
					columns[update.Field] = update.Value
				}

				switch action.Table {
				case "masters":
					model := Master{Id:id}
					err = db.Model(&model).Updates(columns)
					break
				case "services":
					model := Service{Id:id}
					err = db.Model(&model).Updates(columns)
					break
				case "cars":
					model := Car{Id:id}
					err = db.Model(&model).Updates(columns)
					break
				case "works":
					model := Work{Id:id}
					err = db.Model(&model).Updates(columns)
					break
				}
				if err != nil {
					answers = append(answers, fmt.Sprint(err.Error))
				} else {
					answers = append(answers, "success")
				}
				break

			case "DELETE":

				var err *gorm.DB
				switch action.Table {
				case "masters":
					model := Master{Id:id}
					err = db.Delete(&model)
					break
				case "services":
					model := Service{Id:id}
					err = db.Delete(&model)
					break
				case "cars":
					model := Car{Id:id}
					err = db.Delete(&model)
					break
				case "works":
					model := Work{Id:id}
					err = db.Delete(&model)
					break
				}
				if err != nil {
					answers = append(answers, fmt.Sprint(err.Error))
				} else {
					answers = append(answers, "success")
				}
				break

			case "INSERT":

				var err *gorm.DB
				columns := make(map[string]interface{},0)
				for _, update := range action.Insert {
					columns[update.Field] = update.Value
				}

				switch action.Table {
				case "masters":
					model := Master{
					Name: columns["name"].(string)}
					err = db.Create(&model)
					break
				case "services":
					costOur, _ := strconv.ParseFloat(columns["cost_our"].(string), 64)
					costForeign, _ := strconv.ParseFloat(columns["cost_foreign"].(string),64)
					model := Service{
						Name: columns["name"].(string),
						CostOur: costOur,
						CostForeign: costForeign}
					err = db.Create(&model)
					break
				case "cars":
					isForeign, _ := strconv.Atoi(columns["is_foreign"].(string))
					model := Car{
						Num: columns["num"].(string),
						Color: columns["color"].(string),
						Mark: columns["mark"].(string),
						IsForeign: isForeign}
					err = db.Create(&model)
					break
				case "works":
					mId, _ := strconv.Atoi(columns["master_id"].(string))
					sId, _ := strconv.Atoi(columns["service_id"].(string))
					cId, _ := strconv.Atoi(columns["car_id"].(string))
					model := Work{
						Date: time.Now(),
						MasterId: mId,
						CarId: cId,
						ServiceId: sId}
					err = db.Create(&model)
					break
				}

				if err != nil {
					answers = append(answers, fmt.Sprint(err.Error))
				} else {
					answers = append(answers, "success")
				}
				break
			}
		}
		c.JSON(200, gin.H{"answers": answers})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
	return
}
