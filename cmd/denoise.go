package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	asctools "github.com/kgabis/asctools/internal"
)

func Denoise(args []string) {
	fs := flag.NewFlagSet("denoise", flag.ExitOnError)

	var window int
	fs.IntVar(&window, "window", 3, "Window size for median filtering (must be odd)")

	fs.Parse(args)

	reader := bufio.NewReader(os.Stdin)
	elevationMap, err := asctools.ParseASCFile(reader)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		return
	}

	denoised, err := elevationMap.Denoise(window)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error denoising elevation map:", err)
		os.Exit(1)
	}

	err = denoised.WriteASC(bufio.NewWriter(os.Stdout))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing denoised map to stdout:", err)
		os.Exit(1)
	}
}
