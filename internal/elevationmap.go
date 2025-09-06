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
	NumRows      int
	NumCols      int
	CellSize     float64
	MinX         float64
	MaxX         float64
	MinY         float64
	MaxY         float64
	Data         [][]float64
	MinElevation float64
	MaxElevation float64
}

func MergeMaps(maps []*ElevationMap) (*ElevationMap, error) {
	if len(maps) == 0 {
		return nil, fmt.Errorf("no maps to merge")
	}

	cellSize := maps[0].CellSize
	for _, m := range maps {
		if m.CellSize != cellSize {
			return nil, fmt.Errorf("incompatible cell sizes")
		}
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
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

	width := maxX - minX
	height := maxY - minY

	numRows := int(height / cellSize)
	numCols := int(width / cellSize)

	merged := &ElevationMap{
		MinX:         minX,
		MinY:         minY,
		MaxX:         maxX,
		MaxY:         maxY,
		NumRows:      numRows,
		NumCols:      numCols,
		CellSize:     cellSize,
		MinElevation: math.MaxFloat64,
		MaxElevation: -math.MaxFloat64,
		Data:         make([][]float64, numRows),
	}

	for i := range merged.Data {
		merged.Data[i] = make([]float64, numCols)
		for j := range merged.Data[i] {
			merged.Data[i][j] = NodataValue
		}
	}

	for _, m := range maps {
		if m.NumRows != len(m.Data) || m.NumCols != len(m.Data[0]) {
			return nil, fmt.Errorf("map dimensions do not match data size")
		}
		for y := m.MinY; y < m.MaxY; y += m.CellSize {
			for x := m.MinX; x < m.MaxX; x += m.CellSize {
				value := m.GetElevation(x, y)
				if value != NodataValue {
					merged.SetElevation(x, y, value)
				}
			}
		}
	}

	merged.fixHoles()

	return merged, nil
}

func (elevationMap *ElevationMap) fixHoles() {
	newData := make([][]float64, elevationMap.NumRows)
	for i := range newData {
		newData[i] = make([]float64, elevationMap.NumCols)
		copy(newData[i], elevationMap.Data[i])
	}

	for y := 1; y < elevationMap.NumRows-1; y++ {
		for x := 1; x < elevationMap.NumCols-1; x++ {
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
	scanner.Buffer(make([]byte, 64*1024), 256*1024*1024)

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
			elevationMap.NumCols, _ = strconv.Atoi(parts[1])
		case "nrows":
			elevationMap.NumRows, _ = strconv.Atoi(parts[1])
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

	elevationMap.Data = make([][]float64, elevationMap.NumRows)
	for i := range elevationMap.Data {
		elevationMap.Data[i] = make([]float64, elevationMap.NumCols)
		if !scanner.Scan() {
			return nil, fmt.Errorf("unexpected end of file at row %d", i)
		}
		row := strings.Fields(scanner.Text())
		if len(row) != elevationMap.NumCols {
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

	for i := 0; i < len(elevationMap.Data)/2; i++ {
		elevationMap.Data[i], elevationMap.Data[len(elevationMap.Data)-1-i] = elevationMap.Data[len(elevationMap.Data)-1-i], elevationMap.Data[i]
	}

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

	width := float64(elevationMap.NumCols) * elevationMap.CellSize
	height := float64(elevationMap.NumRows) * elevationMap.CellSize

	elevationMap.MinX = centerX - width/2
	elevationMap.MaxX = centerX + width/2
	elevationMap.MinY = centerY - height/2
	elevationMap.MaxY = centerY + height/2

	return &elevationMap, nil
}

func (elevationMap *ElevationMap) WriteASC(writer *bufio.Writer) error {
	header := fmt.Sprintf(
		"ncols %d\nnrows %d\nxllcenter %f\nyllcenter %f\ncellsize %f\nNODATA_value %f\n",
		elevationMap.NumCols,
		elevationMap.NumRows,
		(elevationMap.MaxX-elevationMap.MinX)/2,
		(elevationMap.MaxY-elevationMap.MinY)/2,
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

func (elevationMap *ElevationMap) GetElevation(x float64, y float64) float64 {
	if x >= elevationMap.MinX && x < elevationMap.MaxX && y >= elevationMap.MinY && y < elevationMap.MaxY {
		mapY := y - elevationMap.MinY
		mapX := x - elevationMap.MinX
		row := int(mapY / elevationMap.CellSize)
		col := int(mapX / elevationMap.CellSize)
		if row >= 0 && row < len(elevationMap.Data) && col >= 0 && col < len(elevationMap.Data[row]) {
			value := elevationMap.Data[row][col]
			return value
		}
	}

	return NodataValue
}

func (elevationMap *ElevationMap) SetElevation(x float64, y float64, value float64) {
	if x >= elevationMap.MinX && x < elevationMap.MaxX && y >= elevationMap.MinY && y < elevationMap.MaxY {
		mapY := y - elevationMap.MinY
		mapX := x - elevationMap.MinX
		row := int(mapY / elevationMap.CellSize)
		col := int(mapX / elevationMap.CellSize)
		if row >= 0 && row < len(elevationMap.Data) && col >= 0 && col < len(elevationMap.Data[row]) {
			elevationMap.Data[row][col] = value
			if value != NodataValue {
				if value < elevationMap.MinElevation {
					elevationMap.MinElevation = value
				}
				if value > elevationMap.MaxElevation {
					elevationMap.MaxElevation = value
				}
			}
		}
	}
}

func (elevationMap *ElevationMap) Split(verTiles, horTiles int, uniformSize bool) ([][]*ElevationMap, error) {
	if verTiles <= 0 || horTiles <= 0 {
		return nil, fmt.Errorf("invalid dimensions")
	}

	var tileWidth, tileHeight float64
	var usableWidth, usableHeight float64

	if uniformSize {
		tileWidth = elevationMap.GetWidth() / float64(horTiles)
		tileHeight = elevationMap.GetHeight() / float64(verTiles)

		if tileWidth < tileHeight {
			tileWidth = tileHeight
		} else {
			tileHeight = tileWidth
		}

		usableWidth = float64(horTiles) * tileWidth
		usableHeight = float64(verTiles) * tileHeight
	} else {
		tileWidth = elevationMap.GetWidth() / float64(horTiles)
		tileHeight = elevationMap.GetHeight() / float64(verTiles)
		usableWidth = elevationMap.GetWidth()
		usableHeight = elevationMap.GetHeight()
	}

	result := make([][]*ElevationMap, verTiles)
	for i := range result {
		result[i] = make([]*ElevationMap, horTiles)
	}

	for row := 0; row < verTiles; row++ {
		for col := 0; col < horTiles; col++ {
			startX := float64(col) * tileWidth
			startY := float64(row) * tileHeight
			endX := startX + tileWidth
			endY := startY + tileHeight

			if endX > usableWidth {
				endX = usableWidth
			}
			if endY > usableHeight {
				endY = usableHeight
			}

			cropStartX := elevationMap.MinX + startX
			cropStartY := elevationMap.MinY + startY
			cropEndX := elevationMap.MinX + endX
			cropEndY := elevationMap.MinY + endY

			tile, err := elevationMap.Crop(cropStartX, cropStartY, cropEndX, cropEndY)
			if err != nil {
				return nil, err
			}
			result[row][col] = tile
		}
	}

	return result, nil
}

func (elevationMap *ElevationMap) Crop(startX, startY, endX, endY float64) (*ElevationMap, error) {
	if startX > endX {
		temp := startX
		startX = endX
		endX = temp
	}

	if startY > endY {
		temp := startY
		startY = endY
		endY = temp
	}

	if startX < elevationMap.MinX || endX >= elevationMap.MaxX || startY < elevationMap.MinY || endY >= elevationMap.MaxY {
		return nil, fmt.Errorf("position out of range")
	}

	width := endX - startX
	height := endY - startY

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid crop dimensions")
	}

	numRows := int(height / elevationMap.CellSize)
	numCols := int(width / elevationMap.CellSize)

	tileData := make([][]float64, numRows)
	for i := range tileData {
		tileData[i] = make([]float64, numCols)
	}

	startRow := int(startY / elevationMap.CellSize)
	startCol := int(startX / elevationMap.CellSize)

	minElevation := math.MaxFloat64
	maxElevation := -math.MaxFloat64

	for row := 0; row < numRows; row++ {
		for col := 0; col < numCols; col++ {
			sourceRow := startRow + row
			sourceCol := startCol + col

			if sourceRow < len(elevationMap.Data) && sourceCol < len(elevationMap.Data[sourceRow]) {
				value := elevationMap.Data[sourceRow][sourceCol]
				tileData[row][col] = value

				if value != NodataValue {
					if value < minElevation {
						minElevation = value
					}
					if value > maxElevation {
						maxElevation = value
					}
				}
			} else {
				tileData[row][col] = NodataValue
			}
		}
	}

	result := &ElevationMap{
		NumRows:      numRows,
		NumCols:      numCols,
		CellSize:     elevationMap.CellSize,
		MinX:         startX,
		MaxX:         startX + width,
		MinY:         startY,
		MaxY:         startY + height,
		Data:         tileData,
		MinElevation: minElevation,
		MaxElevation: maxElevation,
	}

	return result, nil
}

func (elevationMap *ElevationMap) CropRelative(startX, startY, endX, endY float64) (*ElevationMap, error) {
	if startX > endX {
		temp := startX
		startX = endX
		endX = temp
	}

	if startY > endY {
		temp := startY
		startY = endY
		endY = temp
	}

	if startX < 0 || startX > 1 || startY < 0 || startY > 1 ||
		endX < 0 || endX > 1 || endY < 0 || endY > 1 ||
		startX >= endX || startY >= endY {
		return nil, fmt.Errorf("invalid relative coordinates")
	}

	absStartX := startX * elevationMap.GetWidth()
	absStartY := startY * elevationMap.GetHeight()
	absEndX := endX * elevationMap.GetWidth()
	absEndY := endY * elevationMap.GetHeight()

	if absStartX < 0 {
		absStartX = 0
	}
	if absStartY < 0 {
		absStartY = 0
	}
	if absEndX > elevationMap.GetWidth() {
		absEndX = elevationMap.GetWidth()
	}
	if absEndY > elevationMap.GetHeight() {
		absEndY = elevationMap.GetHeight()
	}

	return elevationMap.Crop(elevationMap.MinX+absStartX, elevationMap.MinY+absStartY, elevationMap.MinX+absEndX, elevationMap.MinY+absEndY)
}

func (elevationMap *ElevationMap) GetWidth() float64 {
	return float64(elevationMap.NumCols) * elevationMap.CellSize
}

func (elevationMap *ElevationMap) GetHeight() float64 {
	return float64(elevationMap.NumRows) * elevationMap.CellSize
}
