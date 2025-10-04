# ASCTools

A collection of command-line tools for working with ASC (ASCII Grid) elevation map files.

## Features

- **Convert** ASC files to PNG images or STL 3D models
- **Visualize** elevation differences between two maps
- **Crop** specific regions from elevation maps
- **Merge** multiple ASC tiles into a single map
- **Split** large maps into smaller tiles
- **Denoise** elevation data using median filtering
- **Downscale** high-resolution maps to reduce file size

## Installation

### Prerequisites

- Go 1.24.0 or higher

### Build from source

```bash
go build
```

This will create an `asctools` executable in the project directory.

## Usage

ASCTools uses a command-based interface. The general format is:

```bash
asctools <command> [flags]
```

Most commands read from stdin and write to stdout, making them easy to chain together with pipes.

### Commands

#### `asc2png` - Convert ASC to PNG

Convert an ASC elevation file to a PNG image with elevation-based coloring.

```bash
asctools asc2png < input.asc > output.png

asctools asc2png -scale=2.0 -absolute_elevation < input.asc > output.png
```

**Flags:**
- `-absolute_elevation` - Encode raw elevation values in the PNG (default: false)
- `-scale` - Scale factor for the output image (default: 1.0)

#### `asc2stl` - Convert ASC to STL

Convert an ASC elevation file to an STL 3D model for 3D printing or visualization.

```bash
asctools asc2stl < input.asc > output.stl

asctools asc2stl -floor=100.0 -floor_margin=10.0 < input.asc > output.stl
```

**Flags:**
- `-scale` - Scale factor for the result (default: 1.0)
- `-floor` - Floor elevation level (default: 0.0)
- `-floor_margin` - Margin to add around the base of the model (only works when floor is not set)

#### `crop` - Crop elevation map

Extract a specific region from an elevation map.

```bash
asctools crop -input=input.asc -relative -start_x=0.25 -start_y=0.25 -end_x=0.75 -end_y=0.75 > cropped.asc

asctools crop -input=input.asc -start_x=100 -start_y=100 -end_x=500 -end_y=500 > cropped.asc

asctools crop -relative -start_x=0.5 -start_y=0.5 -end_x=1.0 -end_y=1.0 < input.asc > cropped.asc
```

**Flags:**
- `-input` - Path to input ASC file (default: stdin)
- `-relative` - Use relative coordinates 0-1 (default: false for absolute indices)
- `-start_x` - Start X coordinate (default: 0.0)
- `-start_y` - Start Y coordinate (default: 0.0)
- `-end_x` - End X coordinate (default: 1.0)
- `-end_y` - End Y coordinate (default: 1.0)

#### `diffasc2png` - Visualize elevation differences

Create a PNG visualization showing the differences between two elevation maps.

```bash
asctools diffasc2png -input1=map2012.asc -input2=map2024.asc > diff.png

# Emphasize differences with power scaling
asctools diffasc2png -input1=map2012.asc -input2=map2024.asc -diff_pow=2 > diff.png

# Skip elevation coloring
asctools diffasc2png -input1=map2012.asc -input2=map2024.asc -skip_elevation > diff.png
```

**Flags:**
- `-input1` - Path to the first ASC file (required)
- `-input2` - Path to the second ASC file (required)
- `-skip_elevation` - Skip elevation-based coloring, only use difference coloring (default: false)
- `-diff_pow` - Power to raise elevation differences for emphasis (default: 1)

#### `merge` - Merge multiple ASC files

Merge multiple ASC tiles from a directory into a single elevation map.

```bash
asctools merge -input_dir=./tiles > merged.asc
```

**Flags:**
- `-input_dir` - Directory containing ASC files to merge (required)

#### `split` - Split ASC into tiles

Split a large ASC file into smaller tiles.

```bash
# Split into 2x2 grid
asctools split -output_dir=./tiles -nrows=2 -ncols=2 -prefix=tile < input.asc

# Split into uniform-sized tiles
asctools split -output_dir=./tiles -nrows=3 -ncols=3 -uniform -prefix=section < input.asc
```

**Flags:**
- `-output_dir` - Directory to save split files (default: ".")
- `-nrows` - Number of rows in the output grid (default: 2)
- `-ncols` - Number of columns in the output grid (default: 2)
- `-uniform` - Make all tiles the same size, discarding extra space (default: false)
- `-prefix` - Prefix for output filenames (default: "tile")

#### `denoise` - Apply median filtering

Remove noise from elevation data using median filtering.

```bash
asctools denoise -window=5 < input.asc > denoised.asc
```

**Flags:**
- `-window` - Window size for median filtering, must be odd (default: 3)

#### `downscale` - Reduce resolution

Downscale an elevation map to reduce its resolution.

```bash
asctools downscale -factor=2 < input.asc > downscaled.asc
```

**Flags:**
- `-factor` - Downscale factor, must be greater than 1 (default: 1)

## Examples

### Complete workflow

```bash
# 1. Crop a region of interest
asctools crop -input=krakow2023.asc -relative -start_x=0.2 -start_y=0.2 -end_x=0.8 -end_y=0.8 > cropped.asc

# 2. Denoise the cropped data
asctools denoise -window=5 < cropped.asc > denoised.asc

# 3. Convert to PNG for visualization
asctools asc2png -scale=2.0 < denoised.asc > visualization.png

# 4. Create a 3D model
asctools asc2stl -floor_margin=10.0 < denoised.asc > model.stl
```

### Compare two time periods

```bash
# Create a difference visualization between 2012 and 2024 maps
asctools diffasc2png -input1=wisla2010.asc -input2=wisla2024.asc -diff_pow=2 > wisla_changes.png
```

## File Format

ASCTools works with ASC (ASCII Grid) format files, which are commonly used for digital elevation models. The format consists of a header followed by elevation data:

```
ncols         1000
nrows         1000
xllcorner     0.0
yllcorner     0.0
cellsize      1.0
NODATA_value  -9999
100.5 100.7 100.9 ...
101.2 101.4 101.6 ...
...
```

## Contributing

Please don't.

## License

MIT