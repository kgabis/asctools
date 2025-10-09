package main

import (
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
		Asc2Png(os.Args[2:])
	case "crop":
		Crop(os.Args[2:])
	case "diffasc2png":
		DiffAsc2Png(os.Args[2:])
	case "merge":
		Merge(os.Args[2:])
	case "split":
		Split(os.Args[2:])
	case "asc2stl":
		Asc2Stl(os.Args[2:])
	case "denoise":
		Denoise(os.Args[2:])
	case "downscale":
		Downscale(os.Args[2:])
	default:
		fmt.Println("Unknown command")
	}
}
