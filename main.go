package main

import (
    "encoding/xml"
    "fmt"
    "os"

    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/lipgloss"
)

type Manifest struct {
    XMLName xml.Name `xml:"manifest"`

    Default struct {
        SyncJ    string `xml:"sync-j,attr"`
        Revision string `xml:"revision,attr"`
    } `xml:"default"`

    Remotes  []Remote  `xml:"remote"`
    Projects []Project `xml:"project"`
}

type Remote struct {
    Fetch    string `xml:"fetch,attr"`
    Name     string `xml:"name,attr"`
    Revision string `xml:"revision,attr"`
    Upstream string `xml:"upstream,attr"`
    Review   string `xml:"review,attr"`
}

type Project struct {
    Path     string `xml:"path,attr"`
    Name     string `xml:"name,attr"`
    Remote   string `xml:"remote,attr"`
    Revision string `xml:"revision,attr"`
    Upstream string `xml:"upstream,attr"`
    Groups   string `xml:"groups,attr"`
}

type item struct {
    title       string
    description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type model struct {
    list list.Model
    width  int
    height int
}

var (
    appStyle = lipgloss.NewStyle().Padding(1, 2)

    titleStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("#FFFDF5")).
            Background(lipgloss.Color("#25A065")).
            Padding(0, 1)

    statusMessageStyle = lipgloss.NewStyle().
                Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
                Render
)


// 

type listKeyMap struct {
    toggleSpinner    key.Binding
    toggleTitleBar   key.Binding
    toggleStatusBar  key.Binding
    togglePagination key.Binding
    toggleHelpMenu   key.Binding
    insertItem       key.Binding
}

func newListKeyMap() *listKeyMap {
    return &listKeyMap{
        insertItem: key.NewBinding(
            key.WithKeys("a"),
            key.WithHelp("a", "add item"),
        ),
        toggleSpinner: key.NewBinding(
            key.WithKeys("s"),
            key.WithHelp("s", "toggle spinner"),
        ),
        toggleTitleBar: key.NewBinding(
            key.WithKeys("T"),
            key.WithHelp("T", "toggle title"),
        ),
        toggleStatusBar: key.NewBinding(
            key.WithKeys("S"),
            key.WithHelp("S", "toggle status"),
        ),
        togglePagination: key.NewBinding(
            key.WithKeys("P"),
            key.WithHelp("P", "toggle pagination"),
        ),
        toggleHelpMenu: key.NewBinding(
            key.WithKeys("H"),
            key.WithHelp("H", "toggle help"),
        ),
    }
}


func NewModel(items []list.Item) model {
    const defaultWidth = 0
    const defaultHeight = 0

    l := list.New(items, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
    l.Title = "Parsed Manifest"

    return model{
        list: l,
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.list.SetSize(msg.Width, msg.Height-2) // leave room for padding/title
        return m, nil
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return lipgloss.NewStyle().Padding(1).Render(m.list.View())
}

func main() {
    data, err := os.ReadFile("default.xml")
    if err != nil {
        fmt.Println("Failed to read default.xml:", err)
        os.Exit(1)
    }

    var manifest Manifest
    if err := xml.Unmarshal(data, &manifest); err != nil {
        fmt.Println("Failed to parse XML:", err)
        os.Exit(1)
    }

    // // items {{{
    // // Default
    // var items []list.Item
    // items = append(items, item{
    //     title:       "Default",
    //     description: fmt.Sprintf("sync-j: %s, revision: %s", manifest.Default.SyncJ, manifest.Default.Revision),
    // })
    // d := NewModel(items)
    // if _, err := tea.NewProgram(d, tea.WithAltScreen()).Run(); err != nil {
    //     fmt.Println("Error running TUI:", err)
    //     os.Exit(1)
    // }
    // // }}}

    // Add Remotes
    var itemsRemotes []list.Item
    for _, r := range manifest.Remotes {
        itemsRemotes = append(itemsRemotes, item{
            title:       "Remote: " + r.Name,
            description: fmt.Sprintf("[fetch: %s] [revision: %s]\n[upstream: %s] [review: %s]", r.Fetch, r.Revision, r.Upstream, r.Review),
        })
    }
    r := NewModel(itemsRemotes)
    if _, err := tea.NewProgram(r, tea.WithAltScreen()).Run(); err != nil {
        fmt.Println("Error running TUI:", err)
        os.Exit(1)
    }

    // Add Projects
    var itemsProjects []list.Item
    for _, p := range manifest.Projects {
        itemsProjects = append(itemsProjects, item{
            title:       "Project: " + p.Name,
            description: fmt.Sprintf("[path: %s] [remote: %s] [revision: %s]\n[upstream: %s] [groups: %s]", p.Path, p.Remote, p.Revision, p.Upstream, p.Groups),
        })
    }
    p := NewModel(itemsProjects)
    if _, err := tea.NewProgram(p, tea.WithAltScreen()).Run(); err != nil {
        fmt.Println("Error running TUI:", err)
        os.Exit(1)
    }

}

