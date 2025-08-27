package docker

import (
	"fmt"

	"github.com/samtaborsky/containerlib/internal/client"
	"github.com/samtaborsky/containerlib/internal/util"
)

func TestFunctionContainers() {
	fmt.Println("Hello Containers!")
	client.TestFunctionClient()
	util.TestFunctionParse()
}
