package main

import (
	"asctools/cmd"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No argument provided")
		return
	}
	switch os.Args[1] {
	case "asc2png":
		cmd.Asc2Png(os.Args[2:])
	case "crop":
		cmd.Crop(os.Args[2:])
	case "diffasc2png":
		cmd.DiffAsc2Png(os.Args[2:])
	case "merge":
		cmd.Merge(os.Args[2:])
	case "split":
		cmd.Split(os.Args[2:])
	case "asc2stl":
		cmd.Asc2Stl(os.Args[2:])
	case "denoise":
		cmd.Denoise(os.Args[2:])
	case "downscale":
		cmd.Downscale(os.Args[2:])
	default:
		fmt.Println("Unknown command")
	}
}
