package main

import (
	"fmt"
	"sync"
)

func main() {
	array := []int{}
	sum1 := 0
	sum2 := 0
	var wg sync.WaitGroup
	wg.Add(3)

	fmt.Println("old array", array)

	go func() {
		for i := 1; i <= 10; i++ {
			array = append(array, i)
		}
		defer wg.Done()
	}()

	go func() {
		for i := 1; i <= 5; i++ {
			sum1 += i
		}
		wg.Done()
	}()

	go func() {
		for i := 6; i <= 10; i++ {
			sum2 += i
		}
		wg.Done()
	}()

	wg.Wait()
	fmt.Println("new array")
	fmt.Println(array)
	fmt.Println(sum1, sum2, sum1+sum2)
}
