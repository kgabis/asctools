package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	asctools "github.com/kgabis/asctools/pkg"

	"golang.org/x/image/draw"
)

func Asc2Png(args []string) {
	fs := flag.NewFlagSet("asc2png", flag.ExitOnError)
	var absoluteElevation bool
	fs.BoolVar(&absoluteElevation, "absolute_elevation", false, "If true, encode raw (unscaled) elevation values in the PNG output")

	var scale float64
	fs.Float64Var(&scale, "scale", 1.0, "Scale factor for the result (must be greater than 1)")

	fs.Parse(args)

	if scale <= 1 {
		fmt.Fprintln(os.Stderr, "Error: scale must be greater than 1")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	elevationMap, err := asctools.ParseASCFile(reader)
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

	if scale != 1.0 {
		newWidth := int(float64(img.Bounds().Dx()) * scale)
		newHeight := int(float64(img.Bounds().Dy()) * scale)
		scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.NearestNeighbor.Scale(scaledImg, scaledImg.Bounds(), img, img.Bounds(), draw.Over, nil)
		img = scaledImg
	}

	err = png.Encode(os.Stdout, img)
	if err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return
	}
}

func renderMapToImage(elevationMap *asctools.ElevationMap) (image.Image, error) {
	imgWidth := int(elevationMap.GetWidth() / elevationMap.CellSize)
	imgHeight := int(elevationMap.GetHeight() / elevationMap.CellSize)
	img := image.NewGray16(image.Rect(0, 0, imgWidth, imgHeight))

	elevationRange := elevationMap.MaxElevation - elevationMap.MinElevation

	for y := 0.0; y < elevationMap.GetHeight(); y += elevationMap.CellSize {
		for x := 0.0; x < elevationMap.GetWidth(); x += elevationMap.CellSize {
			elevation := elevationMap.GetElevation(elevationMap.MinX+x, elevationMap.MinY+y)
			imgX := int(x / elevationMap.CellSize)
			imgY := imgHeight - 1 - int(y/elevationMap.CellSize)
			if elevation == asctools.NodataValue {
				img.SetGray16(imgX, imgY, color.Gray16{Y: 0})
			} else {
				normalized := (elevation - elevationMap.MinElevation) / elevationRange
				grayValue := uint16(normalized * math.MaxUint16)
				img.SetGray16(imgX, imgY, color.Gray16{Y: grayValue})
			}
		}
	}

	return img, nil
}

func renderMapToImageAbsolute(elevationMap *asctools.ElevationMap) (image.Image, error) {
	imgWidth := int(elevationMap.GetWidth() / elevationMap.CellSize)
	imgHeight := int(elevationMap.GetHeight() / elevationMap.CellSize)
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	for y := 0.0; y < elevationMap.GetHeight(); y += elevationMap.CellSize {
		for x := 0.0; x < elevationMap.GetWidth(); x += elevationMap.CellSize {
			elevation := elevationMap.GetElevation(elevationMap.MinX+x, elevationMap.MinY+y)
			imgX := int(x / elevationMap.CellSize)
			imgY := imgHeight - 1 - int(y/elevationMap.CellSize)

			if elevation == asctools.NodataValue {
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

	return img, nil
}
