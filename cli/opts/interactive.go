package opts

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Prompter is the interface for interactive prompts.
// Implement this interface to use custom prompt libraries like bubbletea, survey, or promptui.
type Prompter interface {
	Input(message string, defaultValue string) (string, error)
	Password(message string) (string, error)
	Confirm(message string, defaultValue bool) (bool, error)
	Select(message string, options []string, defaultIndex int) (int, error)
	MultiSelect(message string, options []string, defaults []int) ([]int, error)
}

// DefaultPrompter provides a simple stdin-based prompter implementation.
type DefaultPrompter struct {
	reader io.Reader
	writer io.Writer
}

// NewDefaultPrompter creates a new default prompter using stdin/stdout.
func NewDefaultPrompter() *DefaultPrompter {
	return &DefaultPrompter{
		reader: os.Stdin,
		writer: os.Stdout,
	}
}

// NewPrompterWithIO creates a prompter with custom reader/writer for testing.
func NewPrompterWithIO(r io.Reader, w io.Writer) *DefaultPrompter {
	return &DefaultPrompter{
		reader: r,
		writer: w,
	}
}

func (p *DefaultPrompter) Input(message string, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(p.writer, "%s [%s]: ", message, defaultValue)
	} else {
		fmt.Fprintf(p.writer, "%s: ", message)
	}

	scanner := bufio.NewScanner(p.reader)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return defaultValue, nil
		}
		return input, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return defaultValue, nil
}

func (p *DefaultPrompter) Password(message string) (string, error) {
	fmt.Fprintf(p.writer, "%s: ", message)

	scanner := bufio.NewScanner(p.reader)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

func (p *DefaultPrompter) Confirm(message string, defaultValue bool) (bool, error) {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}
	fmt.Fprintf(p.writer, "%s [%s]: ", message, defaultStr)

	scanner := bufio.NewScanner(p.reader)
	if scanner.Scan() {
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if input == "" {
			return defaultValue, nil
		}
		return input == "y" || input == "yes", nil
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}
	return defaultValue, nil
}

func (p *DefaultPrompter) Select(message string, options []string, defaultIndex int) (int, error) {
	fmt.Fprintf(p.writer, "%s\n", message)
	for i, opt := range options {
		marker := "  "
		if i == defaultIndex {
			marker = "> "
		}
		fmt.Fprintf(p.writer, "%s%d) %s\n", marker, i+1, opt)
	}
	fmt.Fprintf(p.writer, "Enter number [%d]: ", defaultIndex+1)

	scanner := bufio.NewScanner(p.reader)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return defaultIndex, nil
		}
		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(options) {
			return -1, fmt.Errorf("invalid selection: %s", input)
		}
		return num - 1, nil
	}
	if err := scanner.Err(); err != nil {
		return -1, err
	}
	return defaultIndex, nil
}

func (p *DefaultPrompter) MultiSelect(message string, options []string, defaults []int) ([]int, error) {
	defaultSet := make(map[int]bool)
	for _, d := range defaults {
		defaultSet[d] = true
	}

	fmt.Fprintf(p.writer, "%s (comma-separated numbers)\n", message)
	for i, opt := range options {
		marker := "[ ]"
		if defaultSet[i] {
			marker = "[x]"
		}
		fmt.Fprintf(p.writer, "%s %d) %s\n", marker, i+1, opt)
	}

	defaultStr := ""
	for i, d := range defaults {
		if i > 0 {
			defaultStr += ","
		}
		defaultStr += strconv.Itoa(d + 1)
	}
	if defaultStr != "" {
		fmt.Fprintf(p.writer, "Enter numbers [%s]: ", defaultStr)
	} else {
		fmt.Fprintf(p.writer, "Enter numbers: ")
	}

	scanner := bufio.NewScanner(p.reader)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return defaults, nil
		}
		parts := strings.Split(input, ",")
		result := make([]int, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			num, err := strconv.Atoi(part)
			if err != nil || num < 1 || num > len(options) {
				return nil, fmt.Errorf("invalid selection: %s", part)
			}
			result = append(result, num-1)
		}
		return result, nil
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return defaults, nil
}

// InteractiveOpts holds configuration for interactive parsing.
type InteractiveOpts struct {
	Prompter Prompter
	SkipSet  bool // Skip prompting for values that are already set
}

// ParseInteractive prompts for values that weren't provided via command line.
func (n *node) ParseInteractive(opts *InteractiveOpts) error {
	if opts == nil {
		opts = &InteractiveOpts{}
	}
	if opts.Prompter == nil {
		opts.Prompter = NewDefaultPrompter()
	}

	for _, item := range n.flags() {
		if opts.SkipSet && item.set() {
			continue
		}

		if err := promptForItem(opts.Prompter, item); err != nil {
			return fmt.Errorf("prompt for '%s': %w", item.name, err)
		}
	}

	for _, item := range n.args {
		if opts.SkipSet && item.set() {
			continue
		}

		if err := promptForItem(opts.Prompter, item); err != nil {
			return fmt.Errorf("prompt for arg '%s': %w", item.name, err)
		}
	}

	return nil
}

func promptForItem(p Prompter, item *item) error {
	message := item.name
	if item.help != "" {
		message = item.help
	}

	if len(item.enum) > 0 {
		idx, err := p.Select(message, item.enum, 0)
		if err != nil {
			return err
		}
		if idx >= 0 && idx < len(item.enum) {
			return item.Set(item.enum[idx])
		}
		return nil
	}

	if item.noarg {
		confirmed, err := p.Confirm(message, false)
		if err != nil {
			return err
		}
		if confirmed {
			return item.Set("true")
		}
		return item.Set("false")
	}

	value, err := p.Input(message, item.defstr)
	if err != nil {
		return err
	}
	if value != "" {
		return item.Set(value)
	}
	return nil
}

// Interactive returns an option to enable interactive mode.
func Interactive(prompter Prompter) func(*node) {
	return func(n *node) {
		n.prompter = prompter
	}
}
