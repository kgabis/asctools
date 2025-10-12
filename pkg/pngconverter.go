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

type ScalingOperation int

const (
	ScaleNone ScalingOperation = iota
	ScaleDown
	ScaleUp
)

func (elevationMap *ElevationMap) WritePNG(writer *bufio.Writer, scalingOperation ScalingOperation, scale int) error {
	scaleStep := 1
	if scalingOperation == ScaleDown && scale > 1 {
		scaleStep = scale
	}
	imgWidth := int(elevationMap.GetWidth() / elevationMap.CellSize)
	imgHeight := int(elevationMap.GetHeight() / elevationMap.CellSize)

	if scalingOperation == ScaleDown && scale > 1 {
		imgWidth = imgWidth / scale
		imgHeight = imgHeight / scale
	}
	img := image.NewGray16(image.Rect(0, 0, imgWidth, imgHeight))

	elevationRange := elevationMap.MaxElevation - elevationMap.MinElevation

	imgY := imgHeight - 1
	for y := 0.0; y < elevationMap.GetHeight(); y += elevationMap.CellSize * float64(scaleStep) {
		imgX := 0
		for x := 0.0; x < elevationMap.GetWidth(); x += elevationMap.CellSize * float64(scaleStep) {
			elevation := elevationMap.GetElevation(elevationMap.MinX+x, elevationMap.MinY+y)
			if elevation == NodataValue {
				img.SetGray16(imgX, imgY, color.Gray16{Y: 0})
			} else {
				normalized := (elevation - elevationMap.MinElevation) / elevationRange
				grayValue := uint16(normalized * math.MaxUint16)
				img.SetGray16(imgX, imgY, color.Gray16{Y: grayValue})
			}
			imgX++
		}
		imgY--
	}

	if scalingOperation == ScaleUp && scale > 1 {
		newWidth := int(float64(img.Bounds().Dx()) * float64(scale))
		newHeight := int(float64(img.Bounds().Dy()) * float64(scale))
		scaledImg := image.NewGray16(image.Rect(0, 0, newWidth, newHeight))
		draw.NearestNeighbor.Scale(scaledImg, scaledImg.Bounds(), img, img.Bounds(), draw.Over, nil)
		img = scaledImg
	}

	err := png.Encode(writer, img)
	if err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return err
	}
	return writer.Flush()
}
