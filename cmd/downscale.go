package cmd

import (
	"bufio"
	"flag"
	"fmt"
	asctools "asctools/internal"
	"os"
)

func Downscale(args []string) {
	fs := flag.NewFlagSet("downscale", flag.ExitOnError)

	var downscaleFactor int
	fs.IntVar(&downscaleFactor, "factor", 1, "Downscale factor (must be greater than 1)")

	fs.Parse(args)

	reader := bufio.NewReader(os.Stdin)
	elevationMap, err := asctools.ParseASCFile(reader)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		return
	}

	downscaled, err := elevationMap.Downscale(downscaleFactor)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error downscaling elevation map:", err)
		os.Exit(1)
	}

	err = downscaled.WriteASC(bufio.NewWriter(os.Stdout))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing ASC to stdout:", err)
		os.Exit(1)
	}
}
