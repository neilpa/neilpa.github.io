package main

import (
	//"html/template"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func handleIndex(r *http.Request) (string, int) {
	buf := bytes.NewBuffer(nil) // todo: better buffer size

	index, err := os.Open("index.md")
	if err != nil {
		// todo: log
		return "unexpected error", 500
	}

	list, err := ioutil.ReadDir("./static")
	if err != nil {
		// todo: log
		return "unexpected error", 500
	}

	_, err = io.Copy(buf, index)
	if err != nil {
		// todo: log
		return "unexpected error", 500
	}

	for _, info := range list {
		// todo: better parsing of "dated posts"
		if !strings.HasPrefix(info.Name(), "2") {
			continue
		}
		buf.WriteString("* ")
		buf.WriteString(info.Name())
		buf.WriteString("\n")
	}

	return buf.String(), 200
}

const index = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>neilpa.me</title>
  <!-- <link rel="stylesheet" href="style.css"> -->
  <!-- <script src="script.js"></script> -->
</head>
<body>
  Hello World!
</body>
</html>`
