package cli

import (
	"bufio"
	"fmt"
	"os"
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

type PromptInt struct {
	Usage       string
	Value       int64
	Required    bool
	Destination *int64
}

func (p *PromptInt) Excute(r *bufio.Reader) error {
	fmt.Printf(p.Usage)
	if p.Value != 0 {
		fmt.Printf("[%d] ", p.Value)
	}
	input, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if p.Required && input == "" {
		if p.Value == 0 {
			fmt.Println("The input is required")
			return p.Excute(r)
		}
	}
	if input == "" {
		if p.Destination != nil {
			*p.Destination = p.Value
		}
		return nil
	}
	var value int64
	if _, err := fmt.Sscanf(input, "%d", &value); err != nil {
		fmt.Println("The input must be a valid integer")
		return p.Excute(r)
	}
	if p.Destination != nil {
		*p.Destination = value
	}
	return nil
}

// PromptFilePath 用于文件路径输入
type PromptFilePath struct {
	Usage       string
	Value       string
	Required    bool
	MustExist   bool // 是否要求路径必须存在
	IsDir       bool // 是否要求为目录（否则为文件）
	Destination *string
}

func (p *PromptFilePath) Excute(r *bufio.Reader) error {
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
	cleaned := filepath.Clean(input)
	if cleaned == "" {
		fmt.Println("The input is not a valid path")
		return p.Excute(r)
	}
	if p.MustExist {
		info, err := os.Stat(cleaned)
		if err != nil {
			fmt.Println("The path does not exist")
			return p.Excute(r)
		}
		if p.IsDir && !info.IsDir() {
			fmt.Println("The path must be a directory")
			return p.Excute(r)
		}
		if !p.IsDir && info.IsDir() {
			fmt.Println("The path must be a file")
			return p.Excute(r)
		}
	}
	if p.Destination != nil {
		*p.Destination = cleaned
	}
	return nil
}
