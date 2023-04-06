package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Member struct {
	Name     string
	Verified bool
}

var (
	numMember = 10000

	memberQueue = make(chan Member, numMember)
)

func checkVerify(wg *sync.WaitGroup) {
	defer wg.Done()
	for member := range memberQueue {
		fmt.Printf("Checking member %s\n", member.Name)
		// do work here
		if member.Verified == false {
			member.Verified = true
		}
		fmt.Printf("Member %s verified\n", member.Name)
		time.Sleep(time.Second * 3)
	}
}

func main() {
	numWorkers := 10

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 1; i <= numWorkers; i++ {
		go checkVerify(&wg)
	}

	var addWg sync.WaitGroup
	addWg.Add(1)
	go func() {
		defer addWg.Done()
		for i := 1; i <= numMember; i++ {
			strName := strconv.Itoa(i)
			memberQueue <- Member{Name: strName, Verified: false}
		}
		close(memberQueue)
	}()

	wg.Wait()

}
