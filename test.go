package main

import (
	"fmt"
	"github.com/thoas/go-funk"
	"image"
	"image/jpeg"
	"image/png"
	"os"
)

func test0() {
	fmt.Println("Hello, Go")

	r := funk.Map([]int{1, 2, 3, 4}, func(x int) int {
		return x * 2
	})
	fmt.Println(r)
}

func test1() {

	// Create a blank image 10 pixels wide by 4 pixels tall
	img := image.NewRGBA(image.Rect(0, 0, 10, 4))

	// You can access the pixels through myImage.Pix[i]
	// One pixel takes up four bytes/uint8. One for each: RGBA
	// So the first pixel is controlled by the first 4 elements
	// Values for color are 0 black - 255 full color
	// Alpha value is 0 transparent - 255 opaque
	img.Pix[0] = 255 // 1st pixel red
	img.Pix[1] = 0   // 1st pixel green
	img.Pix[2] = 0   // 1st pixel blue
	img.Pix[3] = 255 // 1st pixel alpha

	// myImage.Pix contains all the pixels
	// in a one-dimensional slice
	fmt.Println(img.Pix)

	// Stride is how many bytes take up 1 row of the image
	// Since 4 bytes are used for each pixel, the stride is
	// equal to 4 times the width of the image
	// Since all the pixels are stored in a 1D slice,
	// we need this to calculate where pixels are on different rows.
	fmt.Println(img.Stride) // 40 for an image 10 pixels wide

	outputFile, err := os.Create("test.png")
	if err != nil {
		// Handle error
	}

	// Encode takes a writer interface and an image interface
	// We pass it the File and the RGBA
	png.Encode(outputFile, img)

	// Don't forget to close files
	outputFile.Close()
}

func test2() {
	existingImageFile, err := os.Open("fuji-400.jpg")
	if err != nil {
		// Handle error
	}
	defer existingImageFile.Close()

	// Calling the generic image.Decode() will tell give us the data
	// and type of image it is as a string. We expect "png"
	imageData, imageType, err := image.Decode(existingImageFile)
	if err != nil {
		// Handle error
	}
	fmt.Println(imageData)
	fmt.Println(imageType)

	// We only need this because we already read from the file
	// We have to reset the file pointer back to beginning
	existingImageFile.Seek(0, 0)

	// Alternatively, since we know it is a png already
	// we can call png.Decode() directly
	loadedImage, err := jpeg.Decode(existingImageFile)
	if err != nil {
		// Handle error
	}
	fmt.Println(loadedImage)
}

type Matrix [...][...]int
