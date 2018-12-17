package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ginServer() {
	db := connectGORMDatabase()
	autoMigrate(db)
	r := gin.Default()

	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.PUT("/auth", func(c *gin.Context) {
		login := c.Query("login")
		password := c.Query("password")
		fmt.Println(login, password)
		users := make([]User, 0)
		db.Find(&users)
		answers := make([]string, 0)
		for user := range users {
			if users[user].Login == login && users[user].Password == password{
				answers = append(answers, "success")
				c.JSON(200, gin.H{"answers": answers})
				return
			}
		}
		answers = append(answers, "failed")
		c.JSON(200, gin.H{"answers": answers})
		return
	})

	r.GET("/view/:num", func(c *gin.Context) {
		num := c.Param("num")
		switch num {
		case "6.3.2":
			sum, car := masterRecords(Master{}, db)
			type answer struct {
				Car Car
				Sum float64
			}
			c.JSON(200, answer{Sum: sum, Car: car})
			break
		case "6.3.1":
			m1, _ := strconv.Atoi(c.Query("master1"))
			m2, _ := strconv.Atoi(c.Query("master2"))
			count := mastersServiceCount(Master{Id: m1}, Master{Id: m2}, db)
			c.JSON(200, map[string]int{"Count": count})
			break
		case "6.2":
			cId, _ := strconv.Atoi(c.Query("car"))
			sId, _ := strconv.Atoi(c.Query("service"))
			sum, count := averageServiceForCar(Car{Id: cId}, Service{Id: sId}, db)
			type answer struct {
				Count int
				Sum   float64
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
			service := Service{Id: sId}
			err := service.add10Cost(db)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
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
			res := make([]map[string]float64,0)
			res = append(res, map[string]float64{"Sum": sum})
			c.JSON(200, res)
			break
		case "1.2.2":
			type aCar struct {
				Id         int
				Num        string
				Color      string
				Mark       string
				Sum float64
			}
			aCars := make([]aCar, 0)
			cars := everyCarLastYear(db)
			for car := range cars {
				aCars = append(aCars, aCar{car.Id,car.Num,car.Color,car.Mark, cars[car]})
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
			res := make([]map[string]float64,0)
			res = append(res, map[string]float64{"Our": our, "Foreign": foreign})
			c.JSON(200, res)
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
			dbData := make([]PureWork, 0)
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
		Action string `json:"action" binding:"required"`
		Table  string `json:"table" binding:"required"`
		Id     string `json:"id" binding:"-"`
		Update []struct {
			Column string `json:"column" binding:"-"`
			Value string `json:"value" binding:"-"`
		} `json:"update" binding:"-"`
	}

	r.PUT("/edit", func(c *gin.Context) {
		req := make([]Request, 0)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		answers := make([]string, 0)
		for _, action := range req {
			var err error
			id, _ := strconv.Atoi(action.Id)
			switch action.Action {
			case "UPDATE":

				columns := make(map[string]interface{}, 0)
				for _, update := range action.Update {
					columns[strings.ToLower(update.Column)] = update.Value
				}

				for column := range columns {
					if column == "cost_our" || column == "cost_foreign" {
						value, _ := strconv.ParseFloat(columns[column].(string), 64)
						columns[column] = value
					}
					if column == "is_foreign" || column == "service_id" ||
						column == "master_id" || column == "car_id"  	|| column == "id" {
						value, _ := strconv.Atoi(columns[column].(string))
						columns[column] = value
					}
				}

				switch action.Table {
				case "masters":
					model := Master{Id: id}
					err = db.Model(&model).Updates(columns).Error
					break
				case "services":
					model := Service{Id: id}
					err = db.Model(&model).Updates(columns).Error
					break
				case "cars":
					model := Car{Id: id}
					err = db.Model(&model).Updates(columns).Error
					break
				case "works":
					model := Work{Id: id}
					err = db.Model(&model).Updates(columns).Error
					break
				}
				if err != nil {
					answers = append(answers, fmt.Sprint(err))
				} else {
					answers = append(answers, "success")
				}
				break

			case "DELETE":

				switch action.Table {
				case "masters":
					model := Master{Id: id}
					err = db.Delete(&model).Error
					break
				case "services":
					model := Service{Id: id}
					err = db.Delete(&model).Error
					break
				case "cars":
					model := Car{Id: id}
					err = db.Delete(&model).Error
					break
				case "works":
					model := Work{Id: id}
					err = db.Delete(&model).Error
					break
				}
				if err != nil {
					answers = append(answers, fmt.Sprint(err))
				} else {
					answers = append(answers, "success")
				}
				break

			case "INSERT":
				columns := make(map[string]string, 0)
				for _, update := range action.Update {
					columns[strings.ToLower(update.Column)] = update.Value
				}

				switch action.Table {
				case "masters":
					model := Master{
						Id: id,
						Name: columns["name"]}
					err = db.Create(&model).Error
					break
				case "services":
					costOur, _ := strconv.ParseFloat(columns["cost_our"], 64)
					costForeign, _ := strconv.ParseFloat(columns["cost_foreign"], 64)
					model := Service{
						Id: id,
						Name:         columns["name"],
						Cost_Our:     costOur,
						Cost_Foreign: costForeign}
					err = db.Create(&model).Error
					break
				case "cars":
					isForeign, _ := strconv.Atoi(columns["is_foreign"])
					model := Car{
						Id: id,
						Num:        columns["num"],
						Color:      columns["color"],
						Mark:       columns["mark"],
						Is_Foreign: isForeign}
					err = db.Create(&model).Error
					break
				case "works":
					mId, _ := strconv.Atoi(columns["master_id"])
					sId, _ := strconv.Atoi(columns["service_id"])
					cId, _ := strconv.Atoi(columns["car_id"])
					model := Work{
						Id: id,
						Date:       time.Now(),
						Master_Id:  mId,
						Car_Id:     cId,
						Service_Id: sId}
					err = db.Create(&model).Error
					fmt.Println(err)
					break
				}

				if err != nil {
					answers = append(answers, fmt.Sprint(err))
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
