package main

import (
	"fmt"
	"github.com/thoas/go-funk"

	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"errors"
)

const PixelOutOfRange = -1000

func Filter(arr interface{}, predicate interface{}) interface{} {
	return funk.Filter(arr, predicate)
}
func Map(arr interface{}, mapFunc interface{}) interface{} {
	return funk.Map(arr, mapFunc)
}
func Reduce(arr, reduceFunc, acc interface{}) interface{} {
	return funk.Reduce(arr, reduceFunc, acc)
}
func Sum(arr []int) int {
	return int(Reduce(arr, func(x int, y int) int {
		return x + y
	}, 0).(float64))
}
func CopyArray(matrix [][]int) [][]int {
	duplicate := make([][]int, len(matrix))
	for i := range matrix {
		duplicate[i] = make([]int, len(matrix[i]))
		for j := 0; j < len(duplicate[i]); j++ {
			duplicate[i][j] = matrix[i][j]
		}
	}
	return duplicate
}

//--- helper ----------

func filterInRange(arr []int) []int {
	return Filter(arr, func(x int) bool {
		//fmt.Printf("%d ", x)
		return x != PixelOutOfRange
	}).([]int)
}

func makeMask(matrix [][]int, i int, j int, kernel []int) ([]int, error) {

	size := math.Sqrt(float64(len(kernel)))
	n := int(size)
	if size != math.Trunc(size) || n%2 == 0 {
		err := errors.New("mask size must be odd number of n^2")
		return nil, err
	}

	at := func(i int, j int, c int) int {
		p := getPixel(matrix, i, j)
		if p == PixelOutOfRange {
			return PixelOutOfRange
		}
		return p * kernel[c]
	}

	r := int(n) / 2
	mask := make([]int, 0)

	c := 0
	for row := i - 1; row <= i+r; row++ {
		for col := j - 1; col <= j+r; col++ {
			mask = append(mask, at(row, col, c))
			c++
		}
	}

	return mask, nil
}

func getPixel(matrix [][]int, i int, j int) int {
	if i < 0 || i >= len(matrix) {
		return PixelOutOfRange
	}
	if j < 0 || j >= len(matrix[i]) {
		return PixelOutOfRange
	}
	return matrix[i][j]
}

func applyKernelToImage(input string, output string, kernel func(matrix [][]int, i int, j int) int) {
	//existingImageFile, err := os.Open("fuji-400.jpg")
	existingImageFile, err := os.Open(input)
	if err != nil {
		// Handle error
	}
	defer existingImageFile.Close()
	im, err := jpeg.Decode(existingImageFile)
	if err != nil {
		// Handle error
	}

	b := im.Bounds()
	img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(img, img.Bounds(), im, b.Min, draw.Src)

	fmt.Printf("image size %d x %d\n", b.Dx(), b.Dy())
	fmt.Printf("total pixel %d\n", len(img.Pix))

	var gs [][]int
	for i := 0; i < b.Dy(); i++ {
		row := make([]int, b.Dx())
		gs = append(gs, row)
	}

	for i := 0; i < len(img.Pix); i += 4 {
		pixelAt := int(i / 4)
		row := pixelAt / b.Dx()
		col := pixelAt % b.Dx()

		var r = int(img.Pix[i+0])
		var g = int(img.Pix[i+1])
		var b = int(img.Pix[i+2])
		sum := r + g + b
		avg := sum / 3
		gs[row][col] = avg
	}

	fmt.Println("apply mask")
	gs = applyMask(gs, kernel)

	fmt.Println("set pixels")
	for i := 0; i < len(img.Pix); i += 4 {
		pixelAt := int(i / 4)
		row := pixelAt / b.Dx()
		col := pixelAt % b.Dx()

		img.Pix[i+0] = uint8(gs[row][col]) //R
		img.Pix[i+1] = uint8(gs[row][col]) //G
		img.Pix[i+2] = uint8(gs[row][col]) //B
		img.Pix[i+3] = 255                 //Alpha
	}

	outputFile, err := os.Create(output + ".png")
	if err != nil {
		fmt.Println("fail to create output file!")
	}

	png.Encode(outputFile, img)

	outputFile.Close()
}

func applyMask(matrix [][]int, kernel func(matrix [][]int, i int, j int) int) [][]int {
	output := CopyArray(matrix)
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < len(matrix[i]); j++ {
			output[i][j] = kernel(matrix, i, j)
		}
		//if i%100 == 0 {
		//	fmt.Printf("complete %d rows\n", i)
		//}
	}
	return output
}

func BlurKernel(matrix [][]int, i int, j int) int {
	//size := 3
	//kernel := [9]int{
	//	1, 1, 1,
	//	1, 1, 1,
	//	1, 1, 1,
	//}
	size := 11
	kernel := Map(make([]int, size*size), func(_ int) int {
		return 1
	}).([]int)

	mask, err := makeMask(matrix, i, j, kernel)
	if err != nil {
		os.Exit(1)
	}

	mask = filterInRange(mask)
	sum := Sum(mask)
	return sum / len(mask)
}

func SharpenKernel(matrix [][]int, i int, j int) int {
	kernel := [9]int{
		-1, -1, -1,
		-1, 9, -1,
		-1, -1, -1,
	}
	mask, err := makeMask(matrix, i, j, kernel[:])
	if err != nil {
		os.Exit(1)
	}

	mask = filterInRange(mask)
	sum := Sum(mask)
	return sum
}

func thresholding(matrix [][]int, i int, j int, kernel []int, threshold float64) int {
	mask, err := makeMask(matrix, i, j, kernel)
	if err != nil {
		os.Exit(1)
	}

	mask = filterInRange(mask)
	sum := Sum(mask)
	avg := sum / len(mask)
	if math.Abs(float64(avg)) > threshold {
		return 255
	} else {
		return 0
	}
}

func HorizontalEdgeKernel(matrix [][]int, i int, j int, threshold float64) int {
	kernel := [9]int{
		-1, -1, -1,
		0, 0, 0,
		1, 1, 1,
	}
	return thresholding(matrix, i, j, kernel[:], threshold)
}

func VerticalEdgeKernel(matrix [][]int, i int, j int, threshold float64) int {
	kernel := [9]int{
		-1, 0, 1,
		-1, 0, 1,
		-1, 0, 1,
	}
	return thresholding(matrix, i, j, kernel[:], threshold)
}

func EdgeDetectionKernel(matrix [][]int, i int, j int, threshold float64) int {
	return HorizontalEdgeKernel(matrix, i, j, threshold) | VerticalEdgeKernel(matrix, i, j, threshold)
}

func printSlice2D(arr [][]int) {
	for i := 0; i < len(arr); i++ {
		for j := 0; j < len(arr[i]); j++ {
			fmt.Printf("%2d ", arr[i][j])
		}
		fmt.Println()
	}
}

func main() {

	input := "fuji.jpg"

	applyKernelToImage(input, "output/blur", BlurKernel)
	applyKernelToImage(input, "output/sharpen", SharpenKernel)

	threshold := 8.0
	applyKernelToImage(input, "output/edge", func(matrix [][]int, i int, j int) int {
		return EdgeDetectionKernel(matrix, i, j, threshold)
	})
	applyKernelToImage(input, "output/edge-h", func(matrix [][]int, i int, j int) int {
		return HorizontalEdgeKernel(matrix, i, j, threshold)
	})
	applyKernelToImage(input, "output/edge-v", func(matrix [][]int, i int, j int) int {
		return VerticalEdgeKernel(matrix, i, j, threshold)
	})

	arr := [][]int{
		{0, 0, 5},
		{0, 0, 10},
		{5, 1, 10},
	}
	printSlice2D(arr)
	fmt.Println("--------")
	printSlice2D(applyMask(arr, BlurKernel))
}
