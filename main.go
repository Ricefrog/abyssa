package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/otiai10/gosseract/v2"
	"github.com/radovskyb/watcher"
)

// TODO: add dunst notification

const IMAGE_CACHE_DIR = "/tmp/greenclip/"

func getMostNewestImagePath() string {
	return "42"
}

func main() {
	client := gosseract.NewClient()
	defer client.Close()

	/*
		err := client.SetImage("./example/two_words.png")
		if err != nil {
			log.Fatal(err)
		}

		text, err := client.Text()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("sanity check: %s\n", text)
	*/

	dirWatcher := watcher.New()
	defer dirWatcher.Close()

	dirWatcher.FilterOps(watcher.Create)
	go func() {
		for {
			select {
			case event := <-dirWatcher.Event:
				//fmt.Printf("%s was created.\n", event.Path)

				client.SetImage(event.Path)

				text, err := client.Text()
				if err != nil {
					log.Fatal(err)
				}

				text = strings.TrimSpace(text)
				if len(text) > 0 {
					fmt.Println(text)
				}

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
