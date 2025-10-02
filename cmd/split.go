package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	lidartools "lidartools/internal"
)

func Split(args []string) {
	fs := flag.NewFlagSet("split", flag.ExitOnError)

	var outputDir string
	fs.StringVar(&outputDir, "output_dir", ".", "Directory to save the split ASC files")

	var nrows int
	fs.IntVar(&nrows, "nrows", 2, "Number of rows in the output grid")

	var ncols int
	fs.IntVar(&ncols, "ncols", 2, "Number of columns in the output grid")

	var uniformSize bool
	fs.BoolVar(&uniformSize, "uniform", false, "Make all tiles the same size (smaller of width/ncols and height/nrows), discarding extra space")

	var prefix string
	fs.StringVar(&prefix, "prefix", "tile", "Prefix for output filenames")

	fs.Parse(args)

	inputReader := bufio.NewReader(os.Stdin)

	elevationMap, err := lidartools.ParseASCFile(inputReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	if uniformSize {
		fmt.Printf("Splitting %fx%f map into %dx%d grid with uniform tile sizes\n",
			elevationMap.GetWidth(), elevationMap.GetHeight(), nrows, ncols)
	} else {
		fmt.Printf("Splitting %fx%f map into %dx%d grid\n",
			elevationMap.GetWidth(), elevationMap.GetHeight(), nrows, ncols)
	}

	tiles, err := elevationMap.Split(nrows, ncols, uniformSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error splitting map: %v\n", err)
		os.Exit(1)
	}

	tileCount := 0

	for row := 0; row < nrows; row++ {
		for col := 0; col < ncols; col++ {
			tile := tiles[row][col]
			if tile == nil {
				continue
			}

			filename := fmt.Sprintf("%s_%d_%d.asc", prefix, row, col)
			outputPath := filepath.Join(outputDir, filename)

			if err := writeTileToFile(tile, outputPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing tile %s: %v\n", filename, err)
				continue
			}

			fmt.Printf("Created tile: %s (%fx%f)\n", filename, tile.GetWidth(), tile.GetHeight())
			tileCount++
		}
	}

	fmt.Printf("Successfully split into %d tiles\n", tileCount)
}

func writeTileToFile(tile *lidartools.ElevationMap, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	return tile.WriteASC(writer)
}
