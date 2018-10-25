package main

import (
	"fmt"
	"github.com/go-xorm/xorm"
)

//1.1.1
func (s Services) Sum(db *xorm.Engine) [3]int {
	sums, _ := db.Sums(s, "cost_our", "cost_foreign")
	count, _ := db.Count(s)
	return [3]int{int(count), int(sums[0]), int(sums[1])}
}

//1.1.2
func (w Works) lastMonth(db *xorm.Engine) []Works {
	var works []Works
	db.
		Where("date_work > CURRENT_TIMESTAMP - interval '30 days'").
		And("date_work < CURRENT_TIMESTAMP").
		Find(&works)
	return works
}

//1.2.1
type CarsCounted struct {
	Cars  `xorm:"extends"`
	Count int32 `xorm:"'count'"`
}

func (CarsCounted) TableName() string {
	return "works"
}

func (CarsCounted) twiceLastMonth(db *xorm.Engine) []CarsCounted {
	var works []CarsCounted
	JoinedWork{}.
		get(db).
		Select("cars.id, date_work, is_foreign, num, color, mark, count(*)").
		Where("date_work > CURRENT_TIMESTAMP - interval '30 days'").
		And("date_work < CURRENT_TIMESTAMP").
		And("is_foreign = 1").
		GroupBy("cars.id, date_work, is_foreign, num, color, mark").
		Find(&works)

	res := make([]CarsCounted, 0, 2)
	for _, row := range works {
		if row.Count > 2 {
			res = append(res, row)
		}
	}
	return res
}

func (j CarsCounted) String() string {
	return fmt.Sprintf("Cars Counted < \n"+
		"\t Car: %d num: %s color: %s mark: %s foreign: %d \n"+
		"\t count : %d> \n",
		j.Cars.Id, j.Cars.Num, j.Cars.Color, j.Cars.Mark, j.Cars.IsForeign, j.Count)
}


// 1.2.2

type CarsCountedCost struct {
	Cars  		`xorm:"extends"`
	Sum float32 `xorm:"'sum'"`
}

func (CarsCountedCost) TableName() string {
	return "works"
}

func (CarsCountedCost) getSum(db *xorm.Engine) []CarsCountedCost {
	var works []CarsCountedCost
	JoinedWork{}.
		get(db).
		Select("cars.id, is_foreign, num, color, mark," +
			"sum(CASE \n" +
				"WHEN is_foreign = 0 THEN cost_our \n" +
				"ELSE cost_foreign \n" +
				"END)").
		Where("date_work > CURRENT_TIMESTAMP - interval '365 days'").
		And("date_work < CURRENT_TIMESTAMP").
		GroupBy("cars.id, is_foreign, num, color, mark").
		Desc("sum").
		Find(&works)

	return works
}

func (j CarsCountedCost) String() string {
	return fmt.Sprintf("Cars Counted < \n"+
		"\t Car: %d num: %s color: %s mark: %s foreign: %d \n"+
		"\t sum : %f> \n",
		j.Cars.Id, j.Cars.Num, j.Cars.Color, j.Cars.Mark, j.Cars.IsForeign, j.Sum)
}


// 1.3.1

type SumWorks struct {
	Sum 		float32 `xorm:"'sum_our'"`
	SumForeign 	float32 `xorm:"'sum_foreign'"`
	SumAll 		float32 `xorm:"'sum_all'"`
}

func (SumWorks) TableName() string {
	return "works"
}

func (SumWorks) getSum(db *xorm.Engine) []SumWorks {
	var works []SumWorks
	JoinedWork{}.
		get(db).
		Select("SUM(CASE \n" +
				"WHEN is_foreign = 0 THEN cost_our \n" +
				"ELSE 0 \n" +
				"END) AS sum_our, \n" +
			"SUM(CASE \n" +
				"WHEN is_foreign = 1 THEN cost_foreign \n" +
				"ELSE 0 \n" +
				"END) AS sum_foreign, \n" +
			"SUM(CASE \n" +
				"WHEN is_foreign = 0 THEN cost_our \n" +
				"ELSE cost_foreign \n" +
				"END) AS sum_all \n").
		Find(&works)

	return works
}

func (sw SumWorks) String() string {
	return fmt.Sprintf("Cars Counted < \n" +
		"\t Наши авто : %f Заграничные : %f Общее : %f> \n",
		sw.Sum, sw.SumForeign, sw.SumAll)
}

// 1.3.2

type TopMasters struct {
	Masters 		`xorm:"extends"`
	Count 	int32 	`xorm:"'count'"`
}

func (TopMasters) TableName() string {
	return "works"
}

func (tm TopMasters) String() string {
	return fmt.Sprintf("Master <%d %s сделано: %d>\n", tm.Id, tm.Name, tm.Count)
}

func (TopMasters) get (db *xorm.Engine) []TopMasters {
	var masters []TopMasters
	JoinedWork{}.get(db).
		Select("masters.id, masters.name, count(*)").
		Where("date_work > CURRENT_TIMESTAMP - interval '30 days'").
		And("date_work < CURRENT_TIMESTAMP").
		GroupBy("masters.id, masters.name").
		Desc("count").
		Limit(5).
		Find(&masters)
	return masters
}

// 2.1.1
/*
	service := Services{Name: "go test", CostForeign:200, CostOur:100}
	service.add(db)
	fmt.Println(service)
*/

func (s *Services) add (db *xorm.Engine) *Services {
	_, err := db.Insert(s)
	if err != nil {
		fmt.Println(err)
	}
	return s
}

// 2.1.2
/*
	work := Works{Car:27}
	work.add("go test", db)
	fmt.Println(work)
*/

func (w *Works) add (db *xorm.Engine, serviceName string) *Works {
	service := Services{Name: serviceName}
	db.Get(&service)
	w.Service = service.Id
	db.Insert(w)
	return w
}

// 2.2.1

/*
	car := Cars{ Num:"770", Color:"red", Mark:"Toyota", IsForeign:1}
	service := Services{Name:"Go Service", CostOur:20.20, CostForeign:23.14}
	work, err := transactionCarService(car, service, db)
	fmt.Println(work, err)
*/
func transactionCarService(db *xorm.Engine, car Cars, service Services) (errWork Works, err error) {
	session := db.NewSession()
	errWork = Works{}
	defer session.Close()

	err = session.Begin()
	if err != nil {
		return
	}

	_, err = session.Insert(&car)
	if err != nil {
		session.Rollback()
		return
	}

	_, err = session.Insert(&service)
	if err != nil {
		session.Rollback()
		return
	}

	work := Works{Car:car.Id, Service:service.Id}
	_, err = session.Insert(&work)
	if err != nil {
		session.Rollback()
		return
	}
	session.Commit()
	return work, err
}

// 2.2.2
/*
	master := Masters{Name:"Дэн"}
	car := Cars{Num:"777"}
	service := Services{Name:"косметика"}

	db.Get(&master)
	db.Get(&car)
	db.Get(&service)
	work := Works{ Master:master.Id, Car:car.Id, Service:service.Id}

	res, err := master.addWork(db, work)
	if err != nil {
	panic(err)
	}

	fmt.Println(res)
*/

func (m *Masters) addWork (db *xorm.Engine, work Works) (errWork Works,err error) {
	errWork = Works{}
	has, err := db.
		Where("date_work > CURRENT_TIMESTAMP - interval '1 day'").
		And("master_id = ?", m.Id).
		Exist(&work)
	if err != nil {
		return
	}
	if has {
		err = fmt.Errorf("this master already has work for this day")
		return
	}
	_, err = db.Insert(&work)
	if err != nil {
		return
	}
	return work, nil
}

//3.1.1

func (c *Cars) delete (db *xorm.Engine) error {
	_, err := db.Delete(c)
	if err != nil {
		return err
	}
	return nil
}

//3.2.2

func (m *Masters) deleteServices (db *xorm.Engine) error {
	works := make([]Works,0)
	db.
		Sql("SELECT service_id, count(master_id) " +
			"FROM " +
			"	( SELECT master_id, service_id " +
			"	  FROM works " +
			"	  GROUP BY master_id, service_id" +
			"	) AS tab1 " +
			"GROUP BY service_id " +
			"HAVING COUNT(*) = 1").
		Find(&works)

	in := make([]int64,0)
	for _, v := range works {
		in = append(in, v.Service)
	}
	works = make([]Works,0)
	db.
		In("service_id", in).
		Where("master_id = ?", m.Id).
		GroupBy("service_id").
		Find(&works)

	in = make([]int64,0)
	for _, v := range works {
		in = append(in, v.Service)
	}

	_, err := db.
		In("service_id", in).
		Delete(new(Works))
	if err != nil {
		return err
	}

	_, err = db.
		In("id", in).
		Delete(new(Services))
	if err != nil {
		return err
	}

	_, err = db.Delete(m)
	if err != nil {
		return err
	}

	return nil
}