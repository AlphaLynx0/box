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

// resolveTextInput determines the source of text input and returns processed lines
// It prioritizes stdin input when available and falls back to command line arguments
func resolveTextInput(args []string) ([]string, error) {
	// Check if stdin has available data
	stat, err := os.Stdin.Stat()
	if err != nil {
		// Handle file stat errors gracefully with fallback to argument processing
		// Log the error but continue with argument processing
		fmt.Fprintf(os.Stderr, "Warning: unable to check stdin status (%v), using command line arguments\n", err)
		lines := processArguments(args)
		if len(lines) == 0 {
			return nil, fmt.Errorf("no input provided: unable to read from stdin and no command line arguments given")
		}
		return lines, nil
	}

	// Check if stdin has data (pipe or redirect)
	// When stdin is a character device (terminal), there's no piped data
	// When stdin is NOT a character device, it means data is piped/redirected
	hasStdinData := (stat.Mode() & os.ModeCharDevice) == 0

	if hasStdinData {
		// Read from stdin (preserve existing behavior)
		var lines []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			// Provide clear error message for stdin reading failures
			return nil, fmt.Errorf("failed to read from stdin: %v", err)
		}
		return lines, nil
	}

	// No stdin data available, process command line arguments
	lines := processArguments(args)
	
	// Handle empty input scenarios (no stdin, no arguments)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no input provided: please provide text via stdin or command line arguments")
	}
	
	return lines, nil
}

// processArguments converts command line arguments into text lines
func processArguments(args []string) []string {
	var lines []string
	for _, arg := range args {
		// Skip empty arguments to prevent empty lines
		if arg == "" {
			continue
		}
		
		// Handle newline characters within arguments
		// Split on literal \n sequences to create multiple lines
		argLines := strings.Split(arg, "\\n")
		lines = append(lines, argLines...)
	}
	return lines
}

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
	usedColors map[int]bool
	colorIndex int
}

// getNextColor returns the next color based on the theme and index, avoiding repetition
func getNextColor(theme *ColorTheme, index int) int {
	var candidateColors []int
	
	switch theme.name {
	case "random":
		// For random, pick from unused colors or reset if all used
		if len(theme.usedColors) >= 216 {
			theme.usedColors = make(map[int]bool)
		}
		for {
			color := rand.Intn(216)
			if !theme.usedColors[color] {
				theme.usedColors[color] = true
				return color
			}
		}
	case "gradient":
		// For gradient, continue from where we left off
		color := (theme.startColor + theme.colorIndex) % 216
		theme.colorIndex++
		return color
	case "rainbow":
		// For rainbow, cycle through all 216 colors
		color := theme.colorIndex % 216
		theme.colorIndex++
		return color
	case "pride":
		// Traditional rainbow pride flag colors
		candidateColors = []int{196, 208, 226, 46, 21, 129}
	case "trans":
		// Trans flag colors (light blue, pink, white) - removed duplicates
		candidateColors = []int{51, 213, 15}
	case "bi":
		// Bisexual flag colors (pink, purple, blue)
		candidateColors = []int{213, 129, 21}
	case "pan":
		// Pansexual flag colors (pink, yellow, blue)
		candidateColors = []int{213, 226, 21}
	case "nb":
		// Non-binary flag colors (yellow, white, purple, black)
		candidateColors = []int{226, 15, 129, 0}
	default:
		return theme.startColor
	}
	
	// For flag themes, cycle through colors without repetition
	if len(candidateColors) > 0 {
		// Reset used colors if we've used all available colors in this theme
		if len(theme.usedColors) >= len(candidateColors) {
			theme.usedColors = make(map[int]bool)
		}
		
		// Find next unused color in the theme
		for i := 0; i < len(candidateColors); i++ {
			colorIndex := (theme.colorIndex + i) % len(candidateColors)
			color := candidateColors[colorIndex]
			if !theme.usedColors[color] {
				theme.usedColors[color] = true
				theme.colorIndex = (colorIndex + 1) % len(candidateColors)
				return color
			}
		}
		
		// Fallback: if all colors somehow used, reset and return first
		theme.usedColors = make(map[int]bool)
		color := candidateColors[0]
		theme.usedColors[color] = true
		theme.colorIndex = 1
		return color
	}
	
	return theme.startColor
}

// newColorTheme creates a new color theme with the given name
func newColorTheme(name string) *ColorTheme {
	theme := &ColorTheme{
		name:       name,
		startColor: rand.Intn(216),
		usedColors: make(map[int]bool),
		colorIndex: 0,
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
    Use:   "box [text...]",
    Short: "Create box around text",
    Long: `Box is a CLI tool for creating text boxes in the terminal.
It supports various themes, colors, and nested boxes.

Examples:
  echo "Hello, world!" | box -t "My Title"
  echo "Hello, world!" | box -t "My Title" -b "red" -c "blue" -n 2
  box "Hello, world!" -t "My Title"
  box "Line 1" "Line 2" "Line 3"`,
    Args: cobra.ArbitraryArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1) resolve text input (stdin or arguments)
        lines, err := resolveTextInput(args)
        if err != nil {
            // Check if this is a "no input" error and display usage
            if strings.Contains(err.Error(), "no input provided") {
                cmd.Usage()
                fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
                return nil // Don't return error to avoid double display
            }
            return err
        }

        depth := number

        // 2) split comma-lists
        titles := []string{}
        if title != "" {
            titles = strings.Split(title, ",")
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

        // 3) validate each list: must be either length 1 or exactly depth
        validate := func(name string, list []string) error {
            if len(list) == 0 {
                return nil // flag not provided → skip
            }
            if len(list) != 1 && len(list) != depth {
                return fmt.Errorf(
                    "%s must have either 1 value or %d values, but got %d",
                    name, depth, len(list),
                )
            }
            return nil
        }
        if err := validate("-t/--title", titles); err != nil {
            return err
        }
        if err := validate("-b/--box-colors", boxColors); err != nil {
            return err
        }
        if err := validate("-c/--title-colors", titleColors); err != nil {
            return err
        }

        // 4) normalize single-element lists to full length
        expand := func(list []string, defaultVal string) []string {
            if len(list) == 0 {
                // nothing provided → produce empty slots
                return make([]string, depth)
            }
            if len(list) == 1 {
                // pad to depth
                expanded := make([]string, depth)
                for i := 0; i < depth; i++ {
                    expanded[i] = list[0]
                }
                return expanded
            }
            // already exactly depth
            return list
        }
        boxTitles := expand(titles, "")
        boxColors  = expand(boxColors, "")
        titleColors = expand(titleColors, "")

        // 5) init theme if needed
        var colorTheme *ColorTheme
        if mode != "" {
            colorTheme = newColorTheme(mode)
        }

        // 6) draw
        theme := getTheme(themeName)
        result := createNestedBoxes(
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

        for _, l := range result {
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
