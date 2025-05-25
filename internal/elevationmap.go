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

type ElevationMap struct {
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

func ASCDirToElevationMap(inputDir string) (*ElevationMap, error) {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("Error reading input directory: %v", err)
	}

	var maps []ElevationMap

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".asc") {
			inputFileName := inputDir + "/" + file.Name()
			slice, err := readASCFile(inputFileName)
			if err != nil {
				return nil, fmt.Errorf("Error reading ASC file: %v", err)
			}
			maps = append(maps, slice)
		}
	}

	merged := mergeMaps(maps)

	merged.fixHoles()

	return merged, nil
}

func mergeMaps(maps []ElevationMap) *ElevationMap {
	if len(maps) == 0 {
		return nil
	}

	// Determine the bounds of the merged map
	minX, minY := math.MaxInt, math.MaxInt
	maxX, maxY := -math.MaxInt, -math.MaxInt
	for _, m := range maps {
		if m.MinX < minX {
			minX = m.MinX
		}
		if m.MinY < minY {
			minY = m.MinY
		}
		if m.MaxX > maxX {
			maxX = m.MaxX
		}
		if m.MaxY > maxY {
			maxY = m.MaxY
		}
	}

	// Calculate dimensions of the merged map
	width := maxX - minX
	height := maxY - minY

	// Initialize the merged map
	merged := &ElevationMap{
		MinX:         minX,
		MinY:         minY,
		MaxX:         maxX,
		MaxY:         maxY,
		Width:        width,
		Height:       height,
		CellSize:     maps[0].CellSize,
		MinElevation: math.MaxFloat64,
		MaxElevation: -math.MaxFloat64,
		Data:         make([][]float64, height),
	}

	for i := range merged.Data {
		merged.Data[i] = make([]float64, width)
		for j := range merged.Data[i] {
			merged.Data[i][j] = NodataValue
		}
	}

	// Merge the maps
	for _, m := range maps {
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				globalX := m.MinX + x
				globalY := m.MinY + y
				mergedX := globalX - minX
				mergedY := globalY - minY

				value := m.Data[y][x]
				if value != NodataValue {
					merged.Data[mergedY][mergedX] = value
					if value < merged.MinElevation {
						merged.MinElevation = value
					}
					if value > merged.MaxElevation {
						merged.MaxElevation = value
					}
				}
			}
		}
	}

	return merged
}

func (elevationMap *ElevationMap) fixHoles() {
	newData := make([][]float64, elevationMap.Height)
	for i := range newData {
		newData[i] = make([]float64, elevationMap.Width)
		copy(newData[i], elevationMap.Data[i])
	}

	for y := 1; y < elevationMap.Height-1; y++ {
		for x := 1; x < elevationMap.Width-1; x++ {
			if elevationMap.Data[y][x] != NodataValue {
				continue
			}
			above := elevationMap.Data[y-1][x]
			if above != NodataValue {
				newData[y][x] = above
				continue
			}
			below := elevationMap.Data[y+1][x]
			if below != NodataValue {
				newData[y][x] = below
				continue
			}
			left := elevationMap.Data[y][x-1]
			if left != NodataValue {
				newData[y][x] = left
				continue
			}
			right := elevationMap.Data[y][x+1]
			if right != NodataValue {
				newData[y][x] = right
				continue
			}
		}
	}

	elevationMap.Data = newData
}

func readASCFile(filePath string) (ElevationMap, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return ElevationMap{}, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	elevationMap := ElevationMap{
		MinX:         math.MaxInt,
		MinY:         math.MaxInt,
		MaxX:         -math.MaxInt,
		MaxY:         -math.MaxInt,
		MinElevation: math.MaxFloat32,
		MaxElevation: -math.MaxFloat32,
	}

	mapNodataValue := NodataValue
	centerX := 0.0
	centerY := 0.0
	for i := 0; i < 6; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		if len(parts) != 2 {
			return elevationMap, fmt.Errorf("invalid header line: %s", scanner.Text())
		}
		switch strings.ToLower(parts[0]) {
		case "ncols":
			elevationMap.Width, _ = strconv.Atoi(parts[1])
		case "nrows":
			elevationMap.Height, _ = strconv.Atoi(parts[1])
		case "xllcenter":
			centerX, _ = strconv.ParseFloat(parts[1], 64)
		case "yllcenter":
			centerY, _ = strconv.ParseFloat(parts[1], 64)
		case "cellsize":
			elevationMap.CellSize, _ = strconv.ParseFloat(parts[1], 64)
		case "nodata_value":
			mapNodataValue, _ = strconv.ParseFloat(parts[1], 64)
		}
	}

	// Read grid data
	elevationMap.Data = make([][]float64, elevationMap.Height)
	for i := range elevationMap.Data {
		elevationMap.Data[i] = make([]float64, elevationMap.Width)
		if !scanner.Scan() {
			return elevationMap, fmt.Errorf("unexpected end of file at row %d", i)
		}
		row := strings.Fields(scanner.Text())
		if len(row) != elevationMap.Width {
			return elevationMap, fmt.Errorf("wrong number of columns at row %d", i)
		}
		for j, v := range row {
			val, _ := strconv.ParseFloat(v, 64)
			if val == mapNodataValue {
				val = NodataValue
			}
			elevationMap.Data[i][j] = val
		}
	}

	// Flip the data vertically
	for i := 0; i < len(elevationMap.Data)/2; i++ {
		elevationMap.Data[i], elevationMap.Data[len(elevationMap.Data)-1-i] = elevationMap.Data[len(elevationMap.Data)-1-i], elevationMap.Data[i]
	}

	// Find min and max elevation values
	for _, row := range elevationMap.Data {
		for _, value := range row {
			if value != mapNodataValue {
				if value < elevationMap.MinElevation {
					elevationMap.MinElevation = value
				}
				if value > elevationMap.MaxElevation {
					elevationMap.MaxElevation = value
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ElevationMap{}, fmt.Errorf("error reading file: %v", err)
	}

	elevationMap.MinX = int(centerX - float64(elevationMap.Width)/2)
	elevationMap.MaxX = int(centerX + float64(elevationMap.Width)/2)
	elevationMap.MinY = int(centerY - float64(elevationMap.Height)/2)
	elevationMap.MaxY = int(centerY + float64(elevationMap.Height)/2)

	return elevationMap, nil
}

func (elevationMap *ElevationMap) GetElevation(x int, y int) float64 {
	if x >= elevationMap.MinX && x < elevationMap.MaxX && y >= elevationMap.MinY && y < elevationMap.MaxY {
		mapY := y - elevationMap.MinY
		mapX := x - elevationMap.MinX
		if mapY >= 0 && mapY < len(elevationMap.Data) && mapX >= 0 && mapX < len(elevationMap.Data[mapY]) {
			value := elevationMap.Data[mapY][mapX]
			return value
		}
	}

	return NodataValue
}
