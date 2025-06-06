// SPDX-License-Identifier: MIT
//
// box.go - Ported to Go by AlphaLynx <alphalynx@protonmail.com>
// Original author: Dave Eddy <dave@daveeddy.com>
//
// This is a Go port of the original bash box tool by Dave Eddy, under the MIT License.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
)

var (
	themeUnicode = map[string]string{
		"WE": "━", "NS": "┃", "NW": "┏", "NE": "┓", "SW": "┗", "SE": "┛",
	}
	themeAscii = map[string]string{
		"WE": "-", "NS": "|", "NW": "+", "NE": "+", "SW": "+", "SE": "+",
	}
	themePlain = map[string]string{
		"WE": " ", "NS": " ", "NW": " ", "NE": " ", "SW": " ", "SE": " ",
	}
)

type Theme map[string]string

func getTheme(name string) Theme {
	switch name {
	case "unicode":
		return themeUnicode
	case "ascii":
		return themeAscii
	case "plain":
		return themePlain
	default:
		return themeUnicode
	}
}

func parseColor(name string) *color.Color {
	switch strings.ToLower(name) {
	case "black":
		return color.New(color.FgBlack)
	case "red":
		return color.New(color.FgRed)
	case "green":
		return color.New(color.FgGreen)
	case "yellow":
		return color.New(color.FgYellow)
	case "blue":
		return color.New(color.FgBlue)
	case "magenta":
		return color.New(color.FgMagenta)
	case "cyan":
		return color.New(color.FgCyan)
	case "white":
		return color.New(color.FgWhite)
	case "gray", "bright_black":
		return color.New(color.FgHiBlack)
	case "bright_red":
		return color.New(color.FgHiRed)
	case "bright_green":
		return color.New(color.FgHiGreen)
	case "bright_yellow":
		return color.New(color.FgHiYellow)
	case "bright_blue":
		return color.New(color.FgHiBlue)
	case "bright_magenta":
		return color.New(color.FgHiMagenta)
	case "bright_cyan":
		return color.New(color.FgHiCyan)
	case "bright_white":
		return color.New(color.FgHiWhite)
	default:
		return color.New(color.Reset)
	}
}

func repeatChar(char string, n int) string {
	return strings.Repeat(char, n)
}

func stripAnsi(str string) string {
	// Remove ANSI escape codes
	var result strings.Builder
	inEscape := false
	for _, r := range str {
		if r == 0x1b {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

func maxLineWidth(lines []string) int {
	max := 0
	for _, line := range lines {
		l := utf8.RuneCountInString(stripAnsi(line))
		if l > max {
			max = l
		}
	}
	return max
}

func createBox(
	lines []string,
	boxColor *color.Color,
	contentColor *color.Color,
	boxTitle string,
	titleColor *color.Color,
	theme Theme,
	vpadding, hpadding int,
) []string {
	maxWidth := maxLineWidth(lines) + 2*hpadding
	if boxTitle != "" {
		titleLen := utf8.RuneCountInString(stripAnsi(boxTitle))
		if titleLen > maxWidth {
			maxWidth = titleLen
		}
	}
	border := func(char string) string {
		if boxColor != nil {
			return boxColor.Sprint(char)
		}
		return char
	}
	// Top border
	top := border(theme["NW"])
	if boxTitle != "" {
		title := boxTitle
		if titleColor != nil {
			title = titleColor.Sprint(boxTitle)
		}
		top += title
		rest := maxWidth - utf8.RuneCountInString(stripAnsi(boxTitle))
		if boxColor != nil {
			top += boxColor.Sprint(repeatChar(theme["WE"], rest))
		} else {
			top += repeatChar(theme["WE"], rest)
		}
	} else {
		if boxColor != nil {
			top += boxColor.Sprint(repeatChar(theme["WE"], maxWidth))
		} else {
			top += repeatChar(theme["WE"], maxWidth)
		}
	}
	top += border(theme["NE"])
	boxLines := []string{top}
	// Vertical padding (top)
	for i := 0; i < vpadding; i++ {
		boxLines = append(boxLines, fmt.Sprintf("%s%s%s", border(theme["NS"]), strings.Repeat(" ", maxWidth), border(theme["NS"])))
	}
	// Content lines
	for _, line := range lines {
		stripped := stripAnsi(line)
		padLeft := strings.Repeat(" ", hpadding)
		padRight := strings.Repeat(" ", hpadding)
		content := line
		if contentColor != nil {
			content = contentColor.Sprint(line)
		}
		totalPadding := maxWidth - utf8.RuneCountInString(stripped)
		boxLines = append(boxLines, fmt.Sprintf("%s%s%s%s%s%s", border(theme["NS"]), padLeft, content, padRight, strings.Repeat(" ", totalPadding-2*hpadding), border(theme["NS"])))
	}
	// Vertical padding (bottom)
	for i := 0; i < vpadding; i++ {
		boxLines = append(boxLines, fmt.Sprintf("%s%s%s", border(theme["NS"]), strings.Repeat(" ", maxWidth), border(theme["NS"])))
	}
	// Bottom border
	bottom := border(theme["SW"])
	if boxColor != nil {
		bottom += boxColor.Sprint(repeatChar(theme["WE"], maxWidth))
	} else {
		bottom += repeatChar(theme["WE"], maxWidth)
	}
	bottom += border(theme["SE"])
	boxLines = append(boxLines, bottom)
	return boxLines
}

func createNestedBoxes(
	lines []string,
	depth int,
	boxColors, titleColors, boxTitles []string,
	theme Theme,
	vpadding, hpadding int,
	contentColor *color.Color,
) []string {
	for i := depth - 1; i >= 0; i-- {
		var boxColor, titleColor *color.Color
		var boxTitle string
		if len(boxColors) > 0 {
			if i < len(boxColors) {
				boxColor = parseColor(boxColors[i])
			} else {
				boxColor = parseColor(boxColors[0])
			}
		}
		if len(titleColors) > 0 {
			if i < len(titleColors) {
				titleColor = parseColor(titleColors[i])
			} else {
				titleColor = parseColor(titleColors[0])
			}
		}
		if len(boxTitles) > 0 {
			if i < len(boxTitles) {
				boxTitle = boxTitles[i]
			} else {
				boxTitle = ""
			}
		}
		lines = createBox(lines, boxColor, contentColor, boxTitle, titleColor, theme, vpadding, hpadding)
	}
	return lines
}

func main() {
	var (
		number      = flag.Int("n", 1, "Number of nested boxes to create, defaults to 1")
		title       = flag.String("t", "", "Comma-separated titles for each box")
		boxColor    = flag.String("bc", "", "Comma-separated colors for each box border")
		titleColor  = flag.String("tc", "", "Comma-separated colors for each box title")
		centerColor = flag.String("cc", "", "Color for the center text")
		vpadding    = flag.Int("vp", 0, "Vertical padding")
		hpadding    = flag.Int("hp", 0, "Horizontal padding")
		themeName   = flag.String("T", "unicode", "Theme to use (unicode, ascii, plain)")
		sep         = flag.String("s", "", "Separator character (not used in this version)")
		mode        = flag.String("m", "", "Color mode (not implemented in this version)")
	)
	// Long options
	flag.IntVar(number, "number", 1, "Number of nested boxes to create, defaults to 1")
	flag.StringVar(title, "title", "", "Comma-separated titles for each box")
	flag.StringVar(boxColor, "box-color", "", "Comma-separated colors for each box border")
	flag.StringVar(titleColor, "title-color", "", "Comma-separated colors for each box title")
	flag.StringVar(centerColor, "center-color", "", "Color for the center text")
	flag.IntVar(vpadding, "vpadding", 0, "Vertical padding")
	flag.IntVar(hpadding, "hpadding", 0, "Horizontal padding")
	flag.StringVar(themeName, "theme", "unicode", "Theme to use (unicode, ascii, plain)")
	flag.StringVar(sep, "sep", "", "Separator character (not used in this version)")
	flag.StringVar(mode, "mode", "", "Color mode (not implemented in this version)")
	flag.Parse()

	// Read stdin
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "[FATAL] Error reading stdin:", err)
		os.Exit(1)
	}

	// Parse comma-separated options
	boxTitles := []string{}
	if *title != "" {
		boxTitles = strings.Split(*title, ",")
	}
	boxColors := []string{}
	if *boxColor != "" {
		boxColors = strings.Split(*boxColor, ",")
	}
	titleColors := []string{}
	if *titleColor != "" {
		titleColors = strings.Split(*titleColor, ",")
	}
	var contentColor *color.Color
	if *centerColor != "" {
		contentColor = parseColor(*centerColor)
	}
	depth := *number
	// Pad lists to match depth
	if len(boxTitles) < depth {
		pad := make([]string, depth-len(boxTitles))
		for i := range pad {
			pad[i] = ""
		}
		boxTitles = append(pad, boxTitles...)
	}
	if len(boxColors) < depth && len(boxColors) > 0 {
		pad := make([]string, depth-len(boxColors))
		for i := range pad {
			pad[i] = boxColors[0]
		}
		boxColors = append(pad, boxColors...)
	}
	if len(titleColors) < depth && len(titleColors) > 0 {
		pad := make([]string, depth-len(titleColors))
		for i := range pad {
			pad[i] = titleColors[0]
		}
		titleColors = append(pad, titleColors...)
	}

	// Build nested boxes
	theme := getTheme(*themeName)
	resultLines := createNestedBoxes(
		lines,
		depth,
		boxColors,
		titleColors,
		boxTitles,
		theme,
		*vpadding,
		*hpadding,
		contentColor,
	)
	for _, l := range resultLines {
		fmt.Println(l)
	}
}
