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

// TODO: add dunst notification

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

func main() {

	//replacer := strings.NewReplacer("\"", "\"")

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
					log.Fatal(err)
				}

				//fmt.Printf("text before: %s\n", text)
				text = strings.TrimSpace(text)
				//fmt.Printf("text after: %s\n", text)

				if len(text) < 1 {
					// no text detected
					//fmt.Println("no text detected")
					continue
				}

				err = sendToClipboard(text)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("copied '%s' to clipboard\n", text)

			case err := <-dirWatcher.Error:
				log.Fatal(err)
			case <-dirWatcher.Closed:
				return
			}
		}
	}()

	err := dirWatcher.Add(IMAGE_CACHE_DIR)
	if err != nil {
		log.Fatal(err)
	}

	if err := dirWatcher.Start(time.Millisecond * 500); err != nil {
		log.Fatal(err)
	}

}
