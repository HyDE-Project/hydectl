package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func EditConfigFile(appName, fileName string, fileConfig ConfigFile) {
	fmt.Printf("\n🔧 Editing %s - %s\n", appName, fileConfig.Description)
	fmt.Printf("📁 %s\n\n", fileConfig.Path)

	configPath := ExpandPath(fileConfig.Path)

	if len(fileConfig.PreHook) > 0 {
		fmt.Println("⏳ Running pre-hook...")
		if err := runHook(fileConfig.PreHook); err != nil {
			fmt.Printf("⚠️  Pre-hook failed: %v\n", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {

		editors := []string{"nvim", "vim", "nano", "code", "gedit"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}

	if editor == "" {
		fmt.Println("No editor found. Please set the EDITOR environment variable.")
		return
	}

	fmt.Printf("🚀 Opening %s...\n", editor)
	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running editor: %v\n", err)
		return
	}

	if len(fileConfig.PostHook) > 0 {
		fmt.Println("\n⏳ Running post-hook...")
		if err := runHook(fileConfig.PostHook); err != nil {
			fmt.Printf("⚠️  Post-hook failed: %v\n", err)
		} else {
			fmt.Println("✅ Post-hook completed successfully")
		}
	}

	fmt.Printf("\n✅ Configuration editing completed for %s!\n", appName)
}

func runHook(hook []string) error {
	if len(hook) == 0 {
		return nil
	}

	cmd := exec.Command(hook[0], hook[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
