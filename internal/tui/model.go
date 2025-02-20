package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ma111e/excuses/internal/types"
	"log"
	"math/rand"
	"net/rpc"
	"strings"
	"time"
)

var (
	// Pride flag colors
	rainbowColors = []string{
		"#e60000", // Red
		"#ff8e00", // Orange
		"#ffef00", // Yellow
		"#00821b", // Green
		"#004bff", // Blue
		"#780089", // Purple
	}

	// Styles will be initialized with random color in InitialModel
	borderStyle  lipgloss.Style
	quoteStyle   lipgloss.Style
	helpStyle    lipgloss.Style
	spinnerStyle lipgloss.Style
)

type model struct {
	quote        string
	spinner      spinner.Model
	loading      bool
	err          error
	width        int
	height       int
	color        string
	nextLink     string
	previousLink string
	client       *rpc.Client
}

type fetchMsg struct {
	quote        string
	nextLink     string
	previousLink string
	err          error
}

func getRandomColor() string {
	return rainbowColors[rand.Intn(len(rainbowColors))]
}

func initStyles(color string) {
	borderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color))

	quoteStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Italic(true).
		Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Align(lipgloss.Center)

	spinnerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Align(lipgloss.Center)
}

func InitialModel(serverAddr string) model {
	rand.Seed(time.Now().UnixNano())
	color := getRandomColor()
	initStyles(color)

	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	client, err := rpc.DialHTTP("tcp", serverAddr)
	if err != nil {
		log.Fatal("Connection error:", err)
	}

	return model{
		spinner: s,
		loading: true,
		width:   80,
		height:  10,
		color:   color,
		client:  client,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchQuote(m.client, ""),
	)
}

// Update and View methods remain largely the same, just update fetchQuote calls to include m.client
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "right", "l":
			m.loading = true
			newColor := getRandomColor()
			initStyles(newColor)
			m.color = newColor
			m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(newColor))
			return m, fetchQuote(m.client, m.nextLink)
		case "left", "h":
			m.loading = true
			newColor := getRandomColor()
			initStyles(newColor)
			m.color = newColor
			m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(newColor))
			return m, fetchQuote(m.client, m.previousLink)
		case "a":
			m.loading = true
			newColor := getRandomColor()
			initStyles(newColor)
			m.color = newColor
			m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(newColor))
			return m, fetchQuote(m.client, "/?0")
		case "e":
			m.loading = true
			newColor := getRandomColor()
			initStyles(newColor)
			m.color = newColor
			m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(newColor))
			return m, fetchQuote(m.client, "/?last")
		case "r":
			m.loading = true
			newColor := getRandomColor()
			initStyles(newColor)
			m.color = newColor
			m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(newColor))
			return m, fetchQuote(m.client, "/")
		}

	case fetchMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.quote = strings.TrimSpace(msg.quote)
		m.nextLink = msg.nextLink
		m.previousLink = msg.previousLink
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	contentWidth := m.width - 4

	var content strings.Builder

	if m.loading {
		spinnerText := spinnerStyle.
			Width(contentWidth).
			Render(m.spinner.View() + " Loading...")
		content.WriteString(spinnerText)
	} else if m.err != nil {
		errorText := quoteStyle.
			Width(contentWidth).
			Render("Error: " + m.err.Error())
		content.WriteString(errorText)
	} else {
		quote := quoteStyle.
			Width(contentWidth).
			Render(m.quote)
		content.WriteString(quote + "\n\n")

		helpTop := helpStyle.
			Width(contentWidth).
			Render("← Previous (h) • Next (l) → • Random (r)")
		content.WriteString(helpTop + "\n")

		helpBottom := helpStyle.
			Width(contentWidth).
			Render("First (a) • Last (e) • q to quit")
		content.WriteString(helpBottom)
	}

	return "\n" + borderStyle.
		Width(m.width-2).
		Render(content.String()) + "\n"
}

func fetchQuote(client *rpc.Client, path string) tea.Cmd {
	return func() tea.Msg {
		args := &types.FetchQuoteRequest{Path: path}
		var reply types.FetchQuoteResponse

		err := client.Call("QuoteServer.FetchQuote", args, &reply)
		if err != nil {
			return fetchMsg{err: err}
		}

		if reply.Error != "" {
			return fetchMsg{err: error(nil)}
		}

		return fetchMsg{
			quote:        reply.Quote,
			nextLink:     reply.NextLink,
			previousLink: reply.PreviousLink,
		}
	}
}
