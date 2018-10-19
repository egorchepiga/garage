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
		"\t Car: %d num: %s color: %s mark: %s foreign: %t \n"+
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
		"\t Car: %d num: %s color: %s mark: %s foreign: %t \n"+
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