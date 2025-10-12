package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/justinbather/harness/internal/harness"
	"github.com/justinbather/harness/internal/logger"
	"go.uber.org/zap"
)

type screen int

const (
	topicsScreen screen = iota
	messagesScreen
)

type model struct {
	harness *harness.Harness

	screen screen

	topics   table.Model
	messages table.Model

	selectedTopic string
}

func initialModel(harness *harness.Harness) model {
	topicsTable := newTopicsTable(harness.ListTopics())

	return model{
		harness: harness,
		topics:  topicsTable,
		screen:  topicsScreen,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// key press
	case tea.KeyMsg: // only process globally applicable events here, like quit etc, else defer to the current view
		switch msg.String() {
		case "enter": // defer this to the current view too

		case "esc", "ctrl+o": // back
			if m.screen == messagesScreen {

			}

		case "ctrl+c", "q":
			return m, tea.Quit
		}

	}

	switch m.screen {
	case topicsScreen:
		var cmd tea.Cmd
		m.topics, cmd = m.topics.Update(msg)
		return m, cmd
		// need messages screen here
	}

	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case topicsScreen:
		header := "Harness"
		subHeader := "Kafka Topics"
		footer := "q to quit\nj/k for up/down\n"

		return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", header, subHeader, m.topics.View(), footer)
	}

	return "error"
}

func newTopicsTable(topicMap map[string]int) table.Model {
	columns := []table.Column{{Title: "Topic", Width: 30}, {Title: "# Messages", Width: 30}}
	var rows []table.Row

	for topic, msgs := range topicMap {
		conv := strconv.Itoa(msgs)
		rows = append(rows, table.Row{topic, conv})
	}

	t := table.New(table.WithColumns(columns), table.WithRows(rows), table.WithFocused(true))

	t.SetHeight(10)

	return t
}

func main() {

	ctx := context.Background()
	log, ctx := logger.FromCtx(ctx)

	harness, err := harness.New("localhost:9092")
	if err != nil {
		panic(err)
	}

	harness.Start(ctx)
	defer harness.Shutdown(ctx)
	// janky way to wa
	time.Sleep(100 * time.Millisecond)

	p := tea.NewProgram(initialModel(harness), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Error("running TUI", zap.Error(err))
	}

}
