package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	asctools "github.com/kgabis/asctools/pkg"
)

func Asc2Png(args []string) {
	fs := flag.NewFlagSet("asc2png", flag.ExitOnError)

	var scale float64
	fs.Float64Var(&scale, "scale", 1.0, "Scale factor for the result (must be greater than 1)")

	var scalingOperationVal string
	fs.StringVar(&scalingOperationVal, "scaling_operation", "none", "Scaling operation: 'up' to scale up, 'down' to downscale")

	fs.Parse(args)

	if scale < 1 {
		fmt.Fprintln(os.Stderr, "Error: scale must be greater than or equal to 1")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	elevationMap, err := asctools.ParseASCFile(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		os.Exit(1)
	}

	var scalingOperation asctools.ScalingOperation
	switch scalingOperationVal {
	case "up":
		scalingOperation = asctools.ScaleUp
	case "down":
		scalingOperation = asctools.ScaleDown
	case "none":
		scalingOperation = asctools.ScaleNone
	default:
		scalingOperation = asctools.ScaleNone
	}

	err = elevationMap.WritePNG(bufio.NewWriter(os.Stdout), scalingOperation, int(scale))
	if err != nil {
		fmt.Println("Error rendering map to png:", err)
		os.Exit(1)
	}

}
