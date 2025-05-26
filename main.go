package main

import (
	"fmt"
	"lidartools/cmd/asc2png"
	"lidartools/cmd/diffasc2png"
	"lidartools/cmd/merge"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No argument provided")
	}
	switch os.Args[1] {
	case "asc2png":
		asc2png.Cmd(os.Args[2:])
		break
	case "diffasc2png":
		diffasc2png.Cmd(os.Args[2:])
	case "merge":
		merge.Cmd(os.Args[2:])
	default:
		fmt.Println("Unknown command")
	}
}
