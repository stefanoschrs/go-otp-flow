package main

import (
	"bytes"
	"encoding/base64"
	"github.com/pquerna/otp"
	"html/template"
	"image/png"
	"io/ioutil"
	"strings"
)

func getBase64Image(key *otp.Key) (string, error) {
	img, err := key.Image(250, 250)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func loadTemplate() (*template.Template, error) {
	// TODO: Conditionally load from disk if templates.go not found
	// return template.ParseGlob("templates/*")

	t := template.New("")

	for name, file := range Assets.Files {
		if file.IsDir() || !strings.HasSuffix(name, ".tmpl") {
			continue
		}

		h, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		t, err = t.New(name).Parse(string(h))
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}
