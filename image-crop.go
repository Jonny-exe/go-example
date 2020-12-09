//usr/bin/env go run $0 $@ ; exit

/*

Example Go program to demonstrate how to crop an image in memory.
Also demonstrates how to decode and encode image to base64 and PNG.

Steps:
- read image from file (for testing)
- encode image in base64
- this is what the server would receive
- decode base64 image
- decode binary image to Image object
- crop Image
- encode cropped Image as PNG binary
- write cropped PNG binary to file (for testing)
- encode cropped PNG binary as base64
- this is what server would send to MongoDB

*/

package main

import "bytes"
import "encoding/base64"
import "fmt"
import "image"
import "image/png"
import "io"
import "io/ioutil"

// to install: "go get github.com/anthonynsimon/bild/..."
// installed it into: /home/a/go/src/github.com/anthonynsimon
import "github.com/anthonynsimon/bild/transform"

// Declaring some const variables
const imagePath = "/home/a/Documents/GitHub/go-example/image-600x600.png"
const imageTmpDir = ""              // OS default, e.g. /tmp
const imageTmpPattern = "crop-tmp-" // start of temp file
const x0 = 50                       // sample coordinates
const y0 = 50                       // sample coordinates
const x1 = 550                      // sample coordinates
const y1 = 550

// ReadFile is only used for testing.
// Not needed on final server, as server can do everything in RAM, without files.
// ReadFile: open and read an existing file
func ReadFile(filename string) (data []byte, err error) {
	// https://golang.org/pkg/io/ioutil/#ReadFile
	data, err = ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File reading error: ", err)
		// panic(err)
		return
	}
	fmt.Printf("Contents of file: %v ...\n", string(data)[0:127])
	return
}

// WriteFile is only used for testing.
// Not needed on final server, as server can do everything in RAM, without files.
// WriteFile write to file
func WriteFile(dir, pattern string, data []byte) (filename string, err error) {
	// https://golang.org/pkg/io/ioutil/#WriteFile
	tmpfile, err := ioutil.TempFile(dir, pattern)
	if err != nil {
		fmt.Println("File creation error: ", err)
		// panic(err)
		return
	}
	if _, err = tmpfile.Write(data); err != nil {
		fmt.Println("File write error: ", err)
		return
	}
	if err = tmpfile.Close(); err != nil {
		fmt.Println("File close error: ", err)
		return
	}
	fmt.Printf("Contents written to file: %v ...\n", string(data)[0:127])
	filename = tmpfile.Name()
	// Do NOT delete file! defer os.Remove(filename) // clean up
	return
}

func main() {
	// this file read is just for testing, the server does not need this
	data, err := ReadFile(imagePath)
	if err != nil {
		fmt.Printf("Crop failed. Could not read test image.\n")
	} else {
		// see https://golang.org/pkg/encoding/base64/
		var dataenc string
		var datadec []byte
		fmt.Printf("Contents of file (unencoded): %v ...\n", string(data)[0:127])
		// argument data is []byte, exactly what we need
		dataenc = base64.StdEncoding.EncodeToString(data) // []byte(arg)
		// dataenc is what the server receives in the REST API call
		// a base64 encoded image of any type (JPG, PNG) in a string
		fmt.Printf("Contents of file (encoded): %v ...\n", string(dataenc)[0:127])
		datadec, err = base64.StdEncoding.DecodeString(dataenc)
		if err != nil {
			fmt.Println("Decode failed with error: ", err)
			fmt.Println(err)
		} else {
			fmt.Printf("Contents of file (decoded): %v ...\n", string(datadec)[0:127])
			// create an io.Reader for []byte
			r := bytes.NewReader(datadec)
			// Calling the generic image.Decode() will convert the bytes into an image
			// and give us type of image as a string. E.g. "png"
			imageData, imageType, err := image.Decode(r)
			if err != nil {
				fmt.Println("Decoding image failed with error: ", err)
				fmt.Println(err)
			} else {
				// fmt.Println(imageData) // imageData (type image.Image)
				fmt.Printf("Image type is '%v'.\n", imageType)
				// use the coordinates received by server via REST API
				cropRect := image.Rect(x0, y0, x1, y1) // image.Rect(x0, y0, x1, y1)
				imageCropped := transform.Crop(imageData, cropRect)
				var b bytes.Buffer
				w := io.Writer(&b) // create a byte[] io.Writer using buffer
				// Encode takes a writer interface and an image interface
				// Since we want a PNG as output we convert to PNG
				png.Encode(w, imageCropped)
				imageCroppedString := b.String()
				imageCroppedBytes := b.Bytes()
				fmt.Printf("Contents of cropped image (unencoded): %v ...\n", imageCroppedString[0:127])
				// for testing purpose we write cropped image to file
				// The final server does not need to write to file.
				filenameCropTest, err := WriteFile(imageTmpDir, imageTmpPattern+"*.png", imageCroppedBytes)
				if err != nil {
					fmt.Println("Write cropped image to file for testing failed with error: ", err)
					fmt.Println(err)
				} else {
					fmt.Printf("Final result for testing is in file %v.\n", filenameCropTest)
					// argument imageCroppedBytes is []byte, exactly what we need
					datacropenc := base64.StdEncoding.EncodeToString(imageCroppedBytes) // []byte(arg)
					// datacropenc is what the server will store in MongoDB,
					// a cropped PNG image that is base64 encoded stored in a string.
					fmt.Printf("Contents of cropped PNG image (encoded): %v ...\n", string(datacropenc)[0:127])
					fmt.Printf("Crop completed successfully.\n")
				}
			}
		}
	}
}
