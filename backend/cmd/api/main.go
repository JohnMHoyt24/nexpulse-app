package main

import (
	"fmt"
	"log"
)

func greeting(service string) string {
	return fmt.Sprintf("%s is running!", service)
}

func main() {
	log.Println(greeting("Nexpulse"))
}
