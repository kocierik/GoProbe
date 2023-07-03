package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/process"
	"os"
	"os/exec"
	"sort"
	"strconv"
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
			killAllProcesses(m.table.SelectedRow()[1])
			return m, tea.Batch(
				tea.Printf("process %s killed!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func killAllProcesses(processName string) error {
	cmd := exec.Command("pkill", "-f", processName)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	columns := []table.Column{
		{Title: "pid", Width: 15},
		{Title: "name", Width: 30},
		{Title: "cpu usage", Width: 15},
		{Title: "ram usage", Width: 15},
	}

	allProcesses, err := process.Processes()
	if err != nil {
		fmt.Println("Error while reading all the processes:", err)
		return
	}

	currentPID := int32(os.Getpid())

	processes := make(map[int32]*Process)
	for _, p := range allProcesses {
		cpuPercent, _ := p.CPUPercent()
		processName, _ := p.Name()
		ramPercent, _ := p.MemoryPercent()
		if cpuPercent > 0 && p.Pid != currentPID {
			processes[p.Pid] = &Process{pid: p.Pid, cpu: cpuPercent, name: processName, ram: ramPercent}
		}
	}

	processSlice := make([]*Process, 0, len(processes))
	for _, p := range processes {
		processSlice = append(processSlice, p)
	}

	sort.SliceStable(processSlice, func(i, j int) bool {
		return processSlice[i].cpu > processSlice[j].cpu
	})

	limit := len(processes)
	if len(processSlice) < limit {
		limit = len(processSlice)
	}

	rows := make([]table.Row, limit)
	for i := 0; i < limit; i++ {
		p := processSlice[i]
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
		BorderForeground(lipgloss.Color("257")).
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
