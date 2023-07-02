package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/process"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
}

type Process struct {
	name string
	pid  int32
	cpu  float64
	ram  float32
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func main() {
	columns := []table.Column{
		{Title: "pid", Width: 15},
		{Title: "name", Width: 25},
		{Title: "cpu usage", Width: 15},
		{Title: "ram usage", Width: 15},
	}

	allProcesses, err := process.Processes()
	if err != nil {
		fmt.Println("Error while reading all the processes:", err)
		return
	}

	currentPID := int32(os.Getpid())

	processes := make([]*Process, 0)
	for _, p := range allProcesses {
		cpuPercent, _ := p.CPUPercent()
		processName, _ := p.Name()
		ramPercent, _ := p.MemoryPercent()
		if cpuPercent > 0 && p.Pid != currentPID {
			processes = append(processes, &Process{pid: p.Pid, cpu: cpuPercent, name: processName, ram: ramPercent})
		}
	}

	sort.SliceStable(processes, func(i, j int) bool {
		return processes[i].cpu > processes[j].cpu
	})

	limit := 5
	if len(processes) < limit {
		limit = len(processes)
	}

	rows := make([]table.Row, limit)
	for i := 0; i < limit; i++ {
		p := processes[i]
		rows[i] = table.Row{
			strconv.Itoa(int(p.pid)),
			p.name,
			strconv.FormatFloat(p.cpu, 'f', 2, 64),
			strconv.FormatFloat(float64(p.ram), 'f', 2, 32),
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
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
	t.SetStyles(s)

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
