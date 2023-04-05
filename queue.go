package main

import (
	"time"
)

var GenerateDomain = func() {
	go func() {
		for range time.Tick(time.Second * 5) {

		}
	}()
}
