package main

import (
	"log"

	"github.com/castai/cloud-proxy/internal/castai/dummy"
)

func main() {
	log.Println("Starting mock cast instance")
	mockCast := &dummy.MockCast{}
	if err := mockCast.Run(); err != nil {
		log.Panicln("Error running mock Cast:", err)
	}
}
