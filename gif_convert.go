package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"log"
	"os"
)

// Decode reads and analyzes the given reader as a GIF image
func SplitAnimatedGIF(path string) (frames []string, err error) {
    reader, err := os.Open(path)
    if err != nil {
        log.Fatal(err)
    }

    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("error while decoding: %s", r)
        }
    }()

    gif, err := gif.DecodeAll(reader)

    if err != nil {
        return []string{}, err
    }

    imgWidth, imgHeight := getGifDimensions(gif)

    overpaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
    draw.Draw(overpaintImage, overpaintImage.Bounds(), gif.Image[0], image.ZP, draw.Src)

    for i, srcImg := range gif.Image {
        draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Over)

        // save current frame "stack". This will overwrite an existing file with that name
        os.MkdirAll("tmp", os.ModePerm)
        path := fmt.Sprintf("%s%d%s", "tmp/", i, ".png")
        frames = append(frames, path)
        file, err := os.Create(path)
        if err != nil {
            return []string{}, err
        }

        err = png.Encode(file, overpaintImage)
        if err != nil {
            return []string{}, err
        }

        file.Close()
    }

    return frames, nil
}

func getGifDimensions(gif *gif.GIF) (x, y int) {
    var lowestX int
    var lowestY int
    var highestX int
    var highestY int

    for _, img := range gif.Image {
        if img.Rect.Min.X < lowestX {
            lowestX = img.Rect.Min.X
        }
        if img.Rect.Min.Y < lowestY {
            lowestY = img.Rect.Min.Y
        }
        if img.Rect.Max.X > highestX {
            highestX = img.Rect.Max.X
        }
        if img.Rect.Max.Y > highestY {
            highestY = img.Rect.Max.Y
        }
    }

    return highestX - lowestX, highestY - lowestY
}