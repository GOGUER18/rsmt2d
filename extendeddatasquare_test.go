package rsmt2d

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	zeros     = bytes.Repeat([]byte{0}, shareSize)
	ones      = bytes.Repeat([]byte{1}, shareSize)
	twos      = bytes.Repeat([]byte{2}, shareSize)
	threes    = bytes.Repeat([]byte{3}, shareSize)
	fours     = bytes.Repeat([]byte{4}, shareSize)
	fives     = bytes.Repeat([]byte{5}, shareSize)
	eights    = bytes.Repeat([]byte{8}, shareSize)
	elevens   = bytes.Repeat([]byte{11}, shareSize)
	thirteens = bytes.Repeat([]byte{13}, shareSize)
	fifteens  = bytes.Repeat([]byte{15}, shareSize)
)

func TestComputeExtendedDataSquare(t *testing.T) {
	codec := NewLeoRSCodec()

	type testCase struct {
		name string
		data [][]byte
		want [][][]byte
	}
	testCases := []testCase{
		{
			name: "1x1",
			data: [][]byte{ones},
			want: [][][]byte{
				{ones, ones},
				{ones, ones},
			},
		},
		{
			name: "2x2",
			data: [][]byte{
				ones, twos,
				threes, fours,
			},
			want: [][][]byte{
				{ones, twos, zeros, threes},
				{threes, fours, eights, fifteens},
				{twos, elevens, thirteens, fours},
				{zeros, thirteens, fives, eights},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ComputeExtendedDataSquare(tc.data, codec, NewDefaultTree)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, result.squareRow)
		})
	}

	t.Run("returns an error if chunkSize is not a multiple of 64", func(t *testing.T) {
		chunk := bytes.Repeat([]byte{1}, 65)
		_, err := ComputeExtendedDataSquare([][]byte{chunk}, NewLeoRSCodec(), NewDefaultTree)
		assert.Error(t, err)
	})
}

func TestImportExtendedDataSquare(t *testing.T) {
	t.Run("is able to import an EDS", func(t *testing.T) {
		eds := createExampleEds(t, shareSize)
		got, err := ImportExtendedDataSquare(eds.Flattened(), NewLeoRSCodec(), NewDefaultTree)
		assert.NoError(t, err)
		assert.Equal(t, eds.Flattened(), got.Flattened())
	})
	t.Run("returns an error if chunkSize is not a multiple of 64", func(t *testing.T) {
		chunk := bytes.Repeat([]byte{1}, 65)
		_, err := ImportExtendedDataSquare([][]byte{chunk}, NewLeoRSCodec(), NewDefaultTree)
		assert.Error(t, err)
	})
}

func TestMarshalJSON(t *testing.T) {
	codec := NewLeoRSCodec()
	result, err := ComputeExtendedDataSquare([][]byte{
		ones, twos,
		threes, fours,
	}, codec, NewDefaultTree)
	if err != nil {
		panic(err)
	}

	edsBytes, err := json.Marshal(result)
	if err != nil {
		t.Errorf("failed to marshal EDS: %v", err)
	}

	var eds ExtendedDataSquare
	err = json.Unmarshal(edsBytes, &eds)
	if err != nil {
		t.Errorf("failed to marshal EDS: %v", err)
	}
	if !reflect.DeepEqual(result.squareRow, eds.squareRow) {
		t.Errorf("eds not equal after json marshal/unmarshal")
	}
}

func TestNewExtendedDataSquare(t *testing.T) {
	t.Run("returns an error if edsWidth is not even", func(t *testing.T) {
		edsWidth := uint(1)

		_, err := NewExtendedDataSquare(NewLeoRSCodec(), NewDefaultTree, edsWidth, shareSize)
		assert.Error(t, err)
	})
	t.Run("returns an error if chunkSize is not a multiple of 64", func(t *testing.T) {
		edsWidth := uint(1)
		chunkSize := uint(65)

		_, err := NewExtendedDataSquare(NewLeoRSCodec(), NewDefaultTree, edsWidth, chunkSize)
		assert.Error(t, err)
	})
	t.Run("returns a 4x4 EDS", func(t *testing.T) {
		edsWidth := uint(4)

		got, err := NewExtendedDataSquare(NewLeoRSCodec(), NewDefaultTree, edsWidth, shareSize)
		assert.NoError(t, err)
		assert.Equal(t, edsWidth, got.width)
		assert.Equal(t, uint(shareSize), got.chunkSize)
	})
	t.Run("returns a 4x4 EDS that can be populated via SetCell", func(t *testing.T) {
		edsWidth := uint(4)

		got, err := NewExtendedDataSquare(NewLeoRSCodec(), NewDefaultTree, edsWidth, shareSize)
		assert.NoError(t, err)

		chunk := bytes.Repeat([]byte{1}, int(shareSize))
		err = got.SetCell(0, 0, chunk)
		assert.NoError(t, err)
		assert.Equal(t, chunk, got.squareRow[0][0])
	})
	t.Run("returns an error when SetCell is invoked on an EDS with a chunk that is not the correct size", func(t *testing.T) {
		edsWidth := uint(4)
		incorrectChunkSize := shareSize + 1

		got, err := NewExtendedDataSquare(NewLeoRSCodec(), NewDefaultTree, edsWidth, shareSize)
		assert.NoError(t, err)

		chunk := bytes.Repeat([]byte{1}, int(incorrectChunkSize))
		err = got.SetCell(0, 0, chunk)
		assert.Error(t, err)
	})
}

func TestImmutableRoots(t *testing.T) {
	codec := NewLeoRSCodec()
	result, err := ComputeExtendedDataSquare([][]byte{
		ones, twos,
		threes, fours,
	}, codec, NewDefaultTree)
	if err != nil {
		panic(err)
	}

	mutatedRowRoots, err := result.RowRoots()
	assert.NoError(t, err)

	mutatedRowRoots[0][0]++ // mutate

	rowRoots, err := result.RowRoots()
	assert.NoError(t, err)

	if reflect.DeepEqual(mutatedRowRoots, rowRoots) {
		t.Errorf("Exported EDS RowRoots was mutable")
	}

	mutatedColRoots, err := result.ColRoots()
	assert.NoError(t, err)

	mutatedColRoots[0][0]++ // mutate

	colRoots, err := result.ColRoots()
	assert.NoError(t, err)

	if reflect.DeepEqual(mutatedColRoots, colRoots) {
		t.Errorf("Exported EDS ColRoots was mutable")
	}
}

func TestEDSRowColImmutable(t *testing.T) {
	codec := NewLeoRSCodec()
	result, err := ComputeExtendedDataSquare([][]byte{
		ones, twos,
		threes, fours,
	}, codec, NewDefaultTree)
	if err != nil {
		panic(err)
	}

	row := result.Row(0)
	row[0][0]++
	if reflect.DeepEqual(row, result.Row(0)) {
		t.Errorf("Exported EDS Row was mutable")
	}

	col := result.Col(0)
	col[0][0]++
	if reflect.DeepEqual(col, result.Col(0)) {
		t.Errorf("Exported EDS Col was mutable")
	}
}

// dump acts as a data dump for the benchmarks to stop the compiler from making
// unrealistic optimizations
var dump *ExtendedDataSquare

// BenchmarkExtension benchmarks extending datasquares sizes 4-128 using all
// supported codecs (encoding only)
func BenchmarkExtensionEncoding(b *testing.B) {
	chunkSize := 256
	for i := 4; i < 513; i *= 2 {
		for codecName, codec := range codecs {
			if codec.MaxChunks() < i*i {
				// Only test codecs that support this many chunks
				continue
			}

			square := genRandDS(i, chunkSize)
			b.Run(
				fmt.Sprintf("%s %dx%dx%d ODS", codecName, i, i, len(square[0])),
				func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						eds, err := ComputeExtendedDataSquare(square, codec, NewDefaultTree)
						if err != nil {
							b.Error(err)
						}
						dump = eds
					}
				},
			)
		}
	}
}

// BenchmarkExtension benchmarks extending datasquares sizes 4-128 using all
// supported codecs (both encoding and root computation)
func BenchmarkExtensionWithRoots(b *testing.B) {
	chunkSize := 256
	for i := 4; i < 513; i *= 2 {
		for codecName, codec := range codecs {
			if codec.MaxChunks() < i*i {
				// Only test codecs that support this many chunks
				continue
			}

			square := genRandDS(i, chunkSize)
			b.Run(
				fmt.Sprintf("%s %dx%dx%d ODS", codecName, i, i, len(square[0])),
				func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						eds, err := ComputeExtendedDataSquare(square, codec, NewDefaultTree)
						if err != nil {
							b.Error(err)
						}
						_, _ = eds.RowRoots()
						_, _ = eds.ColRoots()
						dump = eds
					}
				},
			)
		}
	}
}

// genRandDS make a datasquare of random data, with width describing the number
// of shares on a single side of the ds
func genRandDS(width int, chunkSize int) [][]byte {
	var ds [][]byte
	count := width * width
	for i := 0; i < count; i++ {
		share := make([]byte, chunkSize)
		_, err := rand.Read(share)
		if err != nil {
			panic(err)
		}
		ds = append(ds, share)
	}
	return ds
}

func createExampleEds(t *testing.T, chunkSize int) (eds *ExtendedDataSquare) {
	ones := bytes.Repeat([]byte{1}, chunkSize)
	twos := bytes.Repeat([]byte{2}, chunkSize)
	threes := bytes.Repeat([]byte{3}, chunkSize)
	fours := bytes.Repeat([]byte{4}, chunkSize)
	ods := [][]byte{
		ones, twos,
		threes, fours,
	}

	eds, err := ComputeExtendedDataSquare(ods, NewLeoRSCodec(), NewDefaultTree)
	require.NoError(t, err)
	return eds
}
