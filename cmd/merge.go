package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	lidartools "lidartools/internal"
)

func Merge(args []string) {
	fs := flag.NewFlagSet("merge", flag.ExitOnError)

	var inputDir string
	fs.StringVar(&inputDir, "input_dir", "", "Directory containing ASC files to merge")
	fs.Parse(args)

	if inputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: input_dir is required")
		os.Exit(1)
	}

	var maps []*lidartools.ElevationMap

	files, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input directory:", err)
		os.Exit(1)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".asc") {
			path := filepath.Join(inputDir, file.Name())
			reader, err := os.Open(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error opening file:", path, err)
				continue
			}
			bufioreader := bufio.NewReader(reader)
			defer reader.Close()
			slice, err := lidartools.ParseASCFile(bufioreader)
			reader.Close()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading ASC file:", path, err)
				continue
			}
			maps = append(maps, slice)
		}
	}

	if len(maps) == 0 {
		fmt.Fprintln(os.Stderr, "No ASC files found in the input directory")
		os.Exit(1)
	}

	mergedMap, err := lidartools.MergeMaps(maps)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error merging elevation maps:", err)
		os.Exit(1)
	}

	err = mergedMap.WriteASC(bufio.NewWriter(os.Stdout))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing merged map to stdout:", err)
		os.Exit(1)
	}
}
