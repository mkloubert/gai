// MIT License
//
// Copyright (c) 2025 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package types

// GAIRCFile stores the structure of an `.gairc.yaml` file.
type GAIRCFile struct {
	// Defaults stores default setting.
	Defaults GAIRCFileDefaults `yaml:"defaults,omitempty"`
}

// GAIRCFileDefaults stores `defaults` parts in a `GAIRCFile` object.
type GAIRCFileDefaults struct {
	// Flags stores default settings for CLI flags.
	Flags GAIRCFileDefaultsFlags `yaml:"flags,omitempty"`
}

// GAIRCFileDefaultsFlags stores `flags` parts in a `GAIRCFileDefaults` object.
type GAIRCFileDefaultsFlags struct {
	// File stores default settings for CLI flag `--file`.
	File []string `yaml:"file,omitempty"`
	// Files stores default settings for CLI flag `--files`.
	Files []string `yaml:"files,omitempty"`
}
