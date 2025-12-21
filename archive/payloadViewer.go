package main

/*
import (
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type payloadViewer struct {
	currentData map[socketKey]table.Row
	table       table.Model
	currentRows []table.Row
}

func (m *model) payloadViewerUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	m.payloadViewer.table, cmd = m.payloadViewer.table.Update(msg)

	// Ensure columns exist before resize
	if len(m.payloadViewer.table.Columns()) == 0 {
		m.payloadViewerInit()
	}
	m.resizeTable()

	changed := false

	// Build rows if empty once
	if len(m.payloadViewer.currentRows) == 0 {
		m.updateRows(&m.sockets)
		changed = true // ensure SetRows runs at least once
	}

	switch ev := msg.(type) {
	case matchedPkt:
		// if this should update rows, set changed = true and call updateRows
	case SocketTableRefreshedMsg:
		if ev.socks != nil {
			m.sockets = *ev.socks
		}
		m.updateRows(&m.sockets)
		changed = true
	}

	// Apply rows AFTER any updates
	if changed {
		m.payloadViewer.table.SetRows(m.payloadViewer.currentRows)
	}

	return *m, cmd
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
			strconv.Itoa(int(socket.SourcePort)),
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

}

func (m *model) payloadViewerInit() tea.Cmd {
	m.payloadViewer.table = table.New(
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.payloadViewer.table.SetStyles(s)

	m.payloadViewer.table.SetColumns([]table.Column{
		{Title: "Process Name", Width: 18},
		{Title: "Dest IP", Width: 18},
		{Title: "Source Port", Width: 18},
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

func (m *model) resizeTable() {
	padding := 3
	colWidth := (m.width - padding) / len(m.payloadViewer.table.Columns())

	tmp := []table.Column{}
	for _, col := range m.payloadViewer.table.Columns() {
		col.Width = colWidth
		tmp = append(tmp, col)
	}
	//	m.payloadViewer.table.SetWidth(m.width)
	m.payloadViewer.table.SetColumns(tmp)
	m.payloadViewer.table.SetHeight(m.height)
}
*/
