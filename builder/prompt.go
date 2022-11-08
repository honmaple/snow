package builder

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
)

type (
	Prompt interface {
		Excute(*bufio.Reader) error
	}
	Prompts []Prompt
)

func (ps Prompts) Excute(r *bufio.Reader) error {
	for _, p := range ps {
		if err := p.Excute(r); err != nil {
			return err
		}
	}
	return nil
}

type PromptString struct {
	Usage       string
	Value       string
	Required    bool
	FilePath    bool
	Destination *string
}

func (p *PromptString) Excute(r *bufio.Reader) error {
	fmt.Printf(p.Usage)
	if p.Value != "" {
		fmt.Printf("[%s] ", p.Value)
	}
	input, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if p.Required && input == "" {
		if p.Value == "" {
			fmt.Println("The input is required")
			return p.Excute(r)
		}
	}
	if input == "" {
		input = p.Value
	}
	if p.FilePath {
		if n := filepath.Clean(input); n != input {
			return fmt.Errorf("The input is not a valid path")
		}
	}
	if p.Destination != nil {
		*p.Destination = input
	}
	return nil
}

type PromptBool struct {
	Name        string
	Usage       string
	Value       bool
	Destination *bool
}

func (p *PromptBool) Excute(r *bufio.Reader) error {
	fmt.Printf(p.Usage)
	if p.Value {
		fmt.Printf("[Y/n] ")
	} else {
		fmt.Printf("[y/N] ")
	}
	input, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	input = strings.ToUpper(strings.TrimSpace(input))
	if input != "" && input != "Y" && input != "N" {
		fmt.Println("The input must be y or n")
		return p.Excute(r)
	}
	if p.Destination != nil {
		if input == "" {
			*p.Destination = p.Value
		} else {
			*p.Destination = input == "Y"
		}
	}
	return nil
}
