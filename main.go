package main

import (
    "encoding/xml"
    "fmt"
    "os"
    "strings"

    // "github.com/charmbracelet/bubbles/cursor"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type Manifest struct {
    XMLName  xml.Name  `xml:"manifest"`
    Default  Default   `xml:"default"`
    Remotes  []Remote  `xml:"remote"`
    Projects []Project `xml:"project"`
}

type Default struct {
    SyncJ    string `xml:"sync-j,attr"`
    Revision string `xml:"revision,attr"`
}

type Remote struct {
    Fetch    string `xml:"fetch,attr"`
    Name     string `xml:"name,attr"`
    Revision string `xml:"revision,attr"`
    Upstream string `xml:"upstream,attr"`
    Review   string `xml:"review,attr"`
}

type Project struct {
    Path       string `xml:"path,attr"`
    Name       string `xml:"name,attr"`
    Remote     string `xml:"remote,attr"`
    Revision   string `xml:"revision,attr"`
    Upstream   string `xml:"upstream,attr"`
    Groups     string `xml:"groups,attr"`
    DestBranch string `xml:"dest-branch,attr"`
}

type EntryType int

const (
    EntryDefault EntryType = iota
    EntryRemote
    EntryProject
)

type entry struct {
    typeOf EntryType
    title  string
    desc   string
    index  int
}

func (e entry) Title() string       { return e.title }
func (e entry) Description() string { return e.desc }
func (e entry) FilterValue() string { return e.title }

type model struct {
    manifest   Manifest
    list       list.Model
    editing    bool
    editingKey string
    inputs     []textinput.Model
    labels     []string
    cursor     int
    editIndex  int
    editType   EntryType
    saving     bool
    errorMsg   string
    labelPad   int
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func newModel(manifest Manifest) model {
    items := []list.Item{
        entry{EntryDefault, "default", fmt.Sprintf("sync-j: %s, revision: %s", manifest.Default.SyncJ, manifest.Default.Revision), 0},
    }
    for i, r := range manifest.Remotes {
        items = append(items, entry{EntryRemote, fmt.Sprintf("remote: %s", r.Name), fmt.Sprintf("fetch: %s, revision: %s", r.Fetch, r.Revision), i})
    }
    for i, p := range manifest.Projects {
        items = append(items, entry{EntryProject, fmt.Sprintf("project: %s", p.Name), fmt.Sprintf("path: %s, remote: %s", p.Path, p.Remote), i})
    }

    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Styles.Title = titleStyle
    l.Title = fmt.Sprintf("Edit %s file", manifestFile)
    return model{manifest: manifest, list: l}
}

func (m model) Init() tea.Cmd {
    return nil
}

// item style
var (
    // args {{{
    manifestFile string = "default.xml"
    // }}}
    // style {{{
    // blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
    // focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

    // titleMainStyle = lipgloss.NewStyle().
    //                  Foreground(lipgloss.Color("#FFFDF5")).
    //                  Background(lipgloss.Color("#25A065")).
    //                  Padding(0, 1)

    titleStyle     = lipgloss.
                     NewStyle().
                     Bold(false).
                     Foreground(lipgloss.Color("15")).
                     Background(lipgloss.Color("4")).
                     Padding(0, 1)

    menuStyle      = lipgloss.
                     NewStyle().
                     Bold(false).
                     Foreground(lipgloss.Color("7"))
                     // Background(lipgloss.Color("0"))

    blurredStyle   = lipgloss.
                     NewStyle().
                     Foreground(lipgloss.Color("8"))
                     // Background(lipgloss.Color("0"))

    focusedStyle   = lipgloss.
                     NewStyle().
                     Underline(false).
                     Bold(false).
                     // Background(lipgloss.Color("0")).
                     Foreground(lipgloss.Color("13"))

    noStyle        = lipgloss.NewStyle().
                     Background(lipgloss.Color("0")).
                     Foreground(lipgloss.Color("15")).
                     Blink(false)

    // cursorStyle         = focusedStyle
    // helpStyle           = blurredStyle
    // cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
    // focusedButton = focusedStyle.Render("[ Submit ]")
    // blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
    // }}}
)


func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.list.SetSize(msg.Width-4, msg.Height-6)
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+q":
            if m.editing {
                tea.ExitAltScreen()
                m.editing = false
                return m, nil
            } else {
                return m, tea.Quit
            }
        case "ctrl+c":
            return m, tea.Quit
        case "ctrl+k":
            if m.editing {
                m.cursor = (m.cursor - 1 + len(m.inputs)) % len(m.inputs)
                for i := range m.inputs {
                    if i == m.cursor {
                        m.inputs[i].Focus()
                        m.inputs[i].PromptStyle = focusedStyle
                        m.inputs[i].TextStyle = focusedStyle
                        m.inputs[i].Cursor.Style = noStyle
                    } else {
                        m.inputs[i].Blur()
                        m.inputs[i].PromptStyle = blurredStyle
                        m.inputs[i].TextStyle = blurredStyle
                        m.inputs[i].Cursor.Style = noStyle
                    }
                }
                return m, nil
			}
        case "ctrl+j":
            if m.editing {
                m.cursor = (m.cursor + 1) % len(m.inputs)
                for i := range m.inputs {
                    if i == m.cursor {
                        m.inputs[i].Focus()
                        m.inputs[i].PromptStyle = focusedStyle
                        m.inputs[i].TextStyle = focusedStyle
                        m.inputs[i].Cursor.Style = noStyle
                    } else {
                        m.inputs[i].Blur()
                        m.inputs[i].PromptStyle = blurredStyle
                        m.inputs[i].TextStyle = blurredStyle
                        m.inputs[i].Cursor.Style = noStyle
                    }
                }
                return m, nil
			}
        case "tab", "down":
            if m.editing {
                m.cursor = (m.cursor + 1) % len(m.inputs)
                for i := range m.inputs {
                    if i == m.cursor {
                        m.inputs[i].Focus()
                        m.inputs[i].PromptStyle = focusedStyle
                        m.inputs[i].TextStyle = focusedStyle
                        m.inputs[i].Cursor.Style = noStyle
                    } else {
                        m.inputs[i].Blur()
                        m.inputs[i].PromptStyle = blurredStyle
                        m.inputs[i].TextStyle = blurredStyle
                        m.inputs[i].Cursor.Style = noStyle
                    }
                }
                return m, nil
            }
        case "shift+tab", "up":
            if m.editing {
                m.cursor = (m.cursor - 1 + len(m.inputs)) % len(m.inputs)
                for i := range m.inputs {
                    if i == m.cursor {
                        m.inputs[i].Focus()
                        m.inputs[i].PromptStyle = focusedStyle
                        m.inputs[i].TextStyle = focusedStyle
                        m.inputs[i].Cursor.Style = noStyle
                    } else {
                        m.inputs[i].Blur()
                        m.inputs[i].PromptStyle = blurredStyle
                        m.inputs[i].TextStyle = blurredStyle
                        m.inputs[i].Cursor.Style = noStyle
                    }
                }
                return m, nil
            }
        case "ctrl+s":
            if m.editing {
                return m.saveEdit()
            }
        case "ctrl+z":
            if m.editing {
                sel := m.list.SelectedItem().(entry)
                return m.initEdit(sel)
            }
        case "enter":
            if m.editing {
                return m, m.Init()
            }
            if m.editing {
                return m.saveEdit()
            }
            sel := m.list.SelectedItem().(entry)
            return m.initEdit(sel)
        }
    }

    if m.editing {
        for i := range m.inputs {
            m.inputs[i], _ = m.inputs[i].Update(msg)
        }
        return m, nil
    }

    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m model) initEdit(sel entry) (tea.Model, tea.Cmd) {
    m.editing = true
    m.editIndex = sel.index
    m.editType = sel.typeOf
    m.inputs = nil
    m.labels = nil
    m.cursor = 0
    m.labelPad = 0

    makeInput := func(val string) textinput.Model {
        ti := textinput.New()
        ti.Prompt = ""
        ti.ShowSuggestions = true
        ti.SetSuggestions([]string{val})
        ti.Placeholder = val
        ti.CharLimit = 640
        ti.SetValue(val)
        ti.CursorEnd()
        return ti
    }

    switch sel.typeOf {
        case EntryDefault:
            m.labels = []string{"sync-j", "revision"}
            m.inputs = []textinput.Model{makeInput(m.manifest.Default.SyncJ), makeInput(m.manifest.Default.Revision)}
        case EntryRemote:
            r := m.manifest.Remotes[sel.index]
            m.labels = []string{"fetch", "name", "revision", "upstream", "review"}
            m.inputs = []textinput.Model{makeInput(r.Fetch), makeInput(r.Name), makeInput(r.Revision), makeInput(r.Upstream), makeInput(r.Review)}
        case EntryProject:
            p := m.manifest.Projects[sel.index]
            m.labels = []string{"path", "name", "remote", "revision", "upstream", "groups", "dest-branch"}
            m.inputs = []textinput.Model{makeInput(p.Path), makeInput(p.Name), makeInput(p.Remote), makeInput(p.Revision), makeInput(p.Upstream), makeInput(p.Groups), makeInput(p.DestBranch)}
    }

    for _, value := range m.labels {
        if m.labelPad < len(value) {
            m.labelPad = len(value)
        }
    }

    for i := range m.inputs {
        m.inputs[i].Blur()
        m.inputs[i].PromptStyle = blurredStyle
        m.inputs[i].TextStyle = blurredStyle
        m.inputs[i].Cursor.Style = blurredStyle
    }

    m.inputs[0].PromptStyle = focusedStyle
    m.inputs[0].TextStyle = focusedStyle
    m.inputs[0].Cursor.Style = noStyle
    m.inputs[0].Focus()
    return m, nil
}

// func (m model) closeAltScreen() (tea.Model, tea.Cmd) {
//     m.editing = false
//     tea.ExitAltScreen()
//     return m, nil
// }

func (m model) saveEdit() (tea.Model, tea.Cmd) {
    vals := make(map[string]string)
    for i, lbl := range m.labels {
        vals[lbl] = m.inputs[i].Value()
    }
    switch m.editType {
        case EntryDefault:
            m.manifest.Default.SyncJ = vals["sync-j"]
            m.manifest.Default.Revision = vals["revision"]
        case EntryRemote:
            m.manifest.Remotes[m.editIndex] = Remote{
                Fetch:    vals["fetch"],
                Name:     vals["name"],
                Revision: vals["revision"],
                Upstream: vals["upstream"],
                Review:   vals["review"],
            }
        case EntryProject:
            m.manifest.Projects[m.editIndex] = Project{
                Path:       vals["path"],
                Name:       vals["name"],
                Remote:     vals["remote"],
                Revision:   vals["revision"],
                Upstream:   vals["upstream"],
                Groups:     vals["groups"],
                DestBranch: vals["dest-branch"],
            }
    }
    m.editing = false
    return m.writeToFile()
}

func (m model) writeToFile() (tea.Model, tea.Cmd) {
    out, err := xml.MarshalIndent(m.manifest, "", "  ")
    if err != nil {
        m.errorMsg = fmt.Sprintf("Failed to marshal XML: %v", err)
        return m, nil
    }
    out = []byte(xml.Header + string(out) + "\n")
    err = os.WriteFile(manifestFile, out, 0644)
    if err != nil {
        m.errorMsg = fmt.Sprintf("Failed to write XML: %v", err)
        return m, nil
    }

    // Reload file
    data, err := os.ReadFile(manifestFile)
    if err != nil {
        m.errorMsg = fmt.Sprintf("Failed to reload %s: %v", manifestFile, err)
        return m, nil
    }
    var newManifest Manifest
    if err := xml.Unmarshal(data, &newManifest); err != nil {
        m.errorMsg = fmt.Sprintf("Failed to parse updated XML: %v", err)
        return m, nil
    }

    // Rebuild list
    items := []list.Item{
        entry{EntryDefault, "default", fmt.Sprintf("sync-j: %s, revision: %s", newManifest.Default.SyncJ, newManifest.Default.Revision), 0},
    }
    for i, r := range newManifest.Remotes {
        items = append(items, entry{EntryRemote, fmt.Sprintf("remote: %s", r.Name), fmt.Sprintf("fetch: %s, revision: %s", r.Fetch, r.Revision), i})
    }
    for i, p := range newManifest.Projects {
        items = append(items, entry{EntryProject, fmt.Sprintf("project: %s", p.Name), fmt.Sprintf("path: %s, remote: %s", p.Path, p.Remote), i})
    }
    m.manifest = newManifest
    m.list.SetItems(items)
    m.errorMsg = "Saved successfully and reloaded."

    return m, nil
}


func (m model) View() string {
    if m.editing {
        var b strings.Builder
        b.WriteString(titleStyle.Render("Editing")+"\n")
        // b.WriteString(menuStyle.Render("-----------")+"\n")
        b.WriteString("\n")
        for i, input := range m.inputs {
            tmpPadWidth := m.labelPad - len(m.labels[i])
            b.WriteString(menuStyle.Render(fmt.Sprintf("> %s %s :", m.labels[i], strings.Repeat(" ", tmpPadWidth)))+" "+input.View()+"\n")
        }
        // b.WriteString(menuStyle.Render("-----------")+"\n")
        b.WriteString("\n")
        b.WriteString(blurredStyle.Render("(C-q to quit, type to edit, C-s to save)\n")) // "(C-q to quit, type to edit)\n"
        return docStyle.Render(b.String())
    }
    return docStyle.Render(m.list.View()) + "\n" + m.errorMsg // + "\n(q to quit, enter to edit)"
}

func main() {
    argc := len(os.Args)
    if argc == 2 {
        manifestFile = os.Args[1]
    }
    data, err := os.ReadFile(manifestFile)
    if err != nil {
        fmt.Printf("Failed to read %s:%s\n", manifestFile, err)
        return
    }
    var manifest Manifest
    if err := xml.Unmarshal(data, &manifest); err != nil {
        fmt.Println("Failed to parse XML:", err)
        return
    }
    p := tea.NewProgram(newModel(manifest), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Println("Error running program:", err)
        os.Exit(1)
    }
}

