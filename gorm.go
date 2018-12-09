package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"time"
)

func connectGORMDatabase() *gorm.DB{
	config := loadConfiguration("./config.json")
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s ",
		config.Database.Host, config.Database.Port, config.Database.User, config.Database.Db, config.Database.Password)
	var err error
	db, err := gorm.Open("postgres", dbInfo)
	if err!=nil{
		log.Println("engine creation failed", err)
	} else {
		log.Println("Successfully connected")
	}
	return db
}

func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(&Car{}, &Master{}, &Service{}, &Work{}, &PureWork{})
}


type Work struct {
	Id         int			`gorm:"column:id; 			primary_key;	AUTO_INCREMENT"`
	Date       time.Time   `gorm:"column:date_work;"`
	Master_Id  int
	Master     Master
	Car_Id     int
	Car        Car
	Service_Id int
	Service    Service
}

type PureWork struct {
	Id   		int			`gorm:"column:id; 			primary_key;	AUTO_INCREMENT"`
	Date   		time.Time   `gorm:"column:date_work;"`
	Master_Id	int			`gorm:"column:master_id;"`
	Car_Id		int			`gorm:"column:car_id;"`
	Service_Id	int			`gorm:"column:service_id;"`
}

func (PureWork) TableName() string {
	return "works"
}

type Car struct {
	Id         int    `gorm:"column:id; 		primary_key; 	not null;	AUTO_INCREMENT"`
	Num        string `gorm:"column:num;' 	type:varchar(20)"`
	Color      string `gorm:"column:color; 	type:varchar(20)"`
	Mark       string `gorm:"column:mark; 	type:varchar(20)"`
	Is_Foreign int    `gorm:"column:is_foreign; type:smallint"`
}

type Master struct {
	Id   int  				`gorm:"column:id; 		primary_key; 	not null;	AUTO_INCREMENT"`
	Name string				`gorm:"column:name; 	type:varchar(50)"`
}

type Service struct {
	Id           int     `gorm:"column:id; 		primary_key;	AUTO_INCREMENT"`
	Name         string  `gorm:"column:name;"`
	Cost_Our     float64 `gorm:"column:cost_our;"`
	Cost_Foreign float64 `gorm:"column:cost_foreign;"`
}


func (s Service) String() string {
	return fmt.Sprintf("Services<%d %s %g %g>\n", s.Id, s.Name, s.Cost_Our, s.Cost_Foreign)
}

func (w Work) String() string {
	return fmt.Sprintf("Work #%d \n\t date:\t\t %v \n\t master_id:\t %d \n\t car_id:\t %d \n\t service_id: %d \n\t%v \t%v \t%v\n", w.Id, w.Date, w.Master_Id, w.Car_Id, w.Service_Id,  w.Master, w.Car, w.Service)
}

func (m Master) String() string {
	return fmt.Sprintf("Master <%d %s>\n", m.Id, m.Name)
}

func (c Car) String() string {
	return fmt.Sprintf("Car <%d %s %s %s %d>\n", c.Id, c.Num, c.Color, c.Mark, c.Is_Foreign)
}
