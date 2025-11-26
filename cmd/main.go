package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"slices"
	"strconv"

	rDraw "golang.org/x/image/draw"
)

var SIMPLE_CHARS = []string{"@", "#", "S", "%", "?", "*", "+", ";", ":", ",", "."}
var NORMAL_CHARS = []string{
	"$", "@", "B", "%", "8", "&", "W", "M", "#", "*", "o", "a", "h", "k", "b", "d", "p", "q", "w", "m",
	"Z", "O", "0", "Q", "L", "C", "J", "U", "Y", "X", "z", "c", "v", "u", "n", "x", "r", "j", "f", "t",
	"/", "\\", "|", "(", ")", "1", "{", "}", "[", "]", "?", "-", "_", "+", "~", "<", ">",
	"i", "!", "l", "I", ";", ":", ",", "\"", "^", "`", "'", ".", " ",
}

var CHAR_LIST = NORMAL_CHARS

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ascime")

		fmt.Println("usage: cgraph [image_path1, image_path2] [args...]")
		return
	}

	args := os.Args[1:]

	commands := []string{"-c", "-w", "-s"}
	width := 100
	skipCurrent := false
	var imagePaths []string
	for i, arg := range args {
		if !slices.Contains(commands, arg) {
			if skipCurrent {
				skipCurrent = false
				continue
			} else {
				imagePaths = append(imagePaths, arg)
			}
		} else {
			switch arg {
			case "-c":
				continue
			case "-w":
				if (i+1 < len(args)) {
					skipCurrent = true // skip next argument if not command
					intWidth, err := strconv.Atoi(args[i+1])
					if err != nil {
						fmt.Printf("expected number after -w, got: %q", args[i+1])
						return
					}
					width = intWidth
				} else {
					continue
				}

			case "-s":
				CHAR_LIST = SIMPLE_CHARS
			}
		}
	}

	for _, path := range imagePaths {
		asciiStr := imageToAscii(path, width)
		fmt.Println(asciiStr)
	}
}

func imageToAscii(path string, newWidth int) string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	img = resizeImage(img, newWidth)
	img = toGray(img)
	asciiStr := pixelsToAscii(img)

	return asciiStr
}

func resizeImage(src image.Image, newWidth int) image.Image {
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	charRatio := 1.65
	newHeight := int(float64(height) * float64(newWidth) / float64(width) / charRatio)

	if newHeight <= 0 {
		newHeight = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	rDraw.CatmullRom.Scale(
		dst,
		dst.Bounds(),
		src,
		bounds,
		draw.Over,
		nil,
	)

	return dst
}

func toGray(src image.Image) *image.Gray {
	bounds := src.Bounds()

	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := src.At(x, y).RGBA()

			yValue := uint8((299*r + 587*g + 114*b) / 1000 >> 8)

			gray.Set(x, y, color.Gray{Y: yValue})
		}
	}

	return gray
}

func pixelsToAscii(src image.Image) string {
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	asciiStr := ""

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, _, _, _ := src.At(x, y).RGBA()
			gray := uint8(r >> 8)
			index := int(gray) * (len(CHAR_LIST) - 1) / 255
			asciiStr += CHAR_LIST[index]
		}

		asciiStr += "\n"
	}

	return asciiStr
}
