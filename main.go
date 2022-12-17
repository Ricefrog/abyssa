package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/otiai10/gosseract/v2"
	"github.com/radovskyb/watcher"
)

// TODO: figure out daemon functionality

const IMAGE_CACHE_DIR = "/tmp/greenclip/"

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

func main() {
	client := gosseract.NewClient()
	defer client.Close()

	dirWatcher := watcher.New()
	defer dirWatcher.Close()

	dirWatcher.FilterOps(watcher.Create)
	go func() {
		for {
			select {
			case event := <-dirWatcher.Event:

				client.SetImage(event.Path)

				text, err := client.Text()
				if err != nil {
					notification("abyssa: " + err.Error())
					log.Fatal(err)
				}

				text = strings.TrimSpace(text)

				if len(text) < 1 {
					// no text detected
					continue
				}

				err = sendToClipboard(text)
				if err != nil {
					notification("abyssa: " + err.Error())
					log.Fatal(err)
				}

				msg := fmt.Sprintf("Copied '%s' to clipboard.", text)
				go notification(msg)

			case err := <-dirWatcher.Error:
				notification("abyssa: " + err.Error())
				log.Fatal(err)
			case <-dirWatcher.Closed:
				return
			}
		}
	}()

	err := dirWatcher.Add(IMAGE_CACHE_DIR)
	if err != nil {
		notification("abyssa: " + err.Error())
		log.Fatal(err)
	}

	if err := dirWatcher.Start(time.Millisecond * 500); err != nil {
		notification("abyssa: " + err.Error())
		log.Fatal(err)
	}

}
