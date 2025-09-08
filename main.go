package main

import (
	"fmt"
	"lidartools/cmd/asc2png"
	"lidartools/cmd/asc2stl"
	"lidartools/cmd/crop"
	"lidartools/cmd/denoise"
	"lidartools/cmd/diffasc2png"
	"lidartools/cmd/downscale"
	"lidartools/cmd/merge"
	"lidartools/cmd/split"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No argument provided")
	}
	switch os.Args[1] {
	case "asc2png":
		asc2png.Cmd(os.Args[2:])
	case "crop":
		crop.Cmd(os.Args[2:])
	case "diffasc2png":
		diffasc2png.Cmd(os.Args[2:])
	case "merge":
		merge.Cmd(os.Args[2:])
	case "split":
		split.Cmd(os.Args[2:])
	case "asc2stl":
		asc2stl.Cmd(os.Args[2:])
	case "denoise":
		denoise.Cmd(os.Args[2:])
	case "downscale":
		downscale.Cmd(os.Args[2:])
	default:
		fmt.Println("Unknown command")
	}
}
