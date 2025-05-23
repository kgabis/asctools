package diffasc2png

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	lidartools "lidartools/internal"
	"math"
	"os"
)

func Cmd(args []string) {
	fs := flag.NewFlagSet("diffasc2png", flag.ExitOnError)

	var inputDir1 string
	fs.StringVar(&inputDir1, "input_dir1", "", "Path to the input dir 1 with .asc files that will be merged")

	var inputDir2 string
	fs.StringVar(&inputDir2, "input_dir2", "", "Path to the input dir 2 with .asc files that will be merged")

	var outputFileName string
	fs.StringVar(&outputFileName, "output", "", "Path to the output .png file")

	fs.Parse(args)

	if inputDir1 == "" || inputDir2 == "" || outputFileName == "" {
		flag.Usage()
		return
	}

	elevationMap1, err := lidartools.ASCDirToElevationMap(inputDir1)
	if err != nil {
		fmt.Println("Error reading elevation map:", err)
		return
	}

	elevationMap2, err := lidartools.ASCDirToElevationMap(inputDir2)
	if err != nil {
		fmt.Println("Error reading elevation map:", err)
		return
	}

	img, err := renderMapDiffToImage(elevationMap1, elevationMap2)
	if err != nil {
		fmt.Println("Error rendering map to bitmap:", err)
		return
	}

	file, err := os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return
	}

	fmt.Printf("Result written to %s\n", outputFileName)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderMapDiffToImage(elevationMap1 *lidartools.ElevationMap, elevationMap2 *lidartools.ElevationMap) (image.Image, error) {
	minX := max(elevationMap1.MinX, elevationMap2.MinX)
	maxX := min(elevationMap1.MaxX, elevationMap2.MaxX)
	minY := max(elevationMap1.MinY, elevationMap2.MinY)
	maxY := min(elevationMap1.MaxY, elevationMap2.MaxY)

	width := maxX - minX
	height := maxY - minY

	img := image.NewRGBA64(image.Rect(0, 0, width, height))

	elevationRange1 := elevationMap1.MaxElevation - elevationMap1.MinElevation
	elevationRange2 := elevationMap2.MaxElevation - elevationMap2.MinElevation
	elevationRange := math.Max(elevationRange1, elevationRange2)

	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			elevation1 := elevationMap1.GetElevation(x, y)
			elevation2 := elevationMap2.GetElevation(x, y)
			imgX := x - minX
			imgY := x - minY
			if elevation1 == lidartools.NodataValue || elevation2 == lidartools.NodataValue {
				img.SetRGBA64(imgX, imgY, color.RGBA64{R: 0, G: 0, B: 0, A: 0})
			} else {
				normalized := (elevation1 - elevationMap1.MinElevation) / elevationRange
				grayValue := uint16(normalized * math.MaxUint16)
				img.SetRGBA64(imgX, imgY, color.RGBA64{R: grayValue, G: grayValue, B: grayValue, A: math.MaxUint16})
			}
		}
	}

	// Flip the image vertically
	for y := 0; y < height/2; y++ {
		for x := 0; x < width; x++ {
			top := img.RGBA64At(x, y)
			bottom := img.RGBA64At(x, height-y-1)
			img.SetRGBA64(x, y, bottom)
			img.SetRGBA64(x, height-y-1, top)
		}
	}

	return img, nil
}
