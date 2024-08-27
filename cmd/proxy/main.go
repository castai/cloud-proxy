package main

import (
	"fmt"

	"github.com/castai/cloud-proxy/internal/gcpauth"
)

func main() {
	fmt.Println("Hi")

	src := gcpauth.GCPCredentialsSource{}
	fmt.Println(src)
}
