// SPDX-License-Identifier: MIT
//
// box.go - Ported to Go by AlphaLynx <alphalynx@protonmail.com>
// Original author: Dave Eddy <dave@daveeddy.com>
//
// This is a Go port of the original bash box tool by Dave Eddy with additional features, under the MIT License.

package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
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

// ColorTheme represents a color theme mode
type ColorTheme struct {
	name       string
	startColor int
}

// getNextColor returns the next color based on the theme and index
func getNextColor(theme *ColorTheme, index int) int {
	switch theme.name {
	case "random":
		return rand.Intn(216)
	case "gradient":
		return (theme.startColor + index) % 216
	case "rainbow":
		return index % 216
	case "pride":
		// Traditional rainbow pride flag colors
		prideColors := []int{196, 208, 226, 46, 21, 129}
		return prideColors[index%len(prideColors)]
	case "trans":
		// Trans flag colors (light blue, pink, white, pink, light blue)
		transColors := []int{51, 213, 15, 213, 51}
		return transColors[index%len(transColors)]
	case "bi":
		// Bisexual flag colors (pink, purple, blue)
		biColors := []int{213, 129, 21}
		return biColors[index%len(biColors)]
	case "pan":
		// Pansexual flag colors (pink, yellow, blue)
		panColors := []int{213, 226, 21}
		return panColors[index%len(panColors)]
	case "nb":
		// Non-binary flag colors (yellow, white, purple, black)
		nbColors := []int{226, 15, 129, 0}
		return nbColors[index%len(nbColors)]
	default:
		return theme.startColor
	}
}

// newColorTheme creates a new color theme with the given name
func newColorTheme(name string) *ColorTheme {
	theme := &ColorTheme{
		name:       name,
		startColor: rand.Intn(216),
	}
	return theme
}

// getColorFromTheme returns a color.Color based on the theme and index
func getColorFromTheme(theme *ColorTheme, index int) *color.Color {
	colorCode := getNextColor(theme, index)
	// Use ANSI 256-color sequence: ESC[38;5;<n>m for foreground
	return color.New(color.Attribute(38), color.Attribute(5), color.Attribute(colorCode))
}

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

// parseColor parses a color name into a color.Color
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
		// Try to parse as a number for 256-color mode
		if num, err := strconv.Atoi(name); err == nil {
			return color.New(color.FgHiBlack + color.Attribute(num))
		}
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
	colorTheme *ColorTheme,
) []string {
	for i := 0; i < depth; i++ { // OUTERMOST is i=0
		var boxColor, titleColor *color.Color
		var boxTitle string
		if colorTheme != nil {
			// Cycle box border color through the color theme, outermost is index 0
			boxColor = getColorFromTheme(colorTheme, i)
		} else if len(boxColors) > 0 {
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
		// Only the innermost box gets the content color
		var thisContentColor *color.Color
		if i == depth-1 {
			thisContentColor = contentColor
		}
		lines = createBox(lines, boxColor, thisContentColor, boxTitle, titleColor, theme, vpadding, hpadding)
	}
	return lines
}

var rootCmd = &cobra.Command{
	Use:   "box",
	Short: "A tool for creating text boxes in the terminal",
	Long: `Box is a CLI tool for creating text boxes in the terminal.
It supports various themes, colors, and nested boxes.

Examples:
  # Create a single box with a title
  echo "Hello, world!" | box -t "My Title"

  # Create a nested box with a title and colors
  echo "Hello, world!" | box -t "My Title" -b "red" -c "blue" -n 2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read stdin
		var lines []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return err
		}

		// Parse comma-separated options
		boxTitles := []string{}
		if title != "" {
			boxTitles = strings.Split(title, ",")
		}
		boxColors := []string{}
		if boxColor != "" {
			boxColors = strings.Split(boxColor, ",")
		}
		titleColors := []string{}
		if titleColor != "" {
			titleColors = strings.Split(titleColor, ",")
		}
		var contentColor *color.Color
		if centerColor != "" {
			contentColor = parseColor(centerColor)
		}
		depth := number

		// Initialize color theme if mode is specified
		var colorTheme *ColorTheme
		if mode != "" {
			colorTheme = newColorTheme(mode)
		}

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
		theme := getTheme(themeName)
		resultLines := createNestedBoxes(
			lines,
			depth,
			boxColors,
			titleColors,
			boxTitles,
			theme,
			vpadding,
			hpadding,
			contentColor,
			colorTheme,
		)
		for _, l := range resultLines {
			fmt.Println(l)
		}

		return nil
	},
}

var (
	number      int
	title       string
	boxColor    string
	titleColor  string
	centerColor string
	vpadding    int
	hpadding    int
	themeName   string
	sep         string
	mode        string
)

var docsCmd = &cobra.Command{
    Use:    "docs man",
    Short:  "Generate man page",
    Hidden: true, // Hide from help output
    Args:   cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        if args[0] != "man" {
            return fmt.Errorf("only man page generation is supported")
        }

        header := &doc.GenManHeader{
            Title:   "BOX",
            Section: "1",
            Source:  "Box Version 0.1.0",
            Manual:  "User Commands",
        }

        // write the manpage to stdout instead of to box.1
        if err := doc.GenMan(cmd.Root(), header, os.Stdout); err != nil {
            return err
        }
        return nil
    },
}

var completionCmd = &cobra.Command{
	Use:    "completion [bash|zsh|fish|powershell]",
	Short:  "Generate shell completions",
	Hidden: true, // Hide from help output
	Long: "Generate shell completions for bash, zsh, fish, or powershell.",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	rootCmd.Flags().IntVarP(&number, "number", "n", 1, "Number of nested boxes")
	rootCmd.Flags().StringVarP(&title, "title", "t", "", "Box titles (comma-separated)")
	rootCmd.Flags().StringVarP(&boxColor, "box-color", "b", "", "Box border colors (comma-separated)")
	rootCmd.Flags().StringVarP(&titleColor, "title-color", "c", "", "Title colors (comma-separated)")
	rootCmd.Flags().StringVarP(&centerColor, "center-color", "C", "", "Center text color")
	rootCmd.Flags().IntVarP(&vpadding, "vpadding", "v", 0, "Vertical padding")
	rootCmd.Flags().IntVarP(&hpadding, "hpadding", "H", 0, "Horizontal padding")
	rootCmd.Flags().StringVarP(&themeName, "theme", "T", "unicode", "Theme: unicode, ascii, plain")
	rootCmd.Flags().StringVarP(&sep, "sep", "s", "", "Separator char (unused)")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "", "Color mode (random, gradient, rainbow, pride, trans, bi, pan, nb)")

	// Register commands
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(docsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
