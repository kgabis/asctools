package lidartools

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const NodataValue = -9999.0

type MapSlice struct {
	CenterX      float64
	CenterY      float64
	Width        int
	Height       int
	CellSize     float64
	MinX         int
	MaxX         int
	MinY         int
	MaxY         int
	Data         [][]float64
	MinElevation float64
	MaxElevation float64
}

// FixHoles interpolates missing values (NodataValue) in the map slice.
func (slice *MapSlice) FixHoles() {
	for y := 0; y < slice.Height; y++ {
		for x := 0; x < slice.Width; x++ {
			if slice.Data[y][x] == NodataValue {
				slice.Data[y][x] = slice.interpolateValue(x, y)
			}
		}
	}
}

// interpolateValue calculates the average of valid neighboring cells.
func (slice *MapSlice) interpolateValue(x, y int) float64 {
	var sum float64
	var count int

	directions := []struct{ dx, dy int }{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1}, // Cardinal directions
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1}, // Diagonal directions
	}

	for _, dir := range directions {
		neighborX, neighborY := x+dir.dx, y+dir.dy
		if neighborX >= 0 && neighborX < slice.Width && neighborY >= 0 && neighborY < slice.Height {
			neighborValue := slice.Data[neighborY][neighborX]
			if neighborValue != NodataValue {
				sum += neighborValue
				count++
			}
		}
	}

	if count > 0 {
		return sum / float64(count)
	}
	return NodataValue // Return NodataValue if no valid neighbors
}

type ElevationMap struct {
	MapSlices    []MapSlice
	MinX         int
	MaxX         int
	MinY         int
	MaxY         int
	MinElevation float64
	MaxElevation float64
}

func ASCDirToElevationMap(inputDir string) (*ElevationMap, error) {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("Error reading input directory: %v", err)
	}

	result := ElevationMap{
		MinX:         math.MaxInt,
		MaxX:         -math.MaxInt,
		MinElevation: math.MaxFloat32,
		MaxElevation: -math.MaxFloat32,
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".asc") {
			inputFileName := inputDir + "/" + file.Name()
			slice, err := readASCFile(inputFileName)
			if err != nil {
				return nil, fmt.Errorf("Error reading ASC file: %v", err)
			}
			result.MapSlices = append(result.MapSlices, slice)
		}
	}

	for _, slice := range result.MapSlices {
		if slice.MinX < result.MinX {
			result.MinX = slice.MinX
		}
		if slice.MinY < result.MinY {
			result.MinY = slice.MinY
		}
		if slice.MaxX > result.MaxX {
			result.MaxX = slice.MaxX
		}
		if slice.MaxY > result.MaxY {
			result.MaxY = slice.MaxY
		}
		if slice.MinElevation < result.MinElevation {
			result.MinElevation = slice.MinElevation
		}
		if slice.MaxElevation > result.MaxElevation {
			result.MaxElevation = slice.MaxElevation
		}
	}

	// Fix holes in the elevation map
	for i := range result.MapSlices {
		result.MapSlices[i].FixHoles()
	}

	return &result, nil
}

func readASCFile(filePath string) (MapSlice, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return MapSlice{}, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	slice := MapSlice{
		MinX:         math.MaxInt,
		MinY:         math.MaxInt,
		MaxX:         -math.MaxInt,
		MaxY:         -math.MaxInt,
		MinElevation: math.MaxFloat32,
		MaxElevation: -math.MaxFloat32,
	}

	sliceNodataValue := NodataValue
	for i := 0; i < 6; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		if len(parts) != 2 {
			return slice, fmt.Errorf("invalid header line: %s", scanner.Text())
		}
		switch strings.ToLower(parts[0]) {
		case "ncols":
			slice.Width, _ = strconv.Atoi(parts[1])
		case "nrows":
			slice.Height, _ = strconv.Atoi(parts[1])
		case "xllcenter":
			slice.CenterX, _ = strconv.ParseFloat(parts[1], 64)
		case "yllcenter":
			slice.CenterY, _ = strconv.ParseFloat(parts[1], 64)
		case "cellsize":
			slice.CellSize, _ = strconv.ParseFloat(parts[1], 64)
		case "nodata_value":
			sliceNodataValue, _ = strconv.ParseFloat(parts[1], 64)
		}
	}

	// Read grid data
	slice.Data = make([][]float64, slice.Height)
	for i := range slice.Data {
		slice.Data[i] = make([]float64, slice.Width)
		if !scanner.Scan() {
			return slice, fmt.Errorf("unexpected end of file at row %d", i)
		}
		row := strings.Fields(scanner.Text())
		if len(row) != slice.Width {
			return slice, fmt.Errorf("wrong number of columns at row %d", i)
		}
		for j, v := range row {
			val, _ := strconv.ParseFloat(v, 64)
			if val == sliceNodataValue {
				val = NodataValue
			}
			slice.Data[i][j] = val
		}
	}

	// Flip the data vertically
	for i := 0; i < len(slice.Data)/2; i++ {
		slice.Data[i], slice.Data[len(slice.Data)-1-i] = slice.Data[len(slice.Data)-1-i], slice.Data[i]
	}

	// Find min and max elevation values
	for _, row := range slice.Data {
		for _, value := range row {
			if value != sliceNodataValue {
				if value < slice.MinElevation {
					slice.MinElevation = value
				}
				if value > slice.MaxElevation {
					slice.MaxElevation = value
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return MapSlice{}, fmt.Errorf("error reading file: %v", err)
	}

	slice.MinX = int(slice.CenterX - float64(slice.Width)/2)
	slice.MaxX = int(slice.CenterX + float64(slice.Width)/2)
	slice.MinY = int(slice.CenterY - float64(slice.Height)/2)
	slice.MaxY = int(slice.CenterY + float64(slice.Height)/2)

	return slice, nil
}

func (elevationMap *ElevationMap) GetElevation(x int, y int) float64 {
	for _, slice := range elevationMap.MapSlices {
		if x >= slice.MinX && x < slice.MaxX && y >= slice.MinY && y < slice.MaxY {
			sliceY := y - slice.MinY
			sliceX := x - slice.MinX
			if sliceY >= 0 && sliceY < len(slice.Data) && sliceX >= 0 && sliceX < len(slice.Data[sliceY]) {
				value := slice.Data[sliceY][sliceX]
				if value != NodataValue {
					return value
				}
			}
		}
	}
	return NodataValue
}
