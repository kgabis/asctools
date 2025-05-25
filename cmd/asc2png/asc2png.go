package asc2png

import (
	"bufio"
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
	fs := flag.NewFlagSet("asc2png", flag.ExitOnError)

	var outputFileName string
	fs.StringVar(&outputFileName, "output", "", "Path to the output .png file")

	fs.Parse(args)

	reader := bufio.NewReader(os.Stdin)
	elevationMap, err := lidartools.ParseASCFile(reader)

	img, err := renderMapToImage(elevationMap)
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

func renderMapToImage(elevationMap *lidartools.ElevationMap) (image.Image, error) {
	width := elevationMap.MaxX - elevationMap.MinX
	height := elevationMap.MaxY - elevationMap.MinY

	img := image.NewRGBA64(image.Rect(0, 0, width, height))

	elevationRange := elevationMap.MaxElevation - elevationMap.MinElevation

	for y := elevationMap.MinY; y < elevationMap.MaxY; y++ {
		for x := elevationMap.MinX; x < elevationMap.MaxX; x++ {
			elevation := elevationMap.GetElevation(x, y)
			if elevation == lidartools.NodataValue {
				img.SetRGBA64(x-elevationMap.MinX, y-elevationMap.MinY, color.RGBA64{R: 0, G: 0, B: 0, A: 0})
			} else {
				normalized := (elevation - elevationMap.MinElevation) / elevationRange
				grayValue := uint16(normalized * math.MaxUint16)
				img.SetRGBA64(x-elevationMap.MinX, y-elevationMap.MinY, color.RGBA64{R: grayValue, G: grayValue, B: grayValue, A: math.MaxUint16})
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
