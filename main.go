package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

func main() {
	ginServer()
	db := connectGORMDatabase()
	autoMigrate(db)

	t0 := time.Now()


	fmt.Println(topFiveMasters(db))


	t1 := time.Now()
	fmt.Printf("Elapsed time: %v", t1.Sub(t0))


	return
}







