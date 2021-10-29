package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func main() {
	err := _AssignTask2UserAsIs("13819271717", "1")
	stack_err := errors.WithStack(err)
	fmt.Println(stack_err)
}
