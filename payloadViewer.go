package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type payloadViewer struct {
	currentData map[socketKey]table.Row
	table       table.Model
	currentRows []table.Row
}

func (m *model) payloadViewerUpdate(msg tea.Msg) (model, tea.Cmd) {
	if len(m.payloadViewer.currentRows) == 0 {
		m.updateRows(&m.sockets)
	}
	writeToDebug(strconv.Itoa((len(m.payloadViewer.currentRows))))
	m.payloadViewer.table.SetRows(m.payloadViewer.currentRows)
	switch msg.(type) {
	case matchedPkt:
		//update table - will do later!
	case SocketTableRefreshedMsg:
		m.updateRows(&m.sockets)
	}

	//m.payloadViewer.viewport.SetContent(m.payloadViewer.currentData)
	/*
		switch msg.(type) {
		case packetMsg:
			m.payloadViewer.currentData = msg.(packetMsg).NetworkLayer().NetworkFlow().Dst().String()
		default:
			m.payloadViewer.currentData = "no packets yet"
		}
		//m.currentData = "e"
	*/
	return *m, nil
}
func (m *model) updateRows(socketTable *socketMap) {
	tempArr := make(map[socketKey]table.Row)
	rows := []table.Row{}
	for key, socket := range *socketTable {
		connType := ""
		if key.ConnType == 0 {
			connType = "UDP"
		} else if key.ConnType == 1 {
			connType = "TCP"
		}
		row := table.Row{
			key.ProcessName,
			key.DestIP,
			strconv.Itoa(int(key.DestPort)),
			connType,
			strconv.Itoa(int(socket.Pid)),
		}
		rows = append(rows, row)
		tempArr[key] = row
	}
	m.payloadViewer.currentData = tempArr
	m.payloadViewer.currentRows = rows

	// Safely append a textual representation of the 5th row (if present) to debug.txt.
	// Convert table.Row to []string before joining.
	if len(rows) > 4 {
		f, err := os.OpenFile("./debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			_, _ = f.WriteString(strings.Join([]string(rows[4]), ",") + "\n")
			_ = f.Close()
		}
	}
}

func (m *model) payloadViewerInit() tea.Cmd {
	m.payloadViewer.table = table.New()
	m.payloadViewer.table.SetColumns([]table.Column{
		{Title: "Process Name", Width: 18},
		{Title: "Dest IP", Width: 18},
		{Title: "Dest Port", Width: 10},
		{Title: "Conn", Width: 6},
		{Title: "PID", Width: 8},
	})

	m.payloadViewer.currentData = make(map[socketKey]table.Row)
	m.payloadViewer.table.Focus()
	return nil
}

func (m *payloadViewer) View() string {

	return m.table.View()
}

func writeToDebug(str string) {
	f, err := os.OpenFile("./debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		_, _ = f.WriteString(str + "\n")
		_ = f.Close()
	}
}
