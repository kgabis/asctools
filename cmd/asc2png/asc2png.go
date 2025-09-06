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
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ASC file: %v\n", err)
		return
	}

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
	imgWidth := int(elevationMap.GetWidth() / elevationMap.CellSize)
	imgHeight := int(elevationMap.GetHeight() / elevationMap.CellSize)
	img := image.NewGray16(image.Rect(0, 0, imgWidth, imgHeight))

	elevationRange := elevationMap.MaxElevation - elevationMap.MinElevation

	for y := 0.0; y < elevationMap.GetHeight(); y += elevationMap.CellSize {
		for x := 0.0; x < elevationMap.GetWidth(); x += elevationMap.CellSize {
			elevation := elevationMap.GetElevation(elevationMap.MinX+x, elevationMap.MinY+y)
			if elevation == lidartools.NodataValue {
				img.SetGray16(int(x/elevationMap.CellSize), int(y/elevationMap.CellSize), color.Gray16{Y: 0})
			} else {
				normalized := (elevation - elevationMap.MinElevation) / elevationRange
				grayValue := uint16(normalized * math.MaxUint16)
				img.SetGray16(int(x/elevationMap.CellSize), int(y/elevationMap.CellSize), color.Gray16{Y: grayValue})
			}
		}
	}

	// Flip the image vertically
	for y := 0; y < imgHeight/2; y++ {
		for x := 0; x < imgWidth; x++ {
			top := img.Gray16At(x, y)
			bottom := img.Gray16At(x, imgHeight-y-1)
			img.SetGray16(x, y, bottom)
			img.SetGray16(x, imgHeight-y-1, top)
		}
	}

	return img, nil
}

func renderMapToImageAbsolute(elevationMap *lidartools.ElevationMap) (image.Image, error) {
	imgWidth := int(elevationMap.GetWidth() / elevationMap.CellSize)
	imgHeight := int(elevationMap.GetHeight() / elevationMap.CellSize)
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	for y := 0.0; y < elevationMap.GetHeight(); y += elevationMap.CellSize {
		for x := 0.0; x < elevationMap.GetWidth(); x += elevationMap.CellSize {
			elevation := elevationMap.GetElevation(elevationMap.MinX+x, elevationMap.MinY+y)
			imgX := int(x / elevationMap.CellSize)
			imgY := int(y / elevationMap.CellSize)
			// Encode elevation as:
			if elevation == lidartools.NodataValue {
				img.SetNRGBA(imgX, imgY, color.NRGBA{R: 0, G: 0, B: 0, A: 0})
			} else {
				intPart := int(math.Floor(elevation))
				fracPart := elevation - float64(intPart)
				r := uint8((intPart >> 8) & 0xFF)
				g := uint8(intPart & 0xFF)
				b := uint8(fracPart * 255)
				img.SetNRGBA(
					imgX,
					imgY,
					color.NRGBA{R: r, G: g, B: b, A: 255},
				)
			}
		}
	}

	// Flip the image vertically
	for y := 0; y < imgHeight/2; y++ {
		for x := 0; x < imgWidth; x++ {
			top := img.NRGBAAt(x, y)
			bottom := img.NRGBAAt(x, imgHeight-y-1)
			img.SetNRGBA(x, y, bottom)
			img.SetNRGBA(x, imgHeight-y-1, top)
		}
	}

	return img, nil
}
