package main

import (
	"fmt"
	"lidartools/cmd/asc2png"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No argument provided")
	}
	switch os.Args[1] {
	case "asc2png":
		asc2png.Asc2Png(os.Args[2:])
		break
	default:
		fmt.Println("Unknown command")
	}
}
