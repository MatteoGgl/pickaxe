package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Selection holds a path and whether it's a directory.
type Selection struct {
	Path  string
	IsDir bool
}

// PickerResult is returned when the user confirms or cancels.
type PickerResult struct {
	Confirmed  bool
	Selections []Selection
}

type fileItem struct {
	path    string
	name    string
	isDir   bool
	checked bool
}

// Model is the bubbletea model for the file picker.
type Model struct {
	root        string
	cwd         string
	items       []fileItem
	cursor      int
	offset      int // first visible item index (for scrolling)
	height      int // terminal height
	filter      textinput.Model
	filterMode  bool
	preSelected map[string]bool
	done        bool
	result      PickerResult
}

const headerLines = 3 // vault path + controls + blank line

var (
	checkedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	dirStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	filterStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

// NewPicker creates a new picker model rooted at root. preSelected is the set of already-registered paths.
func NewPicker(root string, preSelected map[string]bool) *Model {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 60

	m := &Model{
		root:        root,
		cwd:         root,
		filter:      ti,
		preSelected: preSelected,
		height:      24, // sensible default until WindowSizeMsg arrives
	}
	m.loadItems()
	return m
}

func (m *Model) loadItems() {
	entries, err := os.ReadDir(m.cwd)
	if err != nil {
		m.items = nil
		return
	}

	var items []fileItem
	if m.cwd != m.root {
		items = append(items, fileItem{path: filepath.Dir(m.cwd), name: "..", isDir: true})
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		fullPath := filepath.Join(m.cwd, e.Name())
		item := fileItem{
			path:    fullPath,
			name:    e.Name(),
			isDir:   e.IsDir(),
			checked: m.preSelected[fullPath],
		}
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].name == ".." {
			return true
		}
		if items[j].name == ".." {
			return false
		}
		if items[i].isDir != items[j].isDir {
			return items[i].isDir
		}
		return items[i].name < items[j].name
	})

	m.items = items
	m.cursor = 0
	m.offset = 0
}

func (m *Model) visibleItems() []fileItem {
	query := m.filter.Value()
	if query == "" {
		return m.items
	}
	var filtered []fileItem
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.name), strings.ToLower(query)) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// listHeight returns how many file rows fit in the terminal.
func (m *Model) listHeight() int {
	extra := 0
	if m.filterMode {
		extra = 2 // "Filter: ..." + blank line
	}
	h := m.height - headerLines - extra - 1 // -1 for scroll indicator line
	if h < 1 {
		h = 1
	}
	return h
}

// clampOffset adjusts m.offset so cursor stays in view.
func (m *Model) clampOffset() {
	lh := m.listHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+lh {
		m.offset = m.cursor - lh + 1
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.filterMode {
		return m.updateFilter(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.clampOffset()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			m.result = PickerResult{Confirmed: false}
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.clampOffset()
			}

		case "down", "j":
			visible := m.visibleItems()
			if m.cursor < len(visible)-1 {
				m.cursor++
				m.clampOffset()
			}

		case " ":
			visible := m.visibleItems()
			if m.cursor < len(visible) {
				idx := m.itemIndex(visible[m.cursor])
				if idx >= 0 {
					m.items[idx].checked = !m.items[idx].checked
				}
			}

		case "enter":
			visible := m.visibleItems()
			if m.cursor < len(visible) {
				item := visible[m.cursor]
				if item.isDir && item.name != ".." {
					m.cwd = item.path
					m.loadItems()
				} else if item.name == ".." {
					m.cwd = filepath.Dir(m.cwd)
					m.loadItems()
				}
			}

		case "ctrl+d":
			var sels []Selection
			for _, item := range m.items {
				if item.checked {
					sels = append(sels, Selection{Path: item.path, IsDir: item.isDir})
				}
			}
			m.done = true
			m.result = PickerResult{Confirmed: true, Selections: sels}
			return m, tea.Quit

		case "/":
			m.filterMode = true
			m.filter.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m *Model) itemIndex(visible fileItem) int {
	for i, item := range m.items {
		if item.path == visible.path {
			return i
		}
	}
	return -1
}

func (m *Model) updateFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.filterMode = false
			m.filter.Blur()
			m.filter.SetValue("")
			m.cursor = 0
			m.offset = 0
			return m, nil
		case "enter":
			m.filterMode = false
			m.filter.Blur()
			m.cursor = 0
			m.offset = 0
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	visible := m.visibleItems()
	lh := m.listHeight()

	var sb strings.Builder
	sb.WriteString("  Vault: " + m.cwd + "\n")
	sb.WriteString("  [space] toggle  [enter] open dir  [ctrl+d] confirm  [/] filter  [q] cancel\n\n")

	if m.filterMode {
		sb.WriteString("  Filter: " + filterStyle.Render(m.filter.View()) + "\n\n")
	}

	end := m.offset + lh
	if end > len(visible) {
		end = len(visible)
	}

	for i, item := range visible[m.offset:end] {
		absIdx := m.offset + i

		cursor := "  "
		if absIdx == m.cursor {
			cursor = "> "
		}

		check := "[ ]"
		if item.checked {
			check = checkedStyle.Render("[x]")
		}

		name := item.name
		if item.isDir {
			name = dirStyle.Render(name + "/")
		}

		line := cursor + check + " " + name
		if absIdx == m.cursor {
			line = selectedStyle.Render(line)
		}
		sb.WriteString(line + "\n")
	}

	// Scroll indicator
	if len(visible) > lh {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
			fmt.Sprintf("  %d-%d of %d", m.offset+1, end, len(visible)),
		) + "\n")
	}

	return sb.String()
}

// Result returns the picker result. Only valid after the program exits.
func (m *Model) Result() PickerResult {
	return m.result
}

// Run starts the picker and returns the result.
func Run(root string, preSelected map[string]bool) (PickerResult, error) {
	m := NewPicker(root, preSelected)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return PickerResult{}, err
	}
	return finalModel.(*Model).Result(), nil
}
