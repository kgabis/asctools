package crop

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	lidartools "lidartools/internal"
)

func Cmd(args []string) {
	fs := flag.NewFlagSet("crop", flag.ExitOnError)

	var inputFile string
	fs.StringVar(&inputFile, "input", "", "Path to the input ASC file to crop")

	var outputFile string
	fs.StringVar(&outputFile, "output", "", "Path to the output ASC file (default: stdout)")

	var startX float64
	fs.Float64Var(&startX, "start_x", 0.0, "Start X coordinate (0-1, where 0 is left edge)")

	var startY float64
	fs.Float64Var(&startY, "start_y", 0.0, "Start Y coordinate (0-1, where 0 is top edge)")

	var endX float64
	fs.Float64Var(&endX, "end_x", 1.0, "End X coordinate (0-1, where 1 is right edge)")

	var endY float64
	fs.Float64Var(&endY, "end_y", 1.0, "End Y coordinate (0-1, where 1 is bottom edge)")

	fs.Parse(args)

	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: input file is required")
		fs.Usage()
		os.Exit(1)
	}

	// Validate coordinate ranges
	if startX < 0 || startX > 1 || startY < 0 || startY > 1 ||
		endX < 0 || endX > 1 || endY < 0 || endY > 1 {
		fmt.Fprintln(os.Stderr, "Error: all coordinates must be in range [0, 1]")
		os.Exit(1)
	}

	if startX >= endX || startY >= endY {
		fmt.Fprintln(os.Stderr, "Error: start coordinates must be less than end coordinates")
		os.Exit(1)
	}

	// Open and parse the input ASC file
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	elevationMap, err := lidartools.ParseASCFile(bufio.NewReader(file))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Original map: %dx%d\n", elevationMap.Width, elevationMap.Height)
	fmt.Printf("Cropping from (%.3f, %.3f) to (%.3f, %.3f)\n", startX, startY, endX, endY)

	// Crop the map using relative coordinates
	croppedMap := elevationMap.CropRelative(startX, startY, endX, endY)
	if croppedMap == nil {
		fmt.Fprintln(os.Stderr, "Error: failed to crop map (invalid coordinates or resulting empty region)")
		os.Exit(1)
	}

	fmt.Printf("Cropped map: %dx%d\n", croppedMap.Width, croppedMap.Height)

	// Determine output destination
	var writer *bufio.Writer
	if outputFile == "" {
		// Write to stdout
		writer = bufio.NewWriter(os.Stdout)
	} else {
		// Write to file
		outFile, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer outFile.Close()
		writer = bufio.NewWriter(outFile)
	}
	defer writer.Flush()

	// Write the cropped map
	err = croppedMap.WriteASC(writer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing cropped map: %v\n", err)
		os.Exit(1)
	}

	if outputFile != "" {
		fmt.Printf("Successfully cropped and saved to: %s\n", outputFile)
	}
}
