package main

import (
	"fmt"
)

func main() {
	db := connectDatabase()
	defer db.Close()

	works := TopMasters{}.get(db)
	fmt.Print(works)
}



/*type TopMasters struct {
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
}*/