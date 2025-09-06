package diffasc2png

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
	fs := flag.NewFlagSet("diffasc2png", flag.ExitOnError)

	var input1 string
	fs.StringVar(&input1, "input1", "", "Path to the input 1 .asc file")

	var input2 string
	fs.StringVar(&input2, "input2", "", "Path to the input 2 .asc file")

	var skipElevation bool
	fs.BoolVar(&skipElevation, "skip_elevation", false, "If true, skips elevation-based coloring and only uses difference-based coloring")

	var diffPow float64
	fs.Float64Var(&diffPow, "diff_pow", 1, "Power to which the elevation difference is raised for emphasis")

	fs.Parse(args)

	if input1 == "" || input2 == "" {
		fs.Usage()
		return
	}

	file1, err := os.Open(input1)
	if err != nil {
		fmt.Println("Error opening input file 1:", err)
		return
	}
	defer file1.Close()

	file2, err := os.Open(input2)
	if err != nil {
		fmt.Println("Error opening input file 2:", err)
		return
	}
	defer file2.Close()

	elevationMap1, err := lidartools.ParseASCFile(bufio.NewReader(file1))
	if err != nil {
		fmt.Println("Error reading elevation map:", err)
		return
	}

	elevationMap2, err := lidartools.ParseASCFile(bufio.NewReader(file2))
	if err != nil {
		fmt.Println("Error reading elevation map:", err)
		return
	}

	img, err := renderMapDiffToImage(elevationMap1, elevationMap2, diffPow, skipElevation)
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

func renderMapDiffToImage(elevationMap1 *lidartools.ElevationMap, elevationMap2 *lidartools.ElevationMap, diffPow float64, skipElevation bool) (image.Image, error) {
	minX := math.Max(elevationMap1.MinX, elevationMap2.MinX)
	maxX := math.Min(elevationMap1.MaxX, elevationMap2.MaxX)
	minY := math.Max(elevationMap1.MinY, elevationMap2.MinY)
	maxY := math.Min(elevationMap1.MaxY, elevationMap2.MaxY)

	imgWidth := int(elevationMap1.GetWidth() / elevationMap1.CellSize)
	imgHeight := int(elevationMap1.GetHeight() / elevationMap1.CellSize)

	img := image.NewRGBA64(image.Rect(0, 0, imgWidth, imgHeight))

	maxDiff := -math.MaxFloat64

	step := elevationMap1.CellSize

	for y := minY; y < maxY; y += step {
		for x := minX; x < maxX; x += step {
			elevation1 := elevationMap1.GetElevation(x, y)
			elevation2 := elevationMap2.GetElevation(x, y)
			if elevation1 != lidartools.NodataValue && elevation2 != lidartools.NodataValue {
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
			if elevation1 == lidartools.NodataValue || elevation2 == lidartools.NodataValue {
				img.SetRGBA64(imgX, imgY, color.RGBA64{R: 0, G: 0, B: 0, A: 0})
			} else {
				normalized := (elevation1 - elevationMap1.MinElevation) / elevationRange
				emphasized := math.Pow(normalized, diffPow)

				if emphasized > 1 {
					emphasized = 1
				}

				grayValue := uint16(emphasized * math.MaxUint16)
				elevationColor := color.RGBA64{R: grayValue, G: grayValue, B: grayValue, A: math.MaxUint16}
				if skipElevation {
					elevationColor = color.RGBA64{R: 0, G: 0, B: 0, A: math.MaxUint16}
				}
				interpolationFactor := elevationDiff / maxDiff

				diffColor := color.RGBA64{
					R: uint16(float64(elevationColor.R)*(1-interpolationFactor) + float64(tintColor.R)*interpolationFactor),
					G: uint16(float64(elevationColor.G)*(1-interpolationFactor) + float64(tintColor.G)*interpolationFactor),
					B: uint16(float64(elevationColor.B)*(1-interpolationFactor) + float64(tintColor.B)*interpolationFactor),
					A: math.MaxUint16,
				}
				img.SetRGBA64(imgX, imgY, diffColor)
			}
		}
	}

	// Flip the image vertically
	for imgY := 0; imgY < imgHeight/2; imgY++ {
		for imgX := 0; imgX < imgWidth; imgX++ {
			top := img.RGBA64At(imgX, imgY)
			bottom := img.RGBA64At(imgX, imgHeight-imgY-1)
			img.SetRGBA64(imgX, imgY, bottom)
			img.SetRGBA64(imgX, imgHeight-imgY-1, top)
		}
	}

	return img, nil
}
