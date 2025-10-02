package lidartools

import (
	"bufio"
	"fmt"
	"math"
	"sort"
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

func makeElevationMap(minX, minY, maxX, maxY, cellSize float64) *ElevationMap {
	numRows := int((maxY - minY) / cellSize)
	numCols := int((maxX - minX) / cellSize)
	data := make([][]float64, numRows)
	for i := range data {
		data[i] = make([]float64, numCols)
		for j := range data[i] {
			data[i][j] = NodataValue
		}
	}

	return &ElevationMap{
		NumRows:      numRows,
		NumCols:      numCols,
		CellSize:     cellSize,
		MinX:         minX,
		MaxX:         maxX,
		MinY:         minY,
		MaxY:         maxY,
		Data:         data,
		MinElevation: math.MaxFloat64,
		MaxElevation: -math.MaxFloat64,
	}
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

	merged := makeElevationMap(minX, minY, maxX, maxY, cellSize)

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

	mapNodataValue := NodataValue
	centerX := 0.0
	centerY := 0.0
	numRows := 0
	numCols := 0
	cellSize := 1.0

	for i := 0; i < 6; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header line: %s", scanner.Text())
		}
		switch strings.ToLower(parts[0]) {
		case "ncols":
			numCols, _ = strconv.Atoi(parts[1])
		case "nrows":
			numRows, _ = strconv.Atoi(parts[1])
		case "yllcenter":
			centerY, _ = strconv.ParseFloat(parts[1], 64)
		case "xllcenter":
			centerX, _ = strconv.ParseFloat(parts[1], 64)
		case "cellsize":
			cellSize, _ = strconv.ParseFloat(parts[1], 64)
		case "nodata_value":
			mapNodataValue, _ = strconv.ParseFloat(parts[1], 64)
		}
	}

	width := float64(numCols) * cellSize
	height := float64(numRows) * cellSize

	minX := centerX - width/2
	maxX := centerX + width/2
	minY := centerY - height/2
	maxY := centerY + height/2

	elevationMap := makeElevationMap(minX, minY, maxX, maxY, cellSize)

	if numRows != elevationMap.NumRows || numCols != elevationMap.NumCols {
		return nil, fmt.Errorf("parsed map dimensions do not match expected dimensions, this should not happen")
	}

	for row := 0; row < elevationMap.NumRows; row++ {
		if !scanner.Scan() {
			return nil, fmt.Errorf("unexpected end of file at row %d", row)
		}
		rowData := strings.Fields(scanner.Text())
		if len(rowData) != elevationMap.NumCols {
			return nil, fmt.Errorf("wrong number of columns at row %d", row)
		}
		for col, rawVal := range rowData {
			val, _ := strconv.ParseFloat(rawVal, 64)
			if val == mapNodataValue {
				val = NodataValue
			}
			elevationMap.Data[row][col] = val
		}
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

	return elevationMap, nil
}

func (elevationMap *ElevationMap) WriteASC(writer *bufio.Writer) error {
	header := fmt.Sprintf(
		"ncols %d\nnrows %d\nxllcenter %.2f\nyllcenter %.2f\ncellsize %.2f\nnodata_value %.0f\n",
		elevationMap.NumCols,
		elevationMap.NumRows,
		elevationMap.MinX+elevationMap.GetWidth()/2,
		elevationMap.MinY+elevationMap.GetHeight()/2,
		elevationMap.CellSize,
		NodataValue,
	)
	if _, err := writer.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	for _, row := range elevationMap.Data {
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
			realRow := len(elevationMap.Data) - 1 - row
			value := elevationMap.Data[realRow][col]
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
			realRow := len(elevationMap.Data) - 1 - row
			elevationMap.Data[realRow][col] = value
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

	if startX < elevationMap.MinX || endX > elevationMap.MaxX || startY < elevationMap.MinY || endY > elevationMap.MaxY {
		return nil, fmt.Errorf("position out of range")
	}

	width := endX - startX
	height := endY - startY

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid crop dimensions")
	}

	result := makeElevationMap(startX, startY, endX, endY, elevationMap.CellSize)

	for y := startY; y < endY; y += elevationMap.CellSize {
		for x := startX; x < endX; x += elevationMap.CellSize {
			value := elevationMap.GetElevation(x, y)
			result.SetElevation(x, y, value)
		}
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

func (elevationMap *ElevationMap) Denoise(windowSize int) (*ElevationMap, error) {
	if windowSize%2 == 0 || windowSize < 3 {
		return nil, fmt.Errorf("window size must be an odd number greater than or equal to 3")
	}

	newMap := makeElevationMap(elevationMap.MinX, elevationMap.MinY, elevationMap.MaxX, elevationMap.MaxY, elevationMap.CellSize)

	halfWindow := windowSize / 2

	for y := elevationMap.MinY; y < elevationMap.MaxY; y += elevationMap.CellSize {
		for x := elevationMap.MinX; x < elevationMap.MaxX; x += elevationMap.CellSize {
			neighbours := []float64{}

			for i := -halfWindow; i <= halfWindow; i++ {
				for j := -halfWindow; j <= halfWindow; j++ {
					neighbourX := x + float64(j)*elevationMap.CellSize
					neighbourY := y + float64(i)*elevationMap.CellSize

					if neighbourX < elevationMap.MinX || neighbourX >= elevationMap.MaxX ||
						neighbourY < elevationMap.MinY || neighbourY >= elevationMap.MaxY {
						continue
					}
					value := elevationMap.GetElevation(neighbourX, neighbourY)
					if value == NodataValue {
						continue
					}
					neighbours = append(neighbours, value)
				}
			}
			newMap.SetElevation(x, y, calculateMedian(neighbours))
		}
	}

	return newMap, nil
}

func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sort.Float64s(values)

	mid := len(values) / 2
	if len(values)%2 == 0 {
		return (values[mid-1] + values[mid]) / 2.0
	}

	return values[mid]
}

func (elevationMap *ElevationMap) Downscale(factor int) (*ElevationMap, error) {
	if factor <= 1 {
		return nil, fmt.Errorf("downscale factor must be greater than 1")
	}

	newMap := makeElevationMap(elevationMap.MinX, elevationMap.MinY, elevationMap.MaxX, elevationMap.MaxY, elevationMap.CellSize*float64(factor))
	for y := newMap.MinY; y < newMap.MaxY; y += newMap.CellSize {
		for x := newMap.MinX; x < newMap.MaxX; x += newMap.CellSize {
			var sum float64
			var count int

			for subY := y; subY < y+newMap.CellSize; subY += elevationMap.CellSize {
				for subX := x; subX < x+newMap.CellSize; subX += elevationMap.CellSize {
					elevation := elevationMap.GetElevation(subX, subY)
					if elevation != NodataValue {
						sum += elevation
						count++
					}
				}
			}

			if count > 0 {
				average := sum / float64(count)
				newMap.SetElevation(x+newMap.CellSize/2, y+newMap.CellSize/2, average)
			} else {
				newMap.SetElevation(x+newMap.CellSize/2, y+newMap.CellSize/2, NodataValue)
			}
		}
	}

	return newMap, nil
}
