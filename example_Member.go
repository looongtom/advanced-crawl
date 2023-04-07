package main

import (
	"fmt"
	"strconv"
	"time"
)

type Member struct {
	Name     string
	Verified bool
}

var (
	numMember = 10000

	memberQueue = make(chan Member)
)

func checkVerify() {
	for member := range memberQueue {
		fmt.Printf("Checking member %s\n", member.Name)

		time.Sleep(time.Second * 3)

		if member.Verified == false {
			member.Verified = true
		}

		fmt.Printf("Member %s verified\n", member.Name)
	}
}

func main() {
	numWorkers := 10

	for i := 1; i <= numWorkers; i++ {
		go checkVerify()
	}

	for i := 1; i <= numMember; i++ {
		strName := strconv.Itoa(i)
		memberQueue <- Member{Name: strName, Verified: false}
	}

}
