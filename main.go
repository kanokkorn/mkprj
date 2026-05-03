package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

type Config struct {
	ProjectName string
	GenMode     string // "makefile" or "project"
	Language    string // "c", "cpp", "go"
	Ext         string
	CC          string
	Flags       []string
	Install     bool
	TargetDir   string
}

func main() {
	var cfg Config

	// Header
	fmt.Println("\033[38;5;205m\033[1mmkprj - Interactive boilerplate and Makefile generator\033[0m\n")

	// Step 1: Base Configuration
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Placeholder("my-project").
				Value(&cfg.ProjectName).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("project name cannot be empty")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Generation Mode").
				Options(
					huh.NewOption("Only Makefile", "makefile"),
					huh.NewOption("Whole Project Template", "project"),
				).
				Value(&cfg.GenMode),

			huh.NewSelect[string]().
				Title("Programming Language").
				Options(
					huh.NewOption("C", "c"),
					huh.NewOption("C++", "cpp"),
					huh.NewOption("Go", "go"),
				).
				Value(&cfg.Language),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	// Set language-specific defaults
	switch cfg.Language {
	case "c":
		cfg.Ext, cfg.CC = "c", "cc"
	case "cpp":
		cfg.Ext, cfg.CC = "cpp", "c++"
	case "go":
		cfg.Ext = "go"
	}

	// Step 2: Language-Specific Flags & Install Options
	if err := runOptionsForm(&cfg); err != nil {
		log.Fatal(err)
	}

	// Step 3: Directory Preparation
	cfg.TargetDir = "."
	if cfg.GenMode == "project" {
		cfg.TargetDir = filepath.Join(".", cfg.ProjectName)
		if _, err := os.Stat(cfg.TargetDir); err == nil {
			log.Fatalf("❌ Error: Directory %s already exists!", cfg.TargetDir)
		}
		if err := os.MkdirAll(cfg.TargetDir, 0755); err != nil {
			log.Fatal(err)
		}
	}

	// Step 4: Generation Logic with Spinner
	err := spinner.New().
		Title(" Generating project files...").
		Action(func() {
			if cfg.GenMode == "project" {
				_ = createSourceFile(cfg)
				if cfg.Language == "go" {
					_ = initGoMod(cfg.TargetDir, cfg.ProjectName)
				}
			}
			_ = generateMakefile(cfg)
		}).
		Run()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n✅ \033[32m\033[1mSuccess!\033[0m Project created in: %s\n", cfg.TargetDir)
}

func runOptionsForm(cfg *Config) error {
	var groups []*huh.Group

	if cfg.Language == "go" {
		groups = append(groups, huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Go Build Flags").
				Options(
					huh.NewOption("Strip Binary (-s -w)", "-s -w"),
					huh.NewOption("Static Linking", "-extldflags \"-static\""),
				).
				Value(&cfg.Flags),
		))
	} else {
		groups = append(groups, huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Compiler Flags").
				Options(
					huh.NewOption("Warnings (-Wall -Wextra)", "-Wall -Wextra"),
					huh.NewOption("ISO Standards (-pedantic)", "-pedantic"),
					huh.NewOption("Optimization (-Ofast)", "-Ofast"),
					huh.NewOption("Math Library (-lm)", "-lm"),
					huh.NewOption("Pthread (-lpthread)", "-lpthread"),
				).
				Value(&cfg.Flags),
		))
	}

	groups = append(groups, huh.NewGroup(
		huh.NewConfirm().
			Title("Add install/uninstall targets?").
			Value(&cfg.Install),
	))

	return huh.NewForm(groups...).WithTheme(huh.ThemeCatppuccin()).Run()
}

func createSourceFile(cfg Config) error {
	content := map[string]string{
		"go":  "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n",
		"c":   "#include <stdio.h>\n\nint main() {\n    printf(\"Hello, World!\\n\");\n    return 0;\n}\n",
		"cpp": "#include <iostream>\n\nint main() {\n    std::cout << \"Hello, World!\" << std::endl;\n    return 0;\n}\n",
	}
	return os.WriteFile(filepath.Join(cfg.TargetDir, "main."+cfg.Ext), []byte(content[cfg.Language]), 0644)
}

func initGoMod(dir, modName string) error {
	cmd := exec.Command("go", "mod", "init", modName)
	cmd.Dir = dir
	return cmd.Run()
}

func generateMakefile(cfg Config) error {
	var sb strings.Builder
	flags := strings.Join(cfg.Flags, " ")

	sb.WriteString(fmt.Sprintf("PROJECT_NAME := %s\n", cfg.ProjectName))
	
	if cfg.Language == "go" {
		sb.WriteString(fmt.Sprintf("LDFLAGS      := %s\n\nbuild:\n\tgo mod tidy\n\tgo build -ldflags=\"$(LDFLAGS)\" -o $(PROJECT_NAME) main.go\n", flags))
	} else {
		sb.WriteString(fmt.Sprintf("CC      := %s\nCFLAGS  := %s\nSRC     := $(wildcard *.%s)\n\nbuild: \n\t$(CC) $(SRC) -o $(PROJECT_NAME) $(CFLAGS)\n", cfg.CC, flags, cfg.Ext))
	}

	if cfg.Install {
		sb.WriteString("\ninstall: build\n\tinstall -m 755 $(PROJECT_NAME) /usr/local/bin/\n\nuninstall:\n\trm -f /usr/local/bin/$(PROJECT_NAME)\n")
	}

	sb.WriteString("\nclean:\n\trm -f $(PROJECT_NAME) *.o\n\n.PHONY: build clean install uninstall\n")
	return os.WriteFile(filepath.Join(cfg.TargetDir, "Makefile"), []byte(sb.String()), 0644)
}
