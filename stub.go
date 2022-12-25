package main

// Using this file to test methods for improving OCR accuracy

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strings"

	"golang.org/x/image/draw"

	"github.com/otiai10/gosseract/v2"
)

const expected = `func main() {
start := time.Now()
userName := fetchUser() // 100ms
respch := make(chan any, 2)
wg := &sync.WaitGroup{}
wg.Add(2)
go fetchUserLikes(userName, respch)
go fetchUserMatch(userName, respch)
wg.Wait() // block until 2
close(respch)
for resp := range respch {
fmt.Println("resp: ", resp)
}
fmt.Println("took: ", time.Since(start))
}`

func printDiffs(expected, reference, label string) {
	fmt.Println("\n%s:", label)
	expectedLines := strings.Split(expected, "\n")
	referenceLines := strings.Split(reference, "\n")
	total := 0

	length := len(expectedLines)

	if len(expectedLines) != len(referenceLines) {
		total++
		fmt.Printf("Number of lines differ -> expected: %d, reference: %d\n", len(expectedLines), len(referenceLines))
		if len(expectedLines) > len(referenceLines) {
			length = len(referenceLines)
		}
	}

	for i := 0; i < length; i++ {
		eLine := expectedLines[i]
		rLine := referenceLines[i]
		count := 0

		lineLength := len(eLine)
		if len(eLine) != len(rLine) {
			count++
			if len(eLine) > len(rLine) {
				lineLength = len(rLine)
			}
		}

		for j := 0; j < lineLength; j++ {
			if eLine[j] != rLine[j] {
				count++
			}
		}
		fmt.Print(strings.TrimSpace(rLine))
		if count > 0 {
			fmt.Printf(" -> %d diffs \n", count)
		} else {
			fmt.Println()
		}
		total += count
	}
	fmt.Printf("total diffs: %d\n", total)
}

func getText(client *gosseract.Client, img image.Image) string {
	buf := bytes.NewBuffer([]byte{})

	err := png.Encode(buf, img)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SetImageFromBytes(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	text, err := client.Text()
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(text)
}

func printSize(img image.Image, label string) {
	r := img.Bounds()
	fmt.Printf("%s -> w: %d, h: %d\n", label, r.Max.X, r.Max.Y)
}

func resize(scale float32, src image.Image) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, int(float32(src.Bounds().Max.X)*scale), int(float32(src.Bounds().Max.Y)*scale)))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func stub() {
	var (
		unaltered string
		grayed    string
	)
	client := gosseract.NewClient()
	defer client.Close()

	f, err := os.Open("./example/code_from_video.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	printSize(img, "src")

	upsized := resize(3, img)
	printSize(upsized, "upsized")
	return

	unaltered = getText(client, img)

	grayedRes := image.NewGray(img.Bounds())
	draw.Draw(grayedRes, grayedRes.Bounds(), img, img.Bounds().Min, draw.Src)

	grayed = getText(client, img)

	/*
		out, err := os.Create("./out.png")
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		err = png.Encode(out, result)
		if err != nil {
			log.Fatal(err)
		}
	*/

	fmt.Printf("\nexpected: \n%s\n", expected)
	printDiffs(expected, unaltered, "unaltered")
	printDiffs(expected, grayed, "grayed")
}

func main() {
	stub()
}
