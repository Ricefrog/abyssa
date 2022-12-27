package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/otiai10/gosseract/v2"
	"github.com/radovskyb/watcher"
	"golang.org/x/image/draw"
)

const IMAGE_CACHE_DIR = "/tmp/greenclip/"
const DPID_FILE = "/tmp/abyssa"
const AREA_LOWER_BOUND = 550 * 550

var toggle = true

func sendToClipboard(text string) error {
	cmd := exec.Command("xclip")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, text)
	}()

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func notification(msg string) {
	cmd := exec.Command("notify-send", "--icon=none", msg)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func handleError(err error) {
	notification("abyssa: " + err.Error())
	log.Fatal(err)
}

func handleErrorWithMsg(err error, msg string) {
	notification("abyssa: " + err.Error() + " " + msg)
	log.Fatal(err)
}

func getDaemonPID() string {
	cmd := exec.Command("cat", DPID_FILE)
	b, err := cmd.Output()
	if err != nil {
		handleErrorWithMsg(err, "while trying to retrieve daemon PID.")
	}
	return string(b)
}

func resize(scale float32, src image.Image) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, int(float32(src.Bounds().Max.X)*scale), int(float32(src.Bounds().Max.Y)*scale)))
	draw.ApproxBiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func getText(client *gosseract.Client, path string) (string, error) {
	var img *image.Image
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	unaltered, err := png.Decode(f)
	if err != nil {
		return "", err
	}

	area := unaltered.Bounds().Max.X * unaltered.Bounds().Max.Y
	if area < AREA_LOWER_BOUND {
		scale := (float32(AREA_LOWER_BOUND) / float32(area)) / 2
		temp := resize(scale, unaltered)
		img = &temp
	} else {
		img = &unaltered
	}

	buf := bytes.NewBuffer([]byte{})

	err = png.Encode(buf, *img)
	if err != nil {
		return "", err
	}

	err = client.SetImageFromBytes(buf.Bytes())
	if err != nil {
		return "", err
	}

	text, err := client.Text()
	return text, err
}

func startDaemon() *watcher.Watcher {
	dirWatcher := watcher.New()
	go func() {
		client := gosseract.NewClient()
		defer client.Close()

		dirWatcher.FilterOps(watcher.Create)
		go func() {
			for {
				select {
				case event := <-dirWatcher.Event:
					text, err := getText(client, event.Path)
					if err != nil {
						handleErrorWithMsg(err, "while extracting text from image.")
					}

					text = strings.TrimSpace(text)

					if len(text) < 1 {
						// no text detected
						continue
					}

					err = sendToClipboard(text)
					if err != nil {
						handleErrorWithMsg(err, "while sending text to clipboard.")
					}

					msg := fmt.Sprintf("Copied '%s' to clipboard.", text)
					go notification(msg)
				case err := <-dirWatcher.Error:
					handleError(err)
				case <-dirWatcher.Closed:
					return
				}
			}
		}()

		err := dirWatcher.Add(IMAGE_CACHE_DIR)
		if err != nil {
			handleError(err)
		}

		if err := dirWatcher.Start(time.Millisecond * 500); err != nil {
			handleError(err)
		}
	}()
	return dirWatcher
}

func stopDaemon(dirWatcher *watcher.Watcher) {
	dirWatcher.Close()
}

func main() {
	flag := "none"
	if len(os.Args) > 1 {
		flag = os.Args[1]
	}

	if flag == "daemon" {
		// write pid of daemon to /tmp/abyssa
		file, err := os.Create(DPID_FILE)
		if err != nil {
			handleError(err)
		}

		_, err = file.Write([]byte(strconv.Itoa(os.Getpid())))
		if err != nil {
			handleError(err)
		}
		file.Close()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGUSR1)
		blocker := make(chan struct{}, 1)

		var dirWatcher *watcher.Watcher

		for {
			if toggle {
				go notification("abyssa: activated")
				dirWatcher = startDaemon()
			} else {
				go notification("abyssa: deactivated")
				go stopDaemon(dirWatcher)
			}
			go func() {
				<-sig
				blocker <- struct{}{}
			}()

			fmt.Println("awaiting signal")
			<-blocker
			toggle = !toggle
		}
	} else if flag == "kill" {
		// send killsignal to daemon
		daemonPID := getDaemonPID()
		cmd := exec.Command("kill", "-9", daemonPID)
		if err := cmd.Run(); err != nil {
			handleErrorWithMsg(err, "while trying to send kill signal to daemon.")
		}
	} else {
		// send toggle signal to daemon
		daemonPID := getDaemonPID()
		cmd := exec.Command("kill", "-SIGUSR1", daemonPID)
		if err := cmd.Run(); err != nil {
			handleErrorWithMsg(err, "while trying to send toggle signal to daemon.")
		}
	}

	return
}
