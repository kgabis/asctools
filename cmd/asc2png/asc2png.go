package asc2png

import (
	"flag"
	"fmt"
)

func Asc2Png() {
	var inputDir string
	flag.StringVar(&inputDir, "input_dir", "", "Path to the input dir with .asc files that will be merged")

	var outputFileName string
	flag.StringVar(&outputFileName, "output", "", "Path to the output .stl file")

	var downscaleFactor int
	flag.IntVar(&downscaleFactor, "downscale", 1, "Downscale factor (must be greater than 1)")

	var resultScale float64
	flag.Float64Var((*float64)(&resultScale), "scale", 1.0, "Scale factor for the result (must be greater than 0)")

	var floorElevation float64
	flag.Float64Var(&floorElevation, "floor", 0.0, "Floor elevation level (default is 0.0)")

	flag.Parse()

	fmt.Println("hello world")
}
