# abyssa

A program for extracting text from screenshots.

Install greenclip.

Install Tesseract: https://github.com/tesseract-ocr/tessdoc

`wget https://github.com/tesseract-ocr/tessdata/raw/main/eng.traineddata` and move `eng.traineddata` to your `tessdata` directory. `/usr/share/tessdata/` on my machine.

## Usage

Start the abyssa daemon:
`abyssa daemon`

Toggle the abyssa daemon:
`abyssa`

Kill the abyssa daemon:
`abyssa kill`
