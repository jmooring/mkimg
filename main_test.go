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
	"image/color"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/scripts",
	})
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
