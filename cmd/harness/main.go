package main

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/justinbather/harness/internal/harness"
	"github.com/justinbather/harness/internal/logger"
	"go.uber.org/zap"
)

type model struct {
	topics    map[string]int
	topicList []string
	cursor    int
	selected  map[int]struct{}
}

func initialModel(harness *harness.Harness) model {
	var topicList []string
	topicMp := harness.ListTopics()

	for topic, _ := range topicMp {
		topicList = append(topicList, topic)
	}

	return model{
		topics:    topicMp,
		topicList: topicList,
		cursor:    0,
		selected:  make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// key press
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.topicList)-1 {
				m.cursor++
			}
		}

	}

	return m, nil
}

func (m model) View() string {
	s := "Kafka Topics\n\n"
	for i, topic := range m.topicList {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s %d\n", cursor, topic, m.topics[topic])
	}

	s += "\nPress q to quit.\n"

	return s
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

	p := tea.NewProgram(initialModel(harness))
	if _, err := p.Run(); err != nil {
		log.Error("running TUI", zap.Error(err))
	}

}
