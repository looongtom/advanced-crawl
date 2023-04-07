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
	numMember = 10

	memberQueue = make(chan Member)
)

func checkVerify(results chan<- string) {
	for member := range memberQueue {
		if member.Name != "Waiting to end" {
			fmt.Printf("Checking member %s\n", member.Name)
		}

		time.Sleep(time.Second * 3)

		if member.Verified == false {
			member.Verified = true
		}
		results <- member.Name
		if member.Name != "Waiting to end" {
			fmt.Printf("Member %s verified\n", member.Name)
		}
	}
}

func main() {
	numWorkers := 2
	resultQueue := make(chan string, numMember)
	for i := 1; i <= numWorkers; i++ {
		go checkVerify(resultQueue)
	}

	for i := 1; i <= numMember; i++ {
		strName := strconv.Itoa(i)
		memberQueue <- Member{Name: strName, Verified: false}
	}

	endTask := Member{Name: "Waiting to end"}
	fmt.Println(endTask.Name)

	memberQueue <- endTask
	for result := range resultQueue {
		if result == endTask.Name {
			break
		}
	}

	close(memberQueue)
	close(resultQueue)

}
