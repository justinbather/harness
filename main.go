package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/justinbather/harness/internal/harness"
	"github.com/justinbather/harness/internal/logger"
	"github.com/justinbather/harness/internal/store"
	"go.uber.org/zap"
)

type refreshMsg struct{}

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

	alert        string
	alertExpires time.Time
}

var copyAlert = "message copied to your clipboard!"

func initialModel(harness *harness.Harness) model {
	topicsTable := newTopicsTable(harness.ListTopics())

	return model{
		harness:       harness,
		topicsModel:   topicsTable,
		currentScreen: topicsScreen,
	}
}

func (m *model) setAlert(msg string) {
	m.alert = msg
	m.alertExpires = time.Now().Add(5 * time.Second)
}

func (m model) Init() tea.Cmd {
	// return m.poll() // bit tricky with map reordering and cursor position, will need more work before adding again
	return nil
}

// avoids wipinig old state on refresh like cursor position etc
func (m *model) refreshTopicsList() {
	topics := m.harness.ListTopics()
	rows := make([]table.Row, len(topics))
	for _, t := range topics {
		rows = append(rows, table.Row{t.Name, strconv.Itoa(t.Partitions), strconv.Itoa(t.MessageCount)})
	}

	m.topicsModel.SetRows(rows)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case refreshMsg:
		m.refreshTopicsList()
		return m, m.poll() // continue polling

	// key press
	case tea.KeyMsg: // only process globally applicable events here, like quit etc, else defer to the current view
		switch msg.String() {
		case "enter":
			if m.currentScreen == topicsScreen {
				m.selectedTopic = m.topicsModel.SelectedRow()[0]
				m.messagesModel = newMessageTable(m.harness.ListMessages(m.selectedTopic))
				m.currentScreen = messagesScreen
				m.messagesModel.Focus()
			}

		case "y":
			if m.currentScreen == messagesScreen {
				currRow := m.messagesModel.SelectedRow()

				message, err := m.harness.GetMessage(currRow[OFFSET], m.selectedTopic)
				if err != nil {
					panic(err)
				}

				err = clipboard.WriteAll(string(message.Data))
				if err != nil {
					panic(err)
				}

				m.setAlert(copyAlert)
			}

		case "esc", "ctrl+o": // back
			if m.currentScreen == messagesScreen {
				m.currentScreen = topicsScreen
				m.topicsModel = newTopicsTable(m.harness.ListTopics())
				m.topicsModel.Focus()
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
	alert := ""
	if time.Now().Before(m.alertExpires) {
		alert = m.alert
	}

	brokers := ""
	for _, b := range m.harness.Brokers {
		brokers += b + " "
	}

	switch m.currentScreen {
	case topicsScreen:
		header := "Harness " + brokers
		subHeader := "Kafka Topics"
		footer := "q to quit\nj/k for up/down\n"

		return fmt.Sprintf("%s\n\n%s\n%s\n%s\n\n%s", header, subHeader, alert, m.topicsModel.View(), footer)

	case messagesScreen:
		header := "Harness " + brokers
		subHeader := m.selectedTopic
		footer := "q to quit\nj/k for up/down\nctrl+o or esc to go back\ny to copy message\n"

		return fmt.Sprintf("%s\n\n%s\n%s\n%s\n\n%s", header, subHeader, alert, m.messagesModel.View(), footer)
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

	t.SetHeight(30)

	return t
}

type MessageRow int

const (
	MSG_ROW_IDX MessageRow = iota
	PARTITION
	OFFSET
	DATA_LENGTH
)

// cache this table and check if any new messages have been consumed into this topic
func newMessageTable(messages []store.Message) table.Model {
	cols := []table.Column{{Title: "#", Width: 5}, {Title: "Partition", Width: 25}, {Title: "Offset", Width: 10}, {Title: "Data Length", Width: 30}}

	var rows []table.Row

	for i, msg := range messages {
		rows = append(rows, table.Row{strconv.Itoa(i), msg.Partition, msg.Offset, strconv.Itoa(len(msg.Data))})
	}

	t := table.New(table.WithColumns(cols), table.WithRows(rows), table.WithFocused(true))

	t.SetHeight(50)

	return t
}

func (m model) poll() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return refreshMsg{}
	})
}

func bufferData(data string) string {
	if len(data) < 26 {
		return data
	}

	return data[:27] + "..."
}

func main() {
	defaultKafkaHost := "localhost:9092"
	cliArg := ""

	if len(os.Args) < 2 {
		cliArg = defaultKafkaHost
	} else {
		cliArg = os.Args[1]
	}

	ctx := context.Background()
	log, ctx := logger.FromCtx(ctx)

	harness, err := harness.New(cliArg)
	if err != nil {
		panic(err)
	}

	harness.Start(ctx)
	defer harness.Shutdown(ctx)
	// not sure why if we dont sleep we get 0 message count in topic list
	time.Sleep(20 * time.Millisecond)

	p := tea.NewProgram(initialModel(harness), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Error("running TUI", zap.Error(err))
	}

}
