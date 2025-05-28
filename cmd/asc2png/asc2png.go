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
	var absoluteElevation bool
	fs.BoolVar(&absoluteElevation, "absolute_elevation", false, "If true, encode raw (unscaled) elevation values in the PNG output")

	fs.Parse(args)

	reader := bufio.NewReader(os.Stdin)
	elevationMap, err := lidartools.ParseASCFile(reader)

	renderFunc := renderMapToImage
	if absoluteElevation {
		renderFunc = renderMapToImageAbsolute
	}

	img, err := renderFunc(elevationMap)
	if err != nil {
		fmt.Println("Error rendering map to bitmap:", err)
		return
	}

	err = png.Encode(os.Stdout, img)
	if err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return
	}
}

func renderMapToImage(elevationMap *lidartools.ElevationMap) (image.Image, error) {
	width := elevationMap.MaxX - elevationMap.MinX
	height := elevationMap.MaxY - elevationMap.MinY

	img := image.NewGray16(image.Rect(0, 0, width, height))

	elevationRange := elevationMap.MaxElevation - elevationMap.MinElevation

	for y := elevationMap.MinY; y < elevationMap.MaxY; y++ {
		for x := elevationMap.MinX; x < elevationMap.MaxX; x++ {
			elevation := elevationMap.GetElevation(x, y)
			if elevation == lidartools.NodataValue {
				img.SetGray16(x-elevationMap.MinX, y-elevationMap.MinY, color.Gray16{Y: 0})
			} else {
				normalized := (elevation - elevationMap.MinElevation) / elevationRange
				grayValue := uint16(normalized * math.MaxUint16)
				img.SetGray16(x-elevationMap.MinX, y-elevationMap.MinY, color.Gray16{Y: grayValue})
			}
		}
	}

	// Flip the image vertically
	for y := 0; y < height/2; y++ {
		for x := 0; x < width; x++ {
			top := img.Gray16At(x, y)
			bottom := img.Gray16At(x, height-y-1)
			img.SetGray16(x, y, bottom)
			img.SetGray16(x, height-y-1, top)
		}
	}

	return img, nil
}

func renderMapToImageAbsolute(elevationMap *lidartools.ElevationMap) (image.Image, error) {
	width := elevationMap.MaxX - elevationMap.MinX
	height := elevationMap.MaxY - elevationMap.MinY

	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := elevationMap.MinY; y < elevationMap.MaxY; y++ {
		for x := elevationMap.MinX; x < elevationMap.MaxX; x++ {
			elevation := elevationMap.GetElevation(x, y)
			if elevation == lidartools.NodataValue {
				img.SetNRGBA(x-elevationMap.MinX, y-elevationMap.MinY, color.NRGBA{R: 0, G: 0, B: 0, A: 0})
			} else {
				intPart := int(math.Floor(elevation))
				fracPart := elevation - float64(intPart)
				r := uint8((intPart >> 8) & 0xFF) // high byte of meters
				g := uint8(intPart & 0xFF)        // low byte of meters
				b := uint8(fracPart * 255)        // fractional part in blue
				img.SetNRGBA(
					x-elevationMap.MinX,
					y-elevationMap.MinY,
					color.NRGBA{R: r, G: g, B: b, A: 255},
				)
			}
		}
	}

	// Flip the image vertically
	for y := 0; y < height/2; y++ {
		for x := 0; x < width; x++ {
			top := img.NRGBAAt(x, y)
			bottom := img.NRGBAAt(x, height-y-1)
			img.SetNRGBA(x, y, bottom)
			img.SetNRGBA(x, height-y-1, top)
		}
	}

	return img, nil
}
