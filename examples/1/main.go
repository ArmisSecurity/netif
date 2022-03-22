package main

import (
	"fmt"

	ni "github.com/xtmono/netif"
)

func main() {
	is, err := ni.Parse(
		ni.Path("input"),
	)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	is.Write(
		ni.Path("output"),
	)
}
