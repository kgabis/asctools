package asctools

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"

	"golang.org/x/image/draw"
)

func (elevationMap *ElevationMap) WriteToPNG(writer *bufio.Writer, downscaleFactor int) error {
	img, err := elevationMap.renderToImage()
	if err != nil {
		return err
	}
	if downscaleFactor != 1.0 {
		newWidth := int(float64(img.Bounds().Dx()) * float64(downscaleFactor))
		newHeight := int(float64(img.Bounds().Dy()) * float64(downscaleFactor))
		scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.NearestNeighbor.Scale(scaledImg, scaledImg.Bounds(), img, img.Bounds(), draw.Over, nil)
		img = scaledImg
	}

	err = png.Encode(writer, img)
	if err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return err
	}
	return writer.Flush()
}

func (elevationMap *ElevationMap) renderToImage() (image.Image, error) {
	imgWidth := int(elevationMap.GetWidth() / elevationMap.CellSize)
	imgHeight := int(elevationMap.GetHeight() / elevationMap.CellSize)
	img := image.NewGray16(image.Rect(0, 0, imgWidth, imgHeight))

	elevationRange := elevationMap.MaxElevation - elevationMap.MinElevation

	for y := 0.0; y < elevationMap.GetHeight(); y += elevationMap.CellSize {
		for x := 0.0; x < elevationMap.GetWidth(); x += elevationMap.CellSize {
			elevation := elevationMap.GetElevation(elevationMap.MinX+x, elevationMap.MinY+y)
			imgX := int(x / elevationMap.CellSize)
			imgY := imgHeight - 1 - int(y/elevationMap.CellSize)
			if elevation == NodataValue {
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
