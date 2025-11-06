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
		return fmt.Errorf("error encoding PNG: %v", err)
	}
	return writer.Flush()
}

func WriteDiffPNG(writer *bufio.Writer, elevationMap1 *ElevationMap, elevationMap2 *ElevationMap, diffPow float64, diffOnly bool) error {
	minX := math.Max(elevationMap1.MinX, elevationMap2.MinX)
	maxX := math.Min(elevationMap1.MaxX, elevationMap2.MaxX)
	minY := math.Max(elevationMap1.MinY, elevationMap2.MinY)
	maxY := math.Min(elevationMap1.MaxY, elevationMap2.MaxY)

	if minX >= maxX || minY >= maxY {
		return fmt.Errorf("elevation maps do not overlap")
	}

	imgWidth := int(elevationMap1.GetWidth() / elevationMap1.CellSize)
	imgHeight := int(elevationMap1.GetHeight() / elevationMap1.CellSize)

	img := image.NewRGBA64(image.Rect(0, 0, imgWidth, imgHeight))

	maxDiff := -math.MaxFloat64

	step := elevationMap1.CellSize

	for y := minY; y < maxY; y += step {
		for x := minX; x < maxX; x += step {
			elevation1 := elevationMap1.GetElevation(x, y)
			elevation2 := elevationMap2.GetElevation(x, y)
			if elevation1 != NodataValue && elevation2 != NodataValue {
				diff := math.Abs(elevation2 - elevation1)
				if diff > maxDiff {
					maxDiff = diff
				}
			}
		}
	}

	elevationRange1 := elevationMap1.MaxElevation - elevationMap1.MinElevation
	elevationRange2 := elevationMap2.MaxElevation - elevationMap2.MinElevation
	elevationRange := math.Max(elevationRange1, elevationRange2)

	for imgY := 0; imgY < imgHeight; imgY++ {
		for imgX := 0; imgX < imgWidth; imgX++ {
			flippedImgY := imgHeight - imgY - 1
			mapX := elevationMap1.MinX + float64(imgX)*elevationMap1.CellSize
			mapY := elevationMap1.MinY + float64(imgY)*elevationMap1.CellSize

			elevation1 := elevationMap1.GetElevation(mapX, mapY)
			elevation2 := elevationMap2.GetElevation(mapX, mapY)
			elevationDiff := math.Abs(elevation2 - elevation1)

			var tintColor color.RGBA64
			if elevation2 < elevation1 {
				tintColor = color.RGBA64{R: math.MaxUint16, G: 0, B: 0, A: math.MaxUint16}
			} else {
				tintColor = color.RGBA64{R: 0, G: math.MaxUint16, B: 0, A: math.MaxUint16}
			}
			if elevation1 == NodataValue || elevation2 == NodataValue {
				img.SetRGBA64(imgX, flippedImgY, color.RGBA64{R: 0, G: 0, B: 0, A: 0})
			} else {
				normalized := (elevation1 - elevationMap1.MinElevation) / elevationRange
				emphasized := math.Pow(normalized, diffPow)

				if emphasized > 1 {
					emphasized = 1
				}

				grayValue := uint16(emphasized * math.MaxUint16)
				elevationColor := color.RGBA64{R: grayValue, G: grayValue, B: grayValue, A: math.MaxUint16}
				if diffOnly {
					elevationColor = color.RGBA64{R: 0, G: 0, B: 0, A: math.MaxUint16}
				}
				interpolationFactor := elevationDiff / maxDiff

				diffColor := color.RGBA64{
					R: uint16(float64(elevationColor.R)*(1-interpolationFactor) + float64(tintColor.R)*interpolationFactor),
					G: uint16(float64(elevationColor.G)*(1-interpolationFactor) + float64(tintColor.G)*interpolationFactor),
					B: uint16(float64(elevationColor.B)*(1-interpolationFactor) + float64(tintColor.B)*interpolationFactor),
					A: math.MaxUint16,
				}
				img.SetRGBA64(imgX, flippedImgY, diffColor)
			}
		}
	}
	err := png.Encode(writer, img)
	if err != nil {
		return fmt.Errorf("error encoding PNG: %v", err)
	}

	return nil
}
