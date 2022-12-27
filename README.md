# abyssa

A program for extracting text from screenshots.

## Requirements:
- `greenclip`: https://github.com/erebe/greenclip
    - Clipboard manager
- `Tesseract`: https://github.com/tesseract-ocr/tessdoc
    - OCR engine
    - Can be installed via most Linux distros' package manager
    - After installing: `wget https://github.com/tesseract-ocr/tessdata/raw/main/eng.traineddata` 
    - Move `eng.traineddata` to your `tessdata` directory. (`/usr/share/tessdata/` on my machine)
- `flameshot`: https://github.com/flameshot-org/flameshot
    - Screenshot software
    - Any other Linux screenshot software should work as long as it is compatible with `greenclip`
- `x-clip`
    - Used to insert text into system clipboard.
- `notify-send` 
    - Sends desktop notifications to the user via a notification daemon
    - Included in most Linux distros

## Example usecases:

### Copying code snippets from videos:

https://user-images.githubusercontent.com/48808721/209611091-a08a510d-9c9f-4a53-8c8e-1934208a301d.mp4

### Copying text from documents with buggy text selection:

https://user-images.githubusercontent.com/48808721/209611262-8ded81c8-93be-4c5b-bfc9-12b7882150eb.mp4


## Usage:
- Compile (`go build -o abyssa main.go`) and move `abyssa` binary to the desired `$PATH` directory.
- Start the abyssa daemon: `abyssa daemon`
- Toggle abyssa: `abyssa`
    - While abyssa is activated it will process *every* screenshot that is copied to the clipboard.
    - Binding a toggle key is recommended.
- Kill the abyssa daemon: `abyssa kill`

## Example i3 bindings:

### Starts abyssa daemon on startup:
`exec --no-startup-id abyssa daemon>/dev/null`

### Toggles abyssa:
`bindsym $mod+a	exec --no-startup-id abyssa`
