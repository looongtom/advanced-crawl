package main

import (
	"fmt"
	"sync"
)

func producer(originalChan chan<- int) {
	for i := 0; i < 5; i++ {
		originalChan <- i
	}
	close(originalChan)
}

func consumer(originalChan <-chan int, wg *sync.WaitGroup) {
	for val := range originalChan {
		fmt.Println("Received value:", val)
	}
	wg.Done()
}
func main() {
	originalChan := make(chan int)
	var wg sync.WaitGroup
	wg.Add(1)
	go producer(originalChan)
	go consumer(originalChan, &wg)
	wg.Wait()

}
