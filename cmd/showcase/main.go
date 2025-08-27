package main

import (
	"fmt"

	"github.com/samtaborsky/containerlib/docker"
	"github.com/samtaborsky/containerlib/types"
)

func main() {
	fmt.Println("Hello World")
	docker.TestFunctionEvents()
	var errorVar = types.Error{
		Message: "Hello World",
	}
	fmt.Println(errorVar)
}
