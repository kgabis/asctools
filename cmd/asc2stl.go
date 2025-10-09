package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	asctools "github.com/kgabis/asctools/pkg"
)

func Asc2Stl(args []string) {
	fs := flag.NewFlagSet("asc2stl", flag.ExitOnError)

	var resultScale float64
	fs.Float64Var((*float64)(&resultScale), "scale", 1.0, "Scale factor for the result (must be greater than 0)")

	var floorElevation float64
	fs.Float64Var(&floorElevation, "floor", 0.0, "Floor elevation level (default is 0.0)")

	var floorMargin float64
	fs.Float64Var(&floorMargin, "floor_margin", 0.0, "Margin to add around the base of the model (only works when floor is not set)")

	fs.Parse(args)

	reader := bufio.NewReader(os.Stdin)
	elevationMap, err := asctools.ParseASCFile(reader)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		return
	}

	if floorElevation == 0 && floorMargin > 0 {
		floorElevation = elevationMap.MinElevation - floorMargin
	}

	err = elevationMap.WriteSTL(bufio.NewWriter(os.Stdout), floorElevation)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing STL to stdout:", err)
		os.Exit(1)
	}
}
