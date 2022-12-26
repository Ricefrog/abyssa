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

func writeImage(img image.Image, filename string) {
	out, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = png.Encode(out, img)
	if err != nil {
		log.Fatal(err)
	}
}

func printDiffs(expected, reference, label string) {
	fmt.Printf("\n%s:\n", label)
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

func printScaleDiffs(scale float32, client *gosseract.Client, img image.Image) {
	label_nn := fmt.Sprintf("upsized %fx nn", scale)
	label_abl := fmt.Sprintf("upsized %fx abl", scale)
	label_bl := fmt.Sprintf("upsized %fx bl", scale)

	upsized_nn := resizeNN(scale, img)
	printSize(upsized_nn, label_nn)
	printDiffs(expected, getText(client, upsized_nn), label_nn)

	upsized_abl := resizeABL(scale, img)
	printSize(upsized_abl, label_abl)
	printDiffs(expected, getText(client, upsized_abl), label_abl)

	upsized_bl := resizeBL(scale, img)
	printSize(upsized_bl, label_bl)
	printDiffs(expected, getText(client, upsized_bl), label_bl)
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

func resizeNN(scale float32, src image.Image) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, int(float32(src.Bounds().Max.X)*scale), int(float32(src.Bounds().Max.Y)*scale)))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func resizeABL(scale float32, src image.Image) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, int(float32(src.Bounds().Max.X)*scale), int(float32(src.Bounds().Max.Y)*scale)))
	draw.ApproxBiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func resizeBL(scale float32, src image.Image) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, int(float32(src.Bounds().Max.X)*scale), int(float32(src.Bounds().Max.Y)*scale)))
	draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func interpolations() {
	f, err := os.Open("./example/code_from_video.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	printSize(img, "original size")

	writeImage(resizeNN(2, img), "./nearest_neighbor.png")
	writeImage(resizeNN(2, img), "./approx_bilinear.png")
	writeImage(resizeNN(2, img), "./bilinear.png")
}

func compareBroad() {
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
	printSize(img, "original size")

	unaltered = getText(client, img)

	grayedRes := image.NewGray(img.Bounds())
	draw.Draw(grayedRes, grayedRes.Bounds(), img, img.Bounds().Min, draw.Src)

	grayed = getText(client, grayedRes)

	fmt.Printf("\nexpected: \n%s\n", expected)
	printDiffs(expected, unaltered, "unaltered")
	printDiffs(expected, grayed, "grayed")

	printScaleDiffs(2, client, img)
	printScaleDiffs(2, client, grayedRes)

	printScaleDiffs(3, client, img)
	printScaleDiffs(3, client, grayedRes)

	// abl 2x looks like the best option
}

// trying to determine how scale should be calculated
func compareScale() {
	var unaltered string
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
	printSize(img, "original size")

	area := img.Bounds().Max.X * img.Bounds().Max.Y
	fmt.Println("area of original image: %d\n", area)

	scale_400 := (float32(400*400) / float32(area)) / 2
	fmt.Printf("scale for 400x400 lower bound: %f\n", scale_400)
	scale_450 := (float32(450*450) / float32(area)) / 2
	fmt.Printf("scale for 450x450 lower bound: %f\n", scale_450)
	scale_500 := (float32(500*500) / float32(area)) / 2
	fmt.Printf("scale for 500x500 lower bound: %f\n", scale_500)
	scale_550 := (float32(550*550) / float32(area)) / 2
	fmt.Printf("scale for 550x550 lower bound: %f\n", scale_550)
	scale_600 := (float32(600*600) / float32(area)) / 2
	fmt.Printf("scale for 600x600 lower bound: %f\n", scale_600)

	unaltered = getText(client, img)
	fmt.Printf("\nexpected: \n%s\n", expected)
	printDiffs(expected, unaltered, "unaltered")
	printDiffs(expected, getText(client, resizeABL(scale_400, img)), "400")
	printDiffs(expected, getText(client, resizeABL(scale_450, img)), "450")
	printDiffs(expected, getText(client, resizeABL(scale_500, img)), "500")
	printDiffs(expected, getText(client, resizeABL(scale_550, img)), "550")
	printDiffs(expected, getText(client, resizeABL(scale_600, img)), "600")
	// aiming for 550 looks good
}

func main() {
	compareScale()
	return
}
