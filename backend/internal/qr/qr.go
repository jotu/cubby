package qr

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

// Generate creates a QR code PNG for the given content string.
// size controls the pixel dimensions of the output image.
func Generate(content string, size int) ([]byte, error) {
	modules, err := encode(content)
	if err != nil {
		return nil, fmt.Errorf("encode qr: %w", err)
	}

	n := len(modules)
	quiet := 4 // quiet zone
	total := n + quiet*2
	scale := size / total
	if scale < 1 {
		scale = 1
	}
	imgSize := total * scale

	img := image.NewGray(image.Rect(0, 0, imgSize, imgSize))
	// Fill white
	for y := range imgSize {
		for x := range imgSize {
			img.SetGray(x, y, color.Gray{Y: 255})
		}
	}
	// Draw modules
	for row := range n {
		for col := range n {
			if modules[row][col] {
				for dy := range scale {
					for dx := range scale {
						px := (col+quiet)*scale + dx
						py := (row+quiet)*scale + dy
						img.SetGray(px, py, color.Gray{Y: 0})
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// --- Minimal QR Code encoder (Version 1-4, Byte mode, ECC-L) ---

// encode produces a QR code module grid for the given string.
// Supports up to ~78 characters (Version 4-L byte mode).
func encode(data string) ([][]bool, error) {
	dataBytes := []byte(data)
	version, err := pickVersion(len(dataBytes))
	if err != nil {
		return nil, err
	}

	vi := versionInfo[version]
	bits := encodeBits(dataBytes, vi)
	codewords := bitsToBytes(bits, vi.totalDataCodewords)
	ecWords := generateEC(codewords, vi.ecBytesPerBlock)
	finalBits := interleave(codewords, ecWords, vi)

	size := 17 + version*4
	grid := makeGrid(size)
	mask := makeGrid(size)

	placePatterns(grid, mask, size, version)
	placeData(grid, mask, size, finalBits)
	bestMask := applyBestMask(grid, mask, size)
	placeFormatInfo(grid, size, bestMask)

	if version >= 7 {
		placeVersionInfo(grid, size, version)
	}

	return grid, nil
}

type versionInfoT struct {
	totalDataCodewords int
	ecBytesPerBlock    int
	numBlocks          int
	capacity           int // byte mode capacity at ECC-L
}

var versionInfo = map[int]versionInfoT{
	1: {totalDataCodewords: 19, ecBytesPerBlock: 7, numBlocks: 1, capacity: 17},
	2: {totalDataCodewords: 34, ecBytesPerBlock: 10, numBlocks: 1, capacity: 32},
	3: {totalDataCodewords: 55, ecBytesPerBlock: 15, numBlocks: 1, capacity: 53},
	4: {totalDataCodewords: 80, ecBytesPerBlock: 20, numBlocks: 1, capacity: 78},
}

func pickVersion(dataLen int) (int, error) {
	for v := 1; v <= 4; v++ {
		if dataLen <= versionInfo[v].capacity {
			return v, nil
		}
	}
	return 0, fmt.Errorf("data too long for QR version 1-4 (%d bytes)", dataLen)
}

func encodeBits(data []byte, vi versionInfoT) []bool {
	var bits []bool

	// Mode indicator: byte mode = 0100
	bits = append(bits, false, true, false, false)

	// Character count (8 bits for version 1-4 byte mode)
	count := len(data)
	for i := 7; i >= 0; i-- {
		bits = append(bits, (count>>uint(i))&1 == 1)
	}

	// Data
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>uint(i))&1 == 1)
		}
	}

	// Terminator (up to 4 zeros)
	totalBits := vi.totalDataCodewords * 8
	for range 4 {
		if len(bits) >= totalBits {
			break
		}
		bits = append(bits, false)
	}

	return bits
}

func bitsToBytes(bits []bool, totalCodewords int) []byte {
	// Pad to byte boundary
	for len(bits)%8 != 0 {
		bits = append(bits, false)
	}

	result := make([]byte, 0, totalCodewords)
	for i := 0; i < len(bits); i += 8 {
		var b byte
		for j := range 8 {
			if i+j < len(bits) && bits[i+j] {
				b |= 1 << uint(7-j)
			}
		}
		result = append(result, b)
	}

	// Pad codewords
	padBytes := []byte{0xEC, 0x11}
	idx := 0
	for len(result) < totalCodewords {
		result = append(result, padBytes[idx%2])
		idx++
	}

	return result
}

// --- Reed-Solomon EC ---

var (
	gfExp [512]byte
	gfLog [256]byte
)

func init() {
	x := 1
	for i := range 255 {
		gfExp[i] = byte(x)
		gfLog[x] = byte(i)
		x <<= 1
		if x >= 256 {
			x ^= 0x11D
		}
	}
	for i := 255; i < 512; i++ {
		gfExp[i] = gfExp[i-255]
	}
}

func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[int(gfLog[a])+int(gfLog[b])]
}

func generateEC(data []byte, ecLen int) []byte {
	gen := make([]byte, ecLen+1)
	gen[0] = 1
	for i := range ecLen {
		for j := i + 1; j >= 1; j-- {
			gen[j] = gen[j-1] ^ gfMul(gen[j], gfExp[i])
		}
		gen[0] = gfMul(gen[0], gfExp[i])
	}

	result := make([]byte, len(data)+ecLen)
	copy(result, data)
	for i := range len(data) {
		coef := result[i]
		if coef == 0 {
			continue
		}
		for j := range ecLen + 1 {
			result[i+j] ^= gfMul(gen[j], coef)
		}
	}

	return result[len(data):]
}

func interleave(data, ec []byte, vi versionInfoT) []bool {
	// For version 1-4 with 1 block, just concatenate
	all := make([]byte, 0, len(data)+len(ec))
	all = append(all, data...)
	all = append(all, ec...)

	var bits []bool
	for _, b := range all {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>uint(i))&1 == 1)
		}
	}

	// Add remainder bits
	size := 17 + pickVersionFromInfo(vi)*4
	totalModules := size*size - countFunctionModules(size)
	for len(bits) < totalModules {
		bits = append(bits, false)
	}

	return bits
}

func pickVersionFromInfo(vi versionInfoT) int {
	for v, info := range versionInfo {
		if info.totalDataCodewords == vi.totalDataCodewords {
			return v
		}
	}
	return 1
}

func countFunctionModules(size int) int {
	// Approximate: finder patterns + timing + format + dark module
	count := 3*8*8 + // 3 finder patterns with separators
		2*(size-16) + // timing patterns
		31 + 1 // format info + dark module
	if size >= 25 { // version 2+, alignment pattern
		count += 25
	}
	return count
}

func makeGrid(size int) [][]bool {
	grid := make([][]bool, size)
	for i := range grid {
		grid[i] = make([]bool, size)
	}
	return grid
}

func placePatterns(grid, mask [][]bool, size, version int) {
	// Finder patterns
	placeFinder(grid, mask, 0, 0)
	placeFinder(grid, mask, size-7, 0)
	placeFinder(grid, mask, 0, size-7)

	// Separators
	for i := range 8 {
		setFunction(grid, mask, 7, i, false)
		setFunction(grid, mask, i, 7, false)
		setFunction(grid, mask, size-8, i, false)
		setFunction(grid, mask, size-1-i, 7, false)
		setFunction(grid, mask, 7, size-1-i, false)
		setFunction(grid, mask, i, size-8, false)
	}

	// Timing patterns
	for i := 8; i < size-8; i++ {
		val := i%2 == 0
		setFunction(grid, mask, 6, i, val)
		setFunction(grid, mask, i, 6, val)
	}

	// Alignment pattern (version 2+)
	if version >= 2 {
		centers := alignmentPositions(version)
		for _, r := range centers {
			for _, c := range centers {
				if isFinderArea(r, c, size) {
					continue
				}
				placeAlignment(grid, mask, r, c)
			}
		}
	}

	// Dark module
	setFunction(grid, mask, 8, 4*version+9, true)

	// Reserve format info areas
	for i := range 9 {
		setReserved(mask, 8, i)
		setReserved(mask, i, 8)
	}
	for i := range 8 {
		setReserved(mask, size-1-i, 8)
		setReserved(mask, 8, size-1-i)
	}
}

func placeFinder(grid, mask [][]bool, row, col int) {
	for r := range 7 {
		for c := range 7 {
			val := r == 0 || r == 6 || c == 0 || c == 6 ||
				(r >= 2 && r <= 4 && c >= 2 && c <= 4)
			setFunction(grid, mask, row+r, col+c, val)
		}
	}
}

func placeAlignment(grid, mask [][]bool, row, col int) {
	for r := -2; r <= 2; r++ {
		for c := -2; c <= 2; c++ {
			val := r == -2 || r == 2 || c == -2 || c == 2 || (r == 0 && c == 0)
			setFunction(grid, mask, row+r, col+c, val)
		}
	}
}

func isFinderArea(r, c, size int) bool {
	return (r < 9 && c < 9) || (r < 9 && c > size-9) || (r > size-9 && c < 9)
}

func alignmentPositions(version int) []int {
	switch version {
	case 2:
		return []int{6, 18}
	case 3:
		return []int{6, 22}
	case 4:
		return []int{6, 26}
	default:
		return nil
	}
}

func setFunction(grid, mask [][]bool, row, col int, val bool) {
	grid[row][col] = val
	mask[row][col] = true
}

func setReserved(mask [][]bool, row, col int) {
	mask[row][col] = true
}

func placeData(grid, mask [][]bool, size int, bits []bool) {
	idx := 0
	up := true
	for col := size - 1; col >= 0; col -= 2 {
		if col == 6 {
			col = 5 // Skip timing column
		}
		rows := make([]int, size)
		if up {
			for i := range size {
				rows[i] = size - 1 - i
			}
		} else {
			for i := range size {
				rows[i] = i
			}
		}
		for _, row := range rows {
			for _, dc := range []int{0, -1} {
				c := col + dc
				if c < 0 || mask[row][c] {
					continue
				}
				if idx < len(bits) {
					grid[row][c] = bits[idx]
					idx++
				}
			}
		}
		up = !up
	}
}

func applyBestMask(grid, funcMask [][]bool, size int) int {
	bestMask := 0
	bestPenalty := 1<<31 - 1

	for m := range 8 {
		candidate := copyGrid(grid, size)
		applyMask(candidate, funcMask, size, m)
		penalty := calcPenalty(candidate, size)
		if penalty < bestPenalty {
			bestPenalty = penalty
			bestMask = m
		}
	}

	applyMask(grid, funcMask, size, bestMask)
	return bestMask
}

func applyMask(grid, funcMask [][]bool, size, maskNum int) {
	for r := range size {
		for c := range size {
			if funcMask[r][c] {
				continue
			}
			var invert bool
			switch maskNum {
			case 0:
				invert = (r+c)%2 == 0
			case 1:
				invert = r%2 == 0
			case 2:
				invert = c%3 == 0
			case 3:
				invert = (r+c)%3 == 0
			case 4:
				invert = (r/2+c/3)%2 == 0
			case 5:
				invert = (r*c)%2+(r*c)%3 == 0
			case 6:
				invert = ((r*c)%2+(r*c)%3)%2 == 0
			case 7:
				invert = ((r+c)%2+(r*c)%3)%2 == 0
			}
			if invert {
				grid[r][c] = !grid[r][c]
			}
		}
	}
}

func calcPenalty(grid [][]bool, size int) int {
	penalty := 0

	// Rule 1: runs of 5+ same color
	for r := range size {
		count := 1
		for c := 1; c < size; c++ {
			if grid[r][c] == grid[r][c-1] {
				count++
			} else {
				if count >= 5 {
					penalty += count - 2
				}
				count = 1
			}
		}
		if count >= 5 {
			penalty += count - 2
		}
	}
	for c := range size {
		count := 1
		for r := 1; r < size; r++ {
			if grid[r][c] == grid[r-1][c] {
				count++
			} else {
				if count >= 5 {
					penalty += count - 2
				}
				count = 1
			}
		}
		if count >= 5 {
			penalty += count - 2
		}
	}

	// Rule 2: 2x2 blocks
	for r := range size - 1 {
		for c := range size - 1 {
			val := grid[r][c]
			if grid[r][c+1] == val && grid[r+1][c] == val && grid[r+1][c+1] == val {
				penalty += 3
			}
		}
	}

	return penalty
}

func copyGrid(grid [][]bool, size int) [][]bool {
	cp := make([][]bool, size)
	for i := range size {
		cp[i] = make([]bool, size)
		copy(cp[i], grid[i])
	}
	return cp
}

func placeFormatInfo(grid [][]bool, size, maskNum int) {
	// ECC Level L = 01, mask pattern
	data := 1<<3 | maskNum // 01 xxx
	// BCH(15,5) encoding
	bits := bchEncode(data)

	// XOR with mask pattern 101010000010010
	bits ^= 0x5412

	coords0 := [][2]int{
		{0, 8},
		{1, 8},
		{2, 8},
		{3, 8},
		{4, 8},
		{5, 8},
		{7, 8},
		{8, 8},
		{8, 7},
		{8, 5},
		{8, 4},
		{8, 3},
		{8, 2},
		{8, 1},
		{8, 0},
	}
	coords1 := [][2]int{
		{8, size - 1},
		{8, size - 2},
		{8, size - 3},
		{8, size - 4},
		{8, size - 5},
		{8, size - 6},
		{8, size - 7},
		{size - 8, 8},
		{size - 7, 8},
		{size - 6, 8},
		{size - 5, 8},
		{size - 4, 8},
		{size - 3, 8},
		{size - 2, 8},
		{size - 1, 8},
	}

	for i := range 15 {
		val := (bits>>(14-i))&1 == 1
		grid[coords0[i][0]][coords0[i][1]] = val
		grid[coords1[i][0]][coords1[i][1]] = val
	}
}

func bchEncode(data int) int {
	d := data << 10
	gen := 0x537 // generator polynomial for BCH(15,5)
	for i := 4; i >= 0; i-- {
		if d&(1<<(i+10)) != 0 {
			d ^= gen << i
		}
	}
	return (data << 10) | d
}

func placeVersionInfo(grid [][]bool, size, version int) {
	// Version info for versions 7+; not needed for 1-4 but included for completeness
	d := version << 12
	gen := 0x1F25
	for i := 5; i >= 0; i-- {
		if d&(1<<(i+12)) != 0 {
			d ^= gen << i
		}
	}
	bits := (version << 12) | d

	for i := range 18 {
		val := (bits>>i)&1 == 1
		row := i / 3
		col := size - 11 + i%3
		grid[row][col] = val
		grid[col][row] = val
	}
}
