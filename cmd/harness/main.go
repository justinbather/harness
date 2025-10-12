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
	"github.com/justinbather/harness/internal/store"
	"go.uber.org/zap"
)

type screen int

const (
	topicsScreen screen = iota
	messagesScreen
)

type model struct {
	harness *harness.Harness

	currentScreen screen

	topicsModel   table.Model
	messagesModel table.Model

	selectedTopic string
}

func initialModel(harness *harness.Harness) model {
	topicsTable := newTopicsTable(harness.ListTopics())

	return model{
		harness:       harness,
		topicsModel:   topicsTable,
		currentScreen: topicsScreen,
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
			if m.currentScreen == topicsScreen {
				m.messagesModel = newMessageTable(m.selectedTopic, m.harness.ListMessages(m.selectedTopic))
				m.selectedTopic = m.topicsModel.SelectedRow()[0]
				m.currentScreen = messagesScreen
				m.messagesModel.Focus()
			}

		case "esc", "ctrl+o": // back
			if m.currentScreen == messagesScreen {
				m.currentScreen = topicsScreen
				return m, nil
			}

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	switch m.currentScreen {
	case topicsScreen:
		var cmd tea.Cmd
		m.topicsModel, cmd = m.topicsModel.Update(msg)
		return m, cmd

	case messagesScreen:
		var cmd tea.Cmd
		m.messagesModel, cmd = m.messagesModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	switch m.currentScreen {
	case topicsScreen:
		header := "Harness"
		subHeader := "Kafka Topics"
		footer := "q to quit\nj/k for up/down\n"

		return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", header, subHeader, m.topicsModel.View(), footer)

	case messagesScreen:
		header := "Harness"
		subHeader := m.selectedTopic
		footer := "q to quit\nj/k for up/down\nctrl+o or esc to go back\n"

		return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", header, subHeader, m.messagesModel.View(), footer)
	}

	return "error"
}

func newTopicsTable(topicMap map[string]*store.Topic) table.Model {
	columns := []table.Column{{Title: "Topic", Width: 30}, {Title: "Partitions", Width: 30}, {Title: "# Messages", Width: 30}}
	var rows []table.Row

	for _, topic := range topicMap {
		convertedMsgCount := strconv.Itoa(topic.MessageCount)
		convertedPartitionCount := strconv.Itoa(topic.Partitions)
		rows = append(rows, table.Row{topic.Name, convertedPartitionCount, convertedMsgCount})
	}

	t := table.New(table.WithColumns(columns), table.WithRows(rows), table.WithFocused(true))

	t.SetHeight(10)

	return t
}

func newMessageTable(topic string, messages []store.Message) table.Model {
	cols := []table.Column{{Title: "#", Width: 5}, {Title: "Partition", Width: 15}, {Title: "Offset", Width: 10}, {Title: "Data", Width: 30}}

	var rows []table.Row

	for i, msg := range messages {
		rows = append(rows, table.Row{strconv.Itoa(i), strconv.Itoa(msg.Partition), strconv.Itoa(msg.Offset), msg.Data})
	}

	t := table.New(table.WithColumns(cols), table.WithRows(rows), table.WithFocused(true))

	t.SetHeight(50)

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
