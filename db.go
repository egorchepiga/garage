package main

import (
	"fmt"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func connectDatabase() *xorm.Engine{
	config := loadConfiguration("./config.json")
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.Db)
	var err error
	db, err := xorm.NewEngine("postgres", dbInfo)
	if err!=nil{
		log.Println("engine creation failed", err)
	}

	err = db.Ping()
	if err !=nil{
		panic(err)
	}

	log.Println("Successfully connected")

	syncTables(db)
	return db
}

func syncTables(db *xorm.Engine){
	err:= db.Sync(new(Cars))
	err= db.Sync(new(Masters))
	err= db.Sync(new(Services))
	err= db.Sync(new(Works))
	if err!=nil{
		log.Println("creation error",err)
		return
	}
	log.Println("Successfully synced")
}

type JoinedWork struct{
	Works `xorm:"extends"`
	Cars `xorm:"extends"`
	Masters `xorm:"extends"`
	Services `xorm:"extends"`
}

func (JoinedWork) TableName() string{
	return "works"
}

type Cars struct {
	Id        int32			`xorm:"'id' pk autoincr"`
	Num       string		`xorm:"'num' varchar(20)"`
	Color     string		`xorm:"'color' varchar(20)"`
	Mark      string		`xorm:"'mark' varchar(20)"`
	IsForeign bool			`xorm:"'is_foreign'"`
}

type Masters struct {
	Id   int32  			`xorm:"'id' pk autoincr"`
	Name string				`xorm:"'name' varchar(50)"`
}

type Works struct {
	Id   		int32		`xorm:"'id' pk autoincr"`
	Date   		time.Time   `xorm:"'date_work'"`
	Master		int32		`xorm:"'master_id'"`
	Car			int32		`xorm:"'car_id'"`
	Service		int32		`xorm:"'service_id'"`
}

type Services struct {
	Id 			int32		`xorm:"'id' pk autoincr"`
	Name 		string		`xorm:"'name'"`
	CostOur		float64		`xorm:"'cost_our'"`
	CostForeign	float64		`xorm:"'cost_foreign'"`
}


func (s Services) String() string {
	return fmt.Sprintf("Services<%d %s %g %g>\n", s.Id, s.Name, s.CostOur, s.CostForeign)
}

func (w Works) String() string {
	return fmt.Sprintf("Work <%d %v %v %v %v>\n", w.Id, w.Date, w.Master, w.Car, w.Service)
}

func (m Masters) String() string {
	return fmt.Sprintf("Master <%d %s>\n", m.Id, m.Name)
}

func (c Cars) String() string {
	return fmt.Sprintf("Car <%d %s %s %s %t>\n", c.Id, c.Num, c.Color, c.Mark, c.IsForeign)
}

func (j JoinedWork) String() string {
	return fmt.Sprintf("JOINED Work < \n" +
		"\t id: %d time: %v\n" +
		"\t Car: %d num: %s color: %s mark: %s foreign: %t \n" +
		"\t master: %s\n" +
		"\t service: %s >\n",
		j.Works.Id, j.Works.Date,
		j.Cars.Id, j.Cars.Num, j.Cars.Color, j.Cars.Mark, j.Cars.IsForeign,
		j.Masters.Name,
		j.Services.Name)
}

func (JoinedWork) get (db *xorm.Engine) (*xorm.Session) {
	return db.
		Join("INNER", "masters", "masters.id = works.master_id").
		Join("INNER", "cars", "cars.id = works.car_id").
		Join("INNER", "services", "services.id = works.service_id")
}
