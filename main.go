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
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"

	"github.com/HugoSmits86/nativewebp"
)

// maxDim is the maximum image width and height in pixels.
const maxDim = 10000

// helpText is the help message. %[1]s is replaced with the program name.
const helpText = `Usage: %[1]s [flags] <width> <height> <color>

Generates a uniform-color image and outputs it as raw bytes
or a base64 string to either stdout or a file.

Arguments:
  width   image width in pixels [1, 10000]
  height  image height in pixels [1, 10000]
  color   hex value or CSS named color (see examples below)

Color formats:
  ff0000        6-digit hex (RGB)
  ff000080      8-digit hex (RGBA)
  f00           3-digit hex (RGB), expanded to ff0000
  f008          4-digit hex (RGBA), expanded to ff000088
  '#ff0000'     hex with # prefix (must be quoted)
  red           CSS named color
  transparent   fully transparent black

Examples:
  # Create a 6x7 semi-transparent red PNG
  %[1]s -o semi-transparent-red.png 6 7 ff000088

  # Generate a 42x42 blue JPEG using shorthand hex
  %[1]s -o blue.jpg 42 42 00f

  # Output a 1x1 transparent WebP as base64 to stdout
  %[1]s -f webp -b 1 1 transparent

  # Output a 1x1 transparent PNG as base64 to stdout (default format)
  %[1]s -b 1 1 transparent

  # Redirect raw bytes to a file or pipe to another tool
  %[1]s 100 100 red > red.png

Flags:
  -h, --help             show this help message
  -v, --version          print version and exit
  -f, --format string    gif, jpeg, png (default), webp
  -o, --output string    write output to file
  -b, --base64           encode output as base64

Notes:
  - If -f is omitted, the format is inferred from the -o file extension
    (.gif, .jpeg, .jpg, .png, .webp); defaults to PNG when writing to
    stdout; an unrecognized extension is an error.
  - JPEG and GIF do not support an alpha channel; the alpha
    component of the specified color is ignored and all pixels
    are fully opaque.
  - WebP output is lossless and supports an alpha channel.
  - Format names are case-insensitive; PNG and png are equivalent.
  - Hex digits are case-insensitive; FF0000 and ff0000 are equivalent.
`

// encoder writes a width x height image filled with c to w.
type encoder func(w io.Writer, width, height int, c color.RGBA) error

// encoders maps each supported format name to its encoder.
var encoders = map[string]encoder{
	"gif": func(w io.Writer, width, height int, c color.RGBA) error {
		// GIF has no alpha channel; use a single opaque palette entry.
		palette := color.Palette{color.RGBA{R: c.R, G: c.G, B: c.B, A: 0xFF}}
		return gif.EncodeAll(w, &gif.GIF{
			Image: []*image.Paletted{image.NewPaletted(image.Rect(0, 0, width, height), palette)},
			Delay: []int{0},
		})
	},
	"jpeg": func(w io.Writer, width, height int, c color.RGBA) error {
		// JPEG has no alpha channel; ignore the alpha component.
		c.A = 0xFF
		return jpeg.Encode(w, fill(width, height, c), nil)
	},
	"png": func(w io.Writer, width, height int, c color.RGBA) error {
		// Uniform-color images compress extremely well, so the extra
		// encoding time for maximum compression is worth the size savings.
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		return enc.Encode(w, fill(width, height, c))
	},
	"webp": func(w io.Writer, width, height int, c color.RGBA) error {
		return nativewebp.Encode(w, fill(width, height, c), nil)
	},
}

// extFormats maps each recognized file extension (lowercase) to a format name.
var extFormats = map[string]string{
	".gif":  "gif",
	".jpeg": "jpeg",
	".jpg":  "jpeg",
	".png":  "png",
	".webp": "webp",
}

var (
	formatList = strings.Join(slices.Sorted(maps.Keys(encoders)), ", ")
	extList    = strings.Join(slices.Sorted(maps.Keys(extFormats)), ", ")
)

// extListOr is extList with "or" before the last extension.
var extListOr = func() string {
	exts := slices.Sorted(maps.Keys(extFormats))
	return strings.Join(exts[:len(exts)-1], ", ") + ", or " + exts[len(exts)-1]
}()

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "(devel)"
}

func main() {
	os.Exit(run(filepath.Base(os.Args[0]), os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the command with the given program name and arguments,
// writing output to stdout and diagnostics to stderr. It returns the
// process exit code.
func run(name string, args []string, stdout, stderr io.Writer) int {
	// usageErr reports a usage error and returns the exit code.
	usageErr := func(format string, a ...any) int {
		fmt.Fprintf(stderr, format+"\n", a...)
		fmt.Fprintf(stderr, "Run '%s -h' for help.\n", name)
		return 1
	}

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
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
	fs.StringVar(&format, "f", "", "gif, jpeg, png (default), webp")
	fs.StringVar(&format, "format", "", "gif, jpeg, png (default), webp")
	fs.StringVar(&outputFile, "o", "", "write output to file")
	fs.StringVar(&outputFile, "output", "", "write output to file")
	fs.BoolVar(&asBase64, "b", false, "encode output as base64")
	fs.BoolVar(&asBase64, "base64", false, "encode output as base64")

	if err := fs.Parse(args); err != nil {
		return usageErr("Error: %v", err)
	}

	if showHelp {
		fmt.Fprintf(stdout, helpText, name)
		return 0
	}

	if showVersion {
		fmt.Fprintln(stdout, version())
		return 0
	}

	if fs.NArg() != 3 {
		return usageErr("Usage: %s [flags] <width> <height> <color>", name)
	}

	width, err := strconv.Atoi(fs.Arg(0))
	if err != nil || width < 1 || width > maxDim {
		return usageErr("Error: width must be an integer between 1 and %d", maxDim)
	}

	height, err := strconv.Atoi(fs.Arg(1))
	if err != nil || height < 1 || height > maxDim {
		return usageErr("Error: height must be an integer between 1 and %d", maxDim)
	}

	c, err := parseColor(fs.Arg(2))
	if err != nil {
		return usageErr("Error: %v", err)
	}

	if format == "" {
		format = "png"
		if outputFile != "" {
			ext := filepath.Ext(outputFile)
			f, ok := extFormats[strings.ToLower(ext)]
			switch {
			case ok:
				format = f
			case ext == "":
				return usageErr("Error: cannot infer format from %q: no file extension; use -f to specify the format or append one of %s", outputFile, extList)
			default:
				return usageErr("Error: cannot infer format from extension %q: use -f to specify the format or change the extension to one of %s", ext, extListOr)
			}
		}
	}

	format = strings.ToLower(format)
	encode, ok := encoders[format]
	if !ok {
		return usageErr("Error: unknown format %q: must be one of %s", format, formatList)
	}

	var buf bytes.Buffer
	if err := encode(&buf, width, height, c); err != nil {
		fmt.Fprintf(stderr, "Error encoding image: %v\n", err)
		return 1
	}

	out := buf.Bytes()
	if asBase64 {
		out = base64.StdEncoding.AppendEncode(nil, out)
		out = append(out, '\n')
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, out, 0o644); err != nil {
			fmt.Fprintf(stderr, "Error writing file: %v\n", err)
			return 1
		}
		return 0
	}
	if _, err := stdout.Write(out); err != nil {
		fmt.Fprintf(stderr, "Error writing output: %v\n", err)
		return 1
	}
	return 0
}

// fill returns a width x height image filled with c. It uses NRGBA
// (non-premultiplied alpha) because parsed colors are straight-alpha values;
// image.RGBA expects premultiplied values.
func fill(width, height int, c color.RGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	px := [4]uint8{c.R, c.G, c.B, c.A}
	for i := range width * height {
		copy(img.Pix[i*4:], px[:])
	}
	return img
}

// parseColor parses a CSS named color (case-insensitive) or a 3, 4, 6, or
// 8-digit hex color with an optional leading '#'.
func parseColor(s string) (color.RGBA, error) {
	if c, ok := namedColors[strings.ToLower(s)]; ok {
		return c, nil
	}
	h := strings.TrimPrefix(s, "#")
	// Expand shorthand: 3 → 6, 4 → 8.
	if len(h) == 3 || len(h) == 4 {
		exp := make([]byte, 0, len(h)*2)
		for i := range len(h) {
			exp = append(exp, h[i], h[i])
		}
		h = string(exp)
	}
	if len(h) != 6 && len(h) != 8 {
		return color.RGBA{}, fmt.Errorf("invalid color %q: expected 6 or 8 hex digits", s)
	}
	b, err := hex.DecodeString(h)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("invalid color %q", s)
	}
	c := color.RGBA{R: b[0], G: b[1], B: b[2], A: 0xFF}
	if len(b) == 4 {
		c.A = b[3]
	}
	return c, nil
}
