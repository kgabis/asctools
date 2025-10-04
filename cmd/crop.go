package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	asctools "asctools/internal"
)

func Crop(args []string) {
	fs := flag.NewFlagSet("crop", flag.ExitOnError)

	var inputFile string
	fs.StringVar(&inputFile, "input", "", "Path to the input ASC file to crop (default: stdin)")

	var relative bool
	fs.BoolVar(&relative, "relative", false, "Use relative coordinates (0-1). If false, use absolute indices (0..width, 0..height)")

	var startX float64
	fs.Float64Var(&startX, "start_x", 0.0, "Start X coordinate (relative: 0-1; absolute: 0..width)")

	var startY float64
	fs.Float64Var(&startY, "start_y", 0.0, "Start Y coordinate (relative: 0-1; absolute: 0..height)")

	var endX float64
	fs.Float64Var(&endX, "end_x", 1.0, "End X coordinate (relative: 0-1; absolute: 0..width)")

	var endY float64
	fs.Float64Var(&endY, "end_y", 1.0, "End Y coordinate (relative: 0-1; absolute: 0..height)")

	fs.Parse(args)

	// Open and parse the input ASC file (from file or stdin)
	var reader *bufio.Reader
	if inputFile == "" {
		reader = bufio.NewReader(os.Stdin)
	} else {
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = bufio.NewReader(file)
	}

	elevationMap, err := asctools.ParseASCFile(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		os.Exit(1)
	}

	var relStartX, relStartY, relEndX, relEndY float64
	if relative {
		if startX < 0 || startX > 1 || startY < 0 || startY > 1 ||
			endX < 0 || endX > 1 || endY < 0 || endY > 1 {
			fmt.Fprintln(os.Stderr, "Error: when -relative is true, all coordinates must be in range [0, 1]")
			os.Exit(1)
		}
		relStartX, relStartY = startX, startY
		relEndX, relEndY = endX, endY

		if relStartX >= relEndX || relStartY >= relEndY {
			fmt.Fprintln(os.Stderr, "Error: start coordinates must be less than end coordinates")
			os.Exit(1)
		}
	} else {
		minX := elevationMap.MinX
		minY := elevationMap.MinY
		maxX := elevationMap.MaxX
		maxY := elevationMap.MaxY

		if startX < minX || startX > maxX || endX < minX || endX > maxX ||
			startY < minY || startY > maxY || endY < minY || endY > maxY {
			fmt.Fprintf(os.Stderr, "Error: when -relative is false, x must be in [%.1f, %.1f], y in [%.1f, %.1f]\n",
				minX, maxX, minY, maxY)
			os.Exit(1)
		}
	}

	var croppedMap *asctools.ElevationMap
	if relative {
		croppedMap, err = elevationMap.CropRelative(relStartX, relStartY, relEndX, relEndY)
	} else {
		croppedMap, err = elevationMap.Crop(startX, startY, endX, endY)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error cropping map: %v\n", err)
		os.Exit(1)
	}

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	err = croppedMap.WriteASC(writer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing cropped map: %v\n", err)
		os.Exit(1)
	}
}
