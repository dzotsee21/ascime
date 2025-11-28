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

// default params
var CHAR_LIST = DEFAULT_CHARS
var colored = false
var grayColored = false
var onlyColored = false
var width = 100

func main() {
	if len(os.Args) < 2 {
		rgb := "\x1b[38;2;0;200;255m"
		reset := "\x1b[0m"

		fmt.Println(rgb + `
		░█▀▀▄░█▀▀░█▀▄░░▀░░█▀▄▀█░█▀▀
		▒█▄▄█░▀▀▄░█░░░░█▀░█░▀░█░█▀▀
		▒█░▒█░▀▀▀░▀▀▀░▀▀▀░▀░░▒▀░▀▀▀
		` + reset)

		fmt.Println("usage: ascime [image_path1, image_path2] [args...]")
		return
	}

	args := os.Args[1:]
	imagePaths, err := setParams(args)
	if err != nil {
		fmt.Println(err)
	}

	run(imagePaths)
}

func setParams(args []string) ([]string, error) {
	var imagePaths []string
	commands := []string{"-c", "-w", "-a"}
	skipCurrent := false
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
				var chosenOption string
				if i+1 < len(args) {
					chosenOption = args[i+1]
				}

				options := []string{"gray", "only", ""}
				if !slices.Contains(options, chosenOption) && !slices.Contains(commands, chosenOption) {
					return []string{}, fmt.Errorf("expected either [gray, only or ''] after -c, got: %q", chosenOption)
				}

				skipCurrent = true // skip next argument if not command
				switch chosenOption {
				case "gray":
					grayColored = true
				case "only":
					colored = true
					onlyColored = true
				default:
					colored = true
				}

			case "-w":
				if i+1 < len(args) {
					skipCurrent = true // skip next argument if not command
					intWidth, err := strconv.Atoi(args[i+1])
					if err != nil {
						return []string{}, fmt.Errorf("expected number after -w, got: %q", args[i+1])
					}
					width = intWidth
				}
			case "-a":
				if i+1 < len(args) {
					option := args[i+1]
					skipCurrent = true
					switch option {
					case "s", "simple":
						CHAR_LIST = SIMPLE_CHARS
					case "d", "default":
						CHAR_LIST = DEFAULT_CHARS
					case "e", "extended":
						CHAR_LIST = EXTENDED_CHARS
					case "u", "unicode":
						CHAR_LIST = UNICODE_CHARS
					default:
						CHAR_LIST = SIMPLE_CHARS
					}
				}
			}
		}
	}

	return imagePaths, nil
}

func run(imagePaths []string) {
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

	var asciiStr string
	if !colored {
		grayImg := toGray(img)
		asciiStr = pixelsToAscii(grayImg)
	} else {
		asciiStr = pixelsToAscii(img)
	}
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
			c := color.RGBAModel.Convert(src.At(x, y)).(color.RGBA)
			gray := c.R

			var index int
			if !onlyColored {
				index = int(gray) * (len(CHAR_LIST) - 1) / 255
			} else {
				index = 0
			}

			r, g, b := c.R, c.G, c.B
			char := CHAR_LIST[index]
			if colored || grayColored {
				asciiStr += fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, char)
			} else {
				asciiStr += char
			}
		}

		asciiStr += "\n"
	}

	return asciiStr
}
