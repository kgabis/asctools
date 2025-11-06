package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	asctools "github.com/kgabis/asctools/pkg"
)

func Subtract(args []string) {
	fs := flag.NewFlagSet("subtract", flag.ExitOnError)

	var input1 string
	fs.StringVar(&input1, "input1", "", "Path to the input 1 .asc file")

	var input2 string
	fs.StringVar(&input2, "input2", "", "Path to the input 2 .asc file")

	fs.Parse(args)

	if input1 == "" || input2 == "" {
		fs.Usage()
		os.Exit(1)
	}

	file1, err := os.Open(input1)
	if err != nil {
		fmt.Println("Error opening input file 1:", err)
		os.Exit(1)
	}
	defer file1.Close()

	file2, err := os.Open(input2)
	if err != nil {
		fmt.Println("Error opening input file 2:", err)
		os.Exit(1)
	}
	defer file2.Close()

	elevationMap1, err := asctools.ParseASCFile(bufio.NewReader(file1))
	if err != nil {
		fmt.Println("Error reading elevation map:", err)
		os.Exit(1)
	}

	elevationMap2, err := asctools.ParseASCFile(bufio.NewReader(file2))
	if err != nil {
		fmt.Println("Error reading elevation map:", err)
		os.Exit(1)
	}

	result, err := elevationMap1.Subtract(elevationMap2)
	if err != nil {
		fmt.Println("Error subtracting elevation maps:", err)
		os.Exit(1)
	}

	err = result.WriteASC(bufio.NewWriter(os.Stdout))
	if err != nil {
		fmt.Println("Error writing result to stdout:", err)
		os.Exit(1)
	}
}
