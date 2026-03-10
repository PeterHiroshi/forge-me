package dashboard

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpToggleOn(t *testing.T) {
	m := Model{width: 80, height: 24}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	updated := newModel.(Model)
	if !updated.showHelp {
		t.Error("pressing ? should toggle showHelp on")
	}
}

func TestHelpToggleOff(t *testing.T) {
	m := Model{width: 80, height: 24, showHelp: true}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	updated := newModel.(Model)
	if updated.showHelp {
		t.Error("pressing ? again should toggle showHelp off")
	}
}

func TestHelpDismissWithEsc(t *testing.T) {
	m := Model{width: 80, height: 24, showHelp: true}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := newModel.(Model)
	if updated.showHelp {
		t.Error("pressing Esc should dismiss help")
	}
}

func TestHelpSuppressesOtherKeys(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		showHelp:  true,
		activeTab: TabOverview,
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	updated := newModel.(Model)
	if updated.activeTab != TabOverview {
		t.Errorf("tab should not change while help is shown, got %d", updated.activeTab)
	}
}

func TestHelpSuppressesQuit(t *testing.T) {
	m := Model{width: 80, height: 24, showHelp: true}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Error("q should not quit while help is shown")
	}
}

func TestRenderHelpContainsShortcuts(t *testing.T) {
	m := Model{width: 80, height: 24}
	content := m.renderHelp()
	shortcuts := []string{"Tab", "j/k", "q", "r", "?", "/", "Enter", "Esc"}
	for _, s := range shortcuts {
		if !strings.Contains(content, s) {
			t.Errorf("help should contain shortcut %q", s)
		}
	}
}
