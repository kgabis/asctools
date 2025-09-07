package asc2stl

import (
	"bufio"
	"flag"
	"fmt"
	lidartools "lidartools/internal"
	"os"
)

func Cmd(args []string) {
	fs := flag.NewFlagSet("asc2stl", flag.ExitOnError)

	var downscaleFactor int
	flag.IntVar(&downscaleFactor, "downscale", 1, "Downscale factor (must be greater than 1)")

	var resultScale float64
	flag.Float64Var((*float64)(&resultScale), "scale", 1.0, "Scale factor for the result (must be greater than 0)")

	var floorElevation float64
	flag.Float64Var(&floorElevation, "floor", 0.0, "Floor elevation level (default is 0.0)")

	fs.Parse(args)

	reader := bufio.NewReader(os.Stdin)
	elevationMap, err := lidartools.ParseASCFile(reader)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		return
	}

	err = elevationMap.WriteSTL(bufio.NewWriter(os.Stdout))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing STL to stdout:", err)
		os.Exit(1)
	}
}
