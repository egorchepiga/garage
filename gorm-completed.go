package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"math"
	"sort"
	"time"
)

// 1.1.1
func sumCarsServices(db *gorm.DB) (our, foreign float64) {
	works := make([]Work, 0)
	db.Set("gorm:auto_preload", true).
		Find(&works)
	for _, work := range works {
		if work.Car.Is_Foreign == 1 {
			foreign += work.Service.Cost_Foreign
		} else {
			our += work.Service.Cost_Our
		}
	}
	return
}

// 1.1.2
func worksLastMonth(db *gorm.DB) (works []Work) {
	db.Where("date_work > CURRENT_TIMESTAMP - interval '30 days'").
		Where("date_work < CURRENT_TIMESTAMP").
		Find(&works)
	return
}

// 1.2.1
func foreignCarsLastMonth(db *gorm.DB) (cars []Car) {
	works := make([]Work,0)
	db.Preload("Car", "is_foreign = ?", 1).
		Where("date_work > CURRENT_TIMESTAMP - interval '30 days'").
		Where("date_work < CURRENT_TIMESTAMP").
		Find(&works)

	carsCounter:= make(map[int]int,0)
	carsId := make([]int,0)
	for _, work := range works {
		carsCounter[work.Car.Id] = 0
	}
	for _, work := range works {
			carsCounter[work.Car.Id]++
	}
	for id, count := range carsCounter {
		if count > 1 {
			carsId = append(carsId, id)
		}
	}
	db.Where("id in (?)", carsId).Find(&cars)
	return
}

// 1.2.2
func everyCarLastYear(db *gorm.DB) (cars map[Car]float64) {
	cars = make(map[Car]float64)

	allCars := make([]Car, 0)
	db.Find(&allCars)

	for i := 0; i < len(allCars); i++ {
		cars[allCars[i]] = 0
	}

	works :=  make([]Work, 0)

	db.Set("gorm:auto_preload", true).
		Find(&works)

	for _, work := range works {
		var cost float64
		if work.Car.Is_Foreign == 1 {
			cost = work.Service.Cost_Foreign
		} else {
			cost = work.Service.Cost_Our
		}
		work.Car.Id = work.Master_Id
		cars[work.Car] += cost
	}

	return
}

// 1.3.1
func sumCarsServicesTogether(db *gorm.DB) float64 {
	our, foreign := sumCarsServices(db)
	return our + foreign
}

// 1.3.2
func topFiveMasters(db *gorm.DB) (masters []Master) {
	type TopMaster struct {
		Master Master
		fixedCars map[int]int
	}

	topMasters :=  make(map[int]TopMaster, 0)
	works :=  make([]Work, 0)
	db.Preload("Master").
		Where("date_work > CURRENT_TIMESTAMP - interval '30 days'").
		Where("date_work < CURRENT_TIMESTAMP").
		Find(&works)
	for _, work := range works {
		topMasters[work.Master.Id] = TopMaster{work.Master, make(map[int]int,0)}
	}
	for _, work := range works {
		topMasters[work.Master.Id].fixedCars[work.Car_Id] = 0
	}

	arrTopMasters := make([]TopMaster,0)
	for _, master := range topMasters {
		arrTopMasters = append(arrTopMasters, master)
	}
	sort.Slice(arrTopMasters, func(curr, next int) bool {
		return len(arrTopMasters[curr].fixedCars) > len(arrTopMasters[next].fixedCars)
	})
	for i := 0; i<5 && i< len(arrTopMasters); i++ {
		masters = append(masters, arrTopMasters[i].Master)
	}
	return
}


//2.1.2


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


// 5.1.1
func servicesAboveGORM(n float64, db *gorm.DB) (services []Service, err error) {
	type joinedWorkService struct {
		Service
		Car
		Sum 	string `gorm:"column:sum1;"`
	}

	result := make([]joinedWorkService,0)

	subQuery := db.Table("works").
		Select("*, ? as sum1",
			gorm.Expr("CASE WHEN is_foreign = ? THEN cost_our ELSE cost_foreign END",0)).
		Joins("INNER JOIN services ON works.service_id = services.id").
		Joins("INNER JOIN cars ON works.car_id = cars.id").
		SubQuery()

	subQuery2 := db.Table("services").
		Select("services.id, SUM(tab1.sum1) as sum1").
		Joins("INNER JOIN ? as tab1 ON tab1.service_id = services.id", subQuery).
		Group("services.id").
		Having("SUM(tab1.sum1) > ?", n).
		SubQuery()

	db.Table("services").
		Select("*").
		Joins("INNER JOIN ? as tab2 ON tab2.id = services.id", subQuery2).
		Scan(&result)

	for _, val := range result {
		fmt.Println(val)
	}
	return
}

//5.1.2
func servicesAboveGORMPreload(n float64, db *gorm.DB) (services map[Service]float64) {
	works :=  make([]Work, 0)
	db.
		Preload("Service").
		Preload("Car").
		Preload("Master").
		Find(&works)

	services = make(map[Service]float64)

	for _, work := range works {
		var cost float64
		if work.Car.Is_Foreign == 1 {
			cost = work.Service.Cost_Foreign
		} else {
			cost = work.Service.Cost_Our
		}
		work.Service.Id = work.Service_Id
		if _, ok := services[work.Service]; ok {
			services[work.Service] += cost
		} else {
			services[work.Service] = cost
		}
	}

	for service, sum := range services {
		if sum < n {
			delete(services,service)
		}
	}
	return
}


//5.2
func mastersAboveGORMPreload(db *gorm.DB) (masters map[Master]float64) {
	masters = make(map[Master]float64)

	allMasters := make([]Master, 0)
	db.Find(&allMasters)

	for i := 0; i < len(allMasters); i++ {
		masters[allMasters[i]] = 0
	}

	works :=  make([]Work, 0)

	db.Set("gorm:auto_preload", true).
		Find(&works)

	for _, work := range works {
		var cost float64
		if work.Car.Is_Foreign == 1 {
			cost = work.Service.Cost_Foreign
		} else {
			cost = work.Service.Cost_Our
		}
		work.Master.Id = work.Master_Id
		masters[work.Master] += cost
	}

	return
}

//6.1
func averageServiceForCars(db *gorm.DB) (cars map[Car]float64) {
	cars = make(map[Car]float64)

	allCars := make([]Car, 0)
	db.Find(&allCars)

	for i := 0; i < len(allCars); i++ {
		cars[allCars[i]] = 0
	}

	works :=  make([]Work, 0)

	db.Set("gorm:auto_preload", true).
		Find(&works)

	for _, work := range works {
		var cost float64
		if work.Car.Is_Foreign == 1 {
			cost = work.Service.Cost_Foreign
		} else {
			cost = work.Service.Cost_Our
		}
		work.Car.Id = work.Master_Id
		cars[work.Car] += cost
	}

	return
}

//6.2
func averageServiceForCar(car Car, service Service, db *gorm.DB) (sum float64, count int) {
	works :=  make([]Work, 0)

	db.Set("gorm:auto_preload", true).
		Where("car_id = ?", car.Id).
		Where("service_id = ?", service.Id).
		Find(&works)

	for _, work := range works {
		if work.Car.Is_Foreign == 1 {
			sum += work.Service.Cost_Foreign
		} else {
			sum += work.Service.Cost_Our
		}
	}
	count = len(works)
	return
}

//6.3.1
func mastersServiceCount(m1,m2 Master, db *gorm.DB) (count int) {
	works :=  make([]Work, 0)

	fmt.Println(m1,m2)
	db.Set("gorm:auto_preload", true).
		Where("master_id in (?)", []int{m1.Id, m2.Id}).
		Find(&works)

	count = len(works)
	return
}

//6.3.2
func masterRecords(master Master, db *gorm.DB) (sum float64, car Car) {
	works :=  make([]Work, 0)

	db.Set("gorm:auto_preload", true).
		Find(&works)

	for _, work := range works {
		var cost float64
		if work.Car.Is_Foreign == 1 {
			cost = work.Service.Cost_Foreign
		} else {
			cost = work.Service.Cost_Our
		}
		if cost > sum {
			sum = cost
		}
		car = work.Car
	}
	//fmt.Println(sum, car)

	return
}

//7.1.1
func (car *Car) BeforeCreate(scope *gorm.Scope) (err error) {
	if !scope.DB().
		Where("num = ?", car.Num).
		First(&car).
		RecordNotFound() {
		err = fmt.Errorf("car NUM already registered")
	}
	return
}

//7.1.2
func (master *Master) BeforeCreate(scope *gorm.Scope) (err error) {
	masters := make([]Master, 0)
	scope.DB().Find(&masters)

	if len(masters) >= 10 {
		err = fmt.Errorf("we have %d masters with limit of %d", len(masters), 10)
	}
	return
}

//7.1.3
func (work *Work) BeforeSave(scope *gorm.Scope) (err error) {
	works :=  make([]Work, 0)
	scope.DB().
		Where("date_work > CURRENT_TIMESTAMP - interval '1 day'").
		Where("master_id = ?", work.Master.Id).
		First(&works)
	if len(works) > 1 {
		err = fmt.Errorf("master has at least 2 works")
	}
	return
}

//7.2.1
func (master *Master) BeforeUpdate(scope *gorm.Scope) (err error) {
	name := master.Name
	scope.DB().Find(master)
	if name != master.Name {
		err = fmt.Errorf("you cant change master's name")
	}
	return
}

//7.2.2
/*func (service *Service) BeforeUpdate(scope *gorm.Scope) (err error) {
	cost_our_new := service.Cost_Our
	scope.DB().Find(service)
	if cost_our_new != 0 && cost_our_new != service.Cost_Our {
		cost_foreign_new := service.Cost_Foreign + service.Cost_Foreign* (cost_our_new - service.Cost_Our) / service.Cost_Our
		scope.DB().Model(service).Update("cost_foreign", cost_foreign_new)
	}
	return
}*/

//7.2.3
func (work *Work) BeforeUpdate(scope *gorm.Scope) (err error) {
	date := work.Date
	date = time.Date(
		date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second(),0,&time.Location{})

	scope.DB().Find(work)

	diff := work.Date.UTC().Sub(date)
	hours := math.Abs(float64(diff.Hours()))

	if hours >= 24 {
		err = fmt.Errorf("cant change date more than 24 hours (1 day). diff = %f hours",hours)
	}
	return
}

//7.3.1
func (car *Car) BeforeDelete(scope *gorm.Scope) (err error) {
	work := new(Work)
	if !scope.DB().
		Where("car_id = ?", car.Id).
		First(work).
		RecordNotFound() {
		err = fmt.Errorf("this car has works")
	}
	return
}

//7.3.2
func (work *Work) BeforeDelete(scope *gorm.Scope) (err error) {

	if scope.DB().
		Where("car_id = ?", work.Car.Id).
		First(work).
		RecordNotFound() {
		scope.DB().Delete(&work.Car)
	}
	if scope.DB().
		Where("service_id = ?", work.Service.Id).
		First(work).
		RecordNotFound() {
		scope.DB().Delete(&work.Service)
	}
	if scope.DB().
		Where("master_id = ?", work.Master.Id).
		First(work).
		RecordNotFound() {
		scope.DB().Delete(&work.Master)
	}
	return
}

//7.3.3
/*func (service *Service) BeforeDelete(scope *gorm.Scope) (err error) {
	works := make([]Work,0)
	worksId := make([]int,0)
	scope.DB().
		Where("service_id = ?", service.Id).
		Find(&works)
	if len(works) > 2 {
		err = fmt.Errorf("you cant delete service with %d works", len(works))
	} else {
		for _, work := range works {
			worksId = append(worksId, work.Id)
		}
		scope.DB().Where("id in (?)", worksId).Delete(&works)
	}
	return
}
*/
