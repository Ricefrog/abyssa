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

## Usage:
- Start the abyssa daemon: `abyssa daemon`
- Toggle abyssa: `abyssa`
    - While abyssa is activated it will process *every* screenshot that is copied to the clipboard.
    - Binding a toggle key is reccomended.
- Kill the abyssa daemon: `abyssa kill`

## Example i3 bindings:

### Starts abyssa daemon on startup:
`exec --no-startup-id abyssa daemon>/dev/null`

### Toggles abyssa:
`bindsym $mod+a	exec --no-startup-id abyssa`

## Example usecases:

### Copying code snippets from videos:

### Copying text from documents with buggy text selection: