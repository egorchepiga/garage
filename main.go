package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

func main() {
	ginServer()


	t0 := time.Now()
	t1 := time.Now()
	fmt.Printf("Elapsed time: %v", t1.Sub(t0))


	return
}







