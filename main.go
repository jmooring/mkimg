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
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "(devel)"
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var (
		showHelp    bool
		showVersion bool
		format      string
		outputFile  string
		asBase64    bool
	)

	fs.BoolVar(&showHelp, "h", false, "show this help message")
	fs.BoolVar(&showHelp, "help", false, "show this help message")
	fs.BoolVar(&showVersion, "v", false, "print version and exit")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")
	fs.StringVar(&format, "f", "png", "output format (png, jpeg/jpg, gif)")
	fs.StringVar(&format, "format", "png", "output format (png, jpeg/jpg, gif)")
	fs.StringVar(&outputFile, "o", "", "write output to file")
	fs.StringVar(&outputFile, "output", "", "write output to file")
	fs.BoolVar(&asBase64, "b", false, "encode output as base64")
	fs.BoolVar(&asBase64, "base64", false, "encode output as base64")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run '%s -h' for help.\n", os.Args[0])
		os.Exit(1)
	}

	if showHelp {
		fmt.Printf("Usage: %s [flags] <width> <height> <color>\n", os.Args[0])
		fmt.Println()
		fmt.Println("Generates a uniform-color image and outputs it as raw bytes")
		fmt.Println("or a base64 string to either stdout or a file.")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  width   image width in pixels (1–10000)")
		fmt.Println("  height  image height in pixels (1–10000)")
		fmt.Println("  color   hex value or CSS named color (see examples below)")
		fmt.Println()
		fmt.Println("Color formats:")
		fmt.Println("  ff0000        6-digit hex (RGB)")
		fmt.Println("  ff000080      8-digit hex (RGBA)")
		fmt.Println("  f00           3-digit hex (RGB), expanded to ff0000")
		fmt.Println("  f008          4-digit hex (RGBA), expanded to ff000088")
		fmt.Println("  '#ff0000'     hex with # prefix (must be quoted)")
		fmt.Println("  red           CSS named color")
		fmt.Println("  transparent   fully transparent black")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Create a 6x7 semi-transparent red PNG")
		fmt.Printf("  %s -o semi-transparent-red.png 6 7 ff000088\n\n", os.Args[0])
		fmt.Println("  # Generate a 42x42 blue JPEG using shorthand hex")
		fmt.Printf("  %s -f jpeg -o blue.jpg 42 42 00f\n\n", os.Args[0])
		fmt.Println("  # Output a 1x1 transparent PNG as base64 to stdout")
		fmt.Printf("  %s -b 1 1 transparent\n\n", os.Args[0])
		fmt.Println("  # Redirect raw bytes to a file or pipe to another tool")
		fmt.Printf("  %s 100 100 red > red.png\n", os.Args[0])
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -h, --help             show this help message")
		fmt.Println("  -v, --version          print version and exit")
		fmt.Println("  -f, --format string    output format: png, jpeg (or jpg), gif (default \"png\")")
		fmt.Println("  -o, --output string    write output to file (default: stdout)")
		fmt.Println("  -b, --base64           encode output as base64")
		fmt.Println()
		fmt.Println("Notes:")
		fmt.Println("  - JPEG and GIF do not support an alpha channel; the alpha")
		fmt.Println("    component of the specified color is ignored and all pixels")
		fmt.Println("    are fully opaque.")
		fmt.Println("  - Width and height are capped at 10000 pixels per side to")
		fmt.Println("    prevent excessive memory use.")
		fmt.Println("  - Hex digits are case-insensitive; FF0000 and ff0000 are equivalent.")
		os.Exit(0)
	}

	if showVersion {
		fmt.Println(version())
		os.Exit(0)
	}

	args := fs.Args()
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <width> <height> <color>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Run '%s -h' for help.\n", os.Args[0])
		os.Exit(1)
	}

	width, err := strconv.Atoi(args[0])
	if err != nil || width <= 0 || width > 10000 {
		fmt.Fprintf(os.Stderr, "Error: width must be an integer between 1 and 10000\n")
		fmt.Fprintf(os.Stderr, "Run '%s -h' for help.\n", os.Args[0])
		os.Exit(1)
	}

	height, err := strconv.Atoi(args[1])
	if err != nil || height <= 0 || height > 10000 {
		fmt.Fprintf(os.Stderr, "Error: height must be an integer between 1 and 10000\n")
		fmt.Fprintf(os.Stderr, "Run '%s -h' for help.\n", os.Args[0])
		os.Exit(1)
	}

	c, err := parseColor(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run '%s -h' for help.\n", os.Args[0])
		os.Exit(1)
	}

	format = strings.ToLower(format)
	switch format {
	case "png", "jpeg", "jpg", "gif":
		// valid
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown format %q: must be one of png, jpeg, gif\n", format)
		fmt.Fprintf(os.Stderr, "Run '%s -h' for help.\n", os.Args[0])
		os.Exit(1)
	}

	bounds := image.Rect(0, 0, width, height)
	var buf bytes.Buffer
	switch format {
	case "png", "jpeg", "jpg":
		img := image.NewRGBA(bounds)
		pix := img.Pix
		for i := 0; i < len(pix); i += 4 {
			pix[i], pix[i+1], pix[i+2], pix[i+3] = c.R, c.G, c.B, c.A
		}
		if format == "png" {
			err = png.Encode(&buf, img)
		} else {
			err = jpeg.Encode(&buf, img, nil)
		}
	case "gif":
		palette := color.Palette{color.RGBA{R: c.R, G: c.G, B: c.B, A: 0xFF}}
		paletted := image.NewPaletted(bounds, palette)
		err = gif.EncodeAll(&buf, &gif.GIF{
			Image: []*image.Paletted{paletted},
			Delay: []int{0},
		})
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding image: %v\n", err)
		os.Exit(1)
	}

	var out []byte
	if asBase64 {
		out = []byte(base64.StdEncoding.EncodeToString(buf.Bytes()) + "\n")
	} else {
		out = buf.Bytes()
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, out, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
	} else {
		if _, err := os.Stdout.Write(out); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
	}
}

func parseColor(s string) (color.RGBA, error) {
	// Try named color first (case-insensitive, no leading #).
	if c, ok := namedColors[strings.ToLower(s)]; ok {
		return c, nil
	}
	orig := s
	s = strings.TrimPrefix(s, "#")
	// Expand shorthand: 3 → 6, 4 → 8
	if len(s) == 3 || len(s) == 4 {
		b := []byte(s)
		exp := make([]byte, len(b)*2)
		for i, ch := range b {
			exp[i*2] = ch
			exp[i*2+1] = ch
		}
		s = string(exp)
	}
	switch len(s) {
	case 6:
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("invalid color %q", orig)
		}
		return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xff}, nil
	case 8:
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("invalid color %q", orig)
		}
		return color.RGBA{R: uint8(v >> 24), G: uint8(v >> 16), B: uint8(v >> 8), A: uint8(v)}, nil
	default:
		return color.RGBA{}, fmt.Errorf("invalid color %q: expected 6 or 8 hex digits", orig)
	}
}
