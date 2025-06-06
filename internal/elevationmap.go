package lidartools

import (
	"bufio"
	"fmt"
	"math"
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

func MergeMaps(maps []*ElevationMap) *ElevationMap {
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

	merged.fixHoles()

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

func ParseASCFile(reader *bufio.Reader) (*ElevationMap, error) {
	scanner := bufio.NewScanner(reader)

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
			return nil, fmt.Errorf("invalid header line: %s", scanner.Text())
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
			return nil, fmt.Errorf("unexpected end of file at row %d", i)
		}
		row := strings.Fields(scanner.Text())
		if len(row) != elevationMap.Width {
			return nil, fmt.Errorf("wrong number of columns at row %d", i)
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
		return nil, fmt.Errorf("error reading data: %v", err)
	}

	elevationMap.MinX = int(centerX - float64(elevationMap.Width)/2)
	elevationMap.MaxX = int(centerX + float64(elevationMap.Width)/2)
	elevationMap.MinY = int(centerY - float64(elevationMap.Height)/2)
	elevationMap.MaxY = int(centerY + float64(elevationMap.Height)/2)

	return &elevationMap, nil
}

func (elevationMap *ElevationMap) WriteASC(writer *bufio.Writer) error {
	header := fmt.Sprintf(
		"ncols %d\nnrows %d\nxllcenter %f\nyllcenter %f\ncellsize %f\nNODATA_value %f\n",
		elevationMap.Width,
		elevationMap.Height,
		float64(elevationMap.MinX)+float64(elevationMap.Width)/2,
		float64(elevationMap.MinY)+float64(elevationMap.Height)/2,
		elevationMap.CellSize,
		NodataValue,
	)
	if _, err := writer.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	for i := len(elevationMap.Data) - 1; i >= 0; i-- {
		row := elevationMap.Data[i]
		values := make([]string, len(row))
		for j, v := range row {
			values[j] = strconv.FormatFloat(v, 'f', -1, 64)
		}
		line := strings.Join(values, " ") + "\n"
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("failed to write data row: %v", err)
		}
	}

	return writer.Flush()
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

// Split splits the elevation map into a grid of smaller maps
// nrows and ncols specify the number of rows and columns in the output grid
// If uniformSize is true, all tiles will have the same dimensions (smaller of width/ncols and height/nrows)
func (elevationMap *ElevationMap) Split(nrows, ncols int, uniformSize bool) [][]*ElevationMap {
	if nrows <= 0 || ncols <= 0 {
		return nil
	}

	var tileWidth, tileHeight int
	var usableWidth, usableHeight int

	if uniformSize {
		// Calculate uniform tile dimensions using the smaller of the two ratios
		widthPerCol := elevationMap.Width / ncols
		heightPerRow := elevationMap.Height / nrows

		// Use the smaller dimension to ensure all tiles fit
		if widthPerCol < heightPerRow {
			tileWidth = widthPerCol
			tileHeight = widthPerCol
		} else {
			tileWidth = heightPerRow
			tileHeight = heightPerRow
		}

		// Calculate usable dimensions (may be smaller than original map)
		usableWidth = ncols * tileWidth
		usableHeight = nrows * tileHeight
	} else {
		// Calculate tile dimensions normally
		tileWidth = elevationMap.Width / ncols
		tileHeight = elevationMap.Height / nrows
		usableWidth = elevationMap.Width
		usableHeight = elevationMap.Height
	}

	// Create result grid
	result := make([][]*ElevationMap, nrows)
	for i := range result {
		result[i] = make([]*ElevationMap, ncols)
	}

	// Split the map into tiles
	for row := 0; row < nrows; row++ {
		for col := 0; col < ncols; col++ {
			// Calculate tile boundaries
			startX := col * tileWidth
			startY := row * tileHeight
			endX := startX + tileWidth
			endY := startY + tileHeight

			// Ensure we don't go beyond the usable map boundaries
			if endX > usableWidth {
				endX = usableWidth
			}
			if endY > usableHeight {
				endY = usableHeight
			}

			actualWidth := endX - startX
			actualHeight := endY - startY

			// Skip empty tiles
			if actualWidth <= 0 || actualHeight <= 0 {
				continue
			}

			// Create the tile
			result[row][col] = elevationMap.Crop(startX, startY, actualWidth, actualHeight)
		}
	}

	return result
}

// Crop creates a new ElevationMap from a rectangular region of this map
// startX, startY specify the top-left corner in absolute cell coordinates
// width, height specify the dimensions of the cropped region
func (elevationMap *ElevationMap) Crop(startX, startY, width, height int) *ElevationMap {
	// Create new tile data
	tileData := make([][]float64, height)
	for i := range tileData {
		tileData[i] = make([]float64, width)
	}

	// Copy data from original map to tile
	minElevation := math.MaxFloat64
	maxElevation := -math.MaxFloat64

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sourceY := startY + y
			sourceX := startX + x

			if sourceY < len(elevationMap.Data) && sourceX < len(elevationMap.Data[sourceY]) {
				value := elevationMap.Data[sourceY][sourceX]
				tileData[y][x] = value

				if value != NodataValue {
					if value < minElevation {
						minElevation = value
					}
					if value > maxElevation {
						maxElevation = value
					}
				}
			} else {
				tileData[y][x] = NodataValue
			}
		}
	}

	// Calculate tile center coordinates
	tileCenterX := float64(elevationMap.MinX) + float64(startX) + float64(width)/2
	tileCenterY := float64(elevationMap.MinY) + float64(startY) + float64(height)/2

	// Create the tile elevation map
	tile := &ElevationMap{
		Width:        width,
		Height:       height,
		CellSize:     elevationMap.CellSize,
		MinX:         int(tileCenterX - float64(width)/2),
		MaxX:         int(tileCenterX + float64(width)/2),
		MinY:         int(tileCenterY - float64(height)/2),
		MaxY:         int(tileCenterY + float64(height)/2),
		Data:         tileData,
		MinElevation: minElevation,
		MaxElevation: maxElevation,
	}

	return tile
}

// CropRelative creates a new ElevationMap from a rectangular region using relative coordinates (0-1)
// startX, startY, endX, endY are all in the range [0, 1] where 0,0 is top-left and 1,1 is bottom-right
func (elevationMap *ElevationMap) CropRelative(startX, startY, endX, endY float64) *ElevationMap {
	// Validate input ranges
	if startX < 0 || startX > 1 || startY < 0 || startY > 1 ||
		endX < 0 || endX > 1 || endY < 0 || endY > 1 ||
		startX >= endX || startY >= endY {
		return nil
	}

	// Convert relative coordinates to absolute cell coordinates
	absStartX := int(startX * float64(elevationMap.Width))
	absStartY := int(startY * float64(elevationMap.Height))
	absEndX := int(endX * float64(elevationMap.Width))
	absEndY := int(endY * float64(elevationMap.Height))

	// Calculate dimensions
	width := absEndX - absStartX
	height := absEndY - absStartY

	// Ensure we don't go beyond map boundaries
	if absStartX < 0 {
		absStartX = 0
	}
	if absStartY < 0 {
		absStartY = 0
	}
	if absEndX > elevationMap.Width {
		absEndX = elevationMap.Width
		width = absEndX - absStartX
	}
	if absEndY > elevationMap.Height {
		absEndY = elevationMap.Height
		height = absEndY - absStartY
	}

	if width <= 0 || height <= 0 {
		return nil
	}

	return elevationMap.Crop(absStartX, absStartY, width, height)
}
