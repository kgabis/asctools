package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	lidartools "lidartools/internal"
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

	elevationMap, err := lidartools.ParseASCFile(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Original map: %fx%f\n", elevationMap.GetWidth(), elevationMap.GetHeight())

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
		// Validate absolute indices against map size
		minX := elevationMap.MinX
		minY := elevationMap.MinY
		w := elevationMap.GetWidth()
		h := elevationMap.GetHeight()

		if startX < minX || startX > (minX+w) || endX < 0 || endX > (minX+w) ||
			startY < minY || startY > (minY+h) || endY < 0 || endY > (minY+h) {
			fmt.Fprintf(os.Stderr, "Error: when -relative is false, start_x must be in [%.1f, %.1f], end_x in [0, %.1f], start_y in [%.1f, %.1f], end_y in [0, %.1f]\n",
				minX, minX+w, minX+w, minY, minY+h, minY+h)
			os.Exit(1)
		}
	}

	var croppedMap *lidartools.ElevationMap
	if relative {
		fmt.Fprintf(os.Stderr, "Cropping (relative) from (%.3f, %.3f) to (%.3f, %.3f)\n", relStartX, relStartY, relEndX, relEndY)
		croppedMap, err = elevationMap.CropRelative(relStartX, relStartY, relEndX, relEndY)
	} else {
		fmt.Fprintf(os.Stderr, "Cropping (absolute indices) from (%.3f, %.3f) to (%.3f, %.3f)\n", startX, startY, endX, endY)
		croppedMap, err = elevationMap.Crop(startX, startY, endX, endY)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error cropping map: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Cropped map: %fx%f\n", croppedMap.GetWidth(), croppedMap.GetHeight())

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	err = croppedMap.WriteASC(writer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing cropped map: %v\n", err)
		os.Exit(1)
	}
}
