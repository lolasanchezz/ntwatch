package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	p := tea.NewProgram(initialModel(), tea.WithMouseAllMotion(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error : %v", err)
		os.Exit(1)

	}

	//	getCStruct()

}
