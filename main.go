package main

import (
	"fmt"
	"os"

	"github.com/ekobayong/gorm-model-generator/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
