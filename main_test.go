/*
Copyright 2026 Veriphor LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/scripts",
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"img": cmdImg,
		},
	})
}

// cmdImg implements the "img" testscript command, which decodes an image and
// verifies its format, dimensions, and pixel colors. Comparing decoded pixels
// instead of encoded bytes keeps the tests stable when an encoder's output
// changes (for example, when a Go release changes zlib compression).
//
// Usage:
//
//	img [-b] [-tol n] <file> <format> <width> <height> <rrggbbaa>
//
// The file may be "stdout", "stderr", or a path relative to the work
// directory. Use -b if the file contains base64-encoded data. Every pixel
// must match the expected non-premultiplied color within -tol per channel.
func cmdImg(ts *testscript.TestScript, neg bool, args []string) {
	fs := flag.NewFlagSet("img", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	isBase64 := fs.Bool("b", false, "file contains base64-encoded data")
	tol := fs.Int("tol", 0, "maximum per-channel difference")
	if err := fs.Parse(args); err != nil || fs.NArg() != 5 {
		ts.Fatalf("usage: img [-b] [-tol n] <file> <format> <width> <height> <rrggbbaa>")
	}

	err := checkImg(ts.ReadFile(fs.Arg(0)), *isBase64, *tol, fs.Args()[1:])
	switch {
	case neg && err == nil:
		ts.Fatalf("image unexpectedly matched")
	case !neg && err != nil:
		ts.Fatalf("%v", err)
	}
}

// checkImg decodes data and compares it to want, which holds the expected
// format, width, height, and color, in that order.
func checkImg(data string, isBase64 bool, tol int, want []string) error {
	wantFormat := want[0]
	wantW, err := strconv.Atoi(want[1])
	if err != nil {
		return fmt.Errorf("invalid width %q", want[1])
	}
	wantH, err := strconv.Atoi(want[2])
	if err != nil {
		return fmt.Errorf("invalid height %q", want[2])
	}
	// Parse the expected color independently of parseColor, the function
	// under test.
	v, err := strconv.ParseUint(want[3], 16, 32)
	if err != nil || len(want[3]) != 8 {
		return fmt.Errorf("invalid color %q: expected 8 hex digits (rrggbbaa)", want[3])
	}
	wantC := color.NRGBA{R: uint8(v >> 24), G: uint8(v >> 16), B: uint8(v >> 8), A: uint8(v)}

	raw := []byte(data)
	if isBase64 {
		raw, err = base64.StdEncoding.DecodeString(string(bytes.TrimSpace(raw)))
		if err != nil {
			return fmt.Errorf("decoding base64: %v", err)
		}
	}

	img, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("decoding image: %v", err)
	}
	if format != wantFormat {
		return fmt.Errorf("format: got %s, want %s", format, wantFormat)
	}
	b := img.Bounds()
	if b.Dx() != wantW || b.Dy() != wantH {
		return fmt.Errorf("dimensions: got %dx%d, want %dx%d", b.Dx(), b.Dy(), wantW, wantH)
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			got := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if !colorWithin(got, wantC, tol) {
				return fmt.Errorf("pixel (%d,%d): got %02x%02x%02x%02x, want %02x%02x%02x%02x (tolerance %d)",
					x, y, got.R, got.G, got.B, got.A, wantC.R, wantC.G, wantC.B, wantC.A, tol)
			}
		}
	}
	return nil
}

// colorWithin reports whether each channel of a and b differs by at most tol.
func colorWithin(a, b color.NRGBA, tol int) bool {
	within := func(x, y uint8) bool {
		d := int(x) - int(y)
		return d >= -tol && d <= tol
	}
	return within(a.R, b.R) && within(a.G, b.G) && within(a.B, b.B) && within(a.A, b.A)
}

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"mkimg": main,
	})
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input   string
		want    color.RGBA
		wantErr bool
	}{
		// 6-digit hex, no prefix
		{"ff0000", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"00ff00", color.RGBA{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF}, false},
		{"0000ff", color.RGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xFF}, false},
		{"ffffff", color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, false},
		{"000000", color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"3b82f6", color.RGBA{R: 0x3B, G: 0x82, B: 0xF6, A: 0xFF}, false},

		// 6-digit hex, with # prefix
		{"#ff0000", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"#3b82f6", color.RGBA{R: 0x3B, G: 0x82, B: 0xF6, A: 0xFF}, false},

		// 8-digit hex (with alpha)
		{"ff000080", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0x80}, false},
		{"#ff000080", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0x80}, false},
		{"00000000", color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}, false},

		// 3-digit shorthand
		{"f00", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"#f00", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"fff", color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, false},
		{"abc", color.RGBA{R: 0xAA, G: 0xBB, B: 0xCC, A: 0xFF}, false},

		// 4-digit shorthand (with alpha)
		{"f00f", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"#f008", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0x88}, false},
		{"0000", color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}, false},

		// Named colors (lowercase)
		{"red", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"green", color.RGBA{R: 0x00, G: 0x80, B: 0x00, A: 0xFF}, false},
		{"blue", color.RGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xFF}, false},
		{"white", color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, false},
		{"black", color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"rebeccapurple", color.RGBA{R: 0x66, G: 0x33, B: 0x99, A: 0xFF}, false},
		{"transparent", color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}, false},

		// Named colors (uppercase / mixed case)
		{"Red", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"RED", color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}, false},
		{"Tomato", color.RGBA{R: 0xFF, G: 0x63, B: 0x47, A: 0xFF}, false},
		{"CORNFLOWERBLUE", color.RGBA{R: 0x64, G: 0x95, B: 0xED, A: 0xFF}, false},

		// Errors
		{"", color.RGBA{}, true},
		{"gg0000", color.RGBA{}, true},
		{"12345", color.RGBA{}, true},
		{"1234567", color.RGBA{}, true},
		{"notacolor", color.RGBA{}, true},
		{"#gg0000", color.RGBA{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseColor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseColor(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseColor(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
