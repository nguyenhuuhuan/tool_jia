package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// --- Main Application ---

func getStatusColor(status string) string {
	switch strings.ToLower(status) {
	case "to do":
		return "[red]"
	case "in progress":
		return "[blue]"
	case "done":
		return "[green]"
	case "selected for development":
		return "[purple]"
	case "in testing":
		return "[yellow]"
	case "ready for test":
		return "[yellow]"

	default:
		return "[white]"
	}
}

func getIssueTypeColor(issueType string) string {
	switch strings.ToLower(issueType) {
	case "bug":
		return "[red]"
	case "story":
		return "[green]"
	case "task":
		return "[blue]"
	case "epic":
		return "[orange]"
	default:
		return "[white]"
	}
}

func getPriorityColor(priority string) string {
	switch strings.ToLower(priority) {
	case "highest":
		return "[red]"
	case "high":
		return "[orange]"
	case "medium":
		return "[yellow]"
	case "low":
		return "[blue]"
	case "lowest":
		return "[gray]"
	default:
		return "[white]"
	}
}

func getTcellColorFromStatus(status string) tcell.Color {
	switch strings.ToLower(status) {
	case "to do":
		return tcell.ColorRed
	case "in progress":
		return tcell.ColorBlue
	case "done":
		return tcell.ColorGreen
	case "selected for development":
		return tcell.ColorPurple
	default:
		return tcell.ColorWhite
	}
}

func updateStatus(app *tview.Application, statusTextView *tview.TextView, message string, isError bool) {
	color := "green"
	if isError {
		color = "red"
	}
	statusTextView.SetText(fmt.Sprintf("[%s]%s", color, message))
	app.Draw()
}

func createIssueList() *tview.List {
	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle("Your Jira Tickets (Press Enter for options)")
	list.SetBorderColor(tcell.ColorGray)
	list.SetTitleColor(tcell.ColorWhite)
	list.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	list.SetSelectedTextColor(tcell.ColorWhite)
	list.SetMainTextColor(tcell.ColorWhite)
	list.SetSecondaryTextColor(tcell.ColorLightGray)
	return list
}

func createSearchField() *tview.InputField {
	searchField := tview.NewInputField()
	searchField.SetLabel("Search: ")
	searchField.SetFieldWidth(0)
	searchField.SetBorder(true)
	searchField.SetBorderColor(tcell.ColorDarkCyan)
	searchField.SetLabelColor(tcell.ColorAqua)
	searchField.SetFieldBackgroundColor(tcell.ColorDefault)
	searchField.SetFieldTextColor(tcell.ColorWhite)
	return searchField
}

func createStatusTextView() *tview.TextView {
	status := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	status.SetBorder(true)
	status.SetBorderColor(tcell.ColorDarkCyan)
	return status
}

func createDetailPane() *tview.TextView {
	detailPane := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetWordWrap(true).
		SetScrollable(true)
	detailPane.SetBorder(true).SetTitle("Ticket Details")
	detailPane.SetBorderColor(tcell.ColorDarkCyan)
	detailPane.SetTitleColor(tcell.ColorAqua)
	detailPane.SetText("Select a ticket to view details.")
	return detailPane
}

func setupActionModal(app *tview.Application, mainFlex *tview.Flex, list *tview.List, displayedIssues *[]Issue, updateStatusFunc func(message string, isError bool)) *tview.Modal {
	modal := tview.NewModal().
		SetText("What do you want to do?").
		AddButtons([]string{"Open in Browser", "Generate Branch Name", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(mainFlex, true).SetFocus(list)
			if buttonLabel == "Cancel" {
				return
			}

			selectedIssueIndex := list.GetCurrentItem()
			if selectedIssueIndex < 0 || selectedIssueIndex >= len(*displayedIssues) {
				updateStatusFunc("Invalid issue selection.", true)
				return
			}
			issue := (*displayedIssues)[selectedIssueIndex]

			switch buttonLabel {
			case "Open in Browser":
				url := fmt.Sprintf("%s/browse/%s", jiraBaseURL, issue.Key)
				if err := OpenBrowser(app, url); err != nil {
					go updateStatusFunc(fmt.Sprintf("Error opening browser: %v", err), true)
				} else {
					go updateStatusFunc(fmt.Sprintf("Opening %s...", issue.Key), false)
				}
			case "Generate Branch Name":
				branchName := GenerateBranchName(issue)
				if err := clipboard.WriteAll(branchName); err != nil {
					go updateStatusFunc(fmt.Sprintf("Error copying to clipboard: %v", err), true)
				} else {
					go updateStatusFunc(fmt.Sprintf("Copied to clipboard: %s", branchName), false)
				}
			}
		})
	return modal
}

func setupListChangedFunc(list *tview.List, detailPane *tview.TextView, searchField *tview.InputField, statusTextView *tview.TextView, displayedIssues *[]Issue) {
	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index < 0 || index >= len(*displayedIssues) {
			detailPane.SetText("Select a ticket to view details.")
			return
		}
		issue := (*displayedIssues)[index]
		statusColor := getTcellColorFromStatus(issue.Fields.Status.Name)
		detailPane.SetTitleColor(statusColor)
		detailPane.SetBorderColor(statusColor)
		list.SetTitleColor(statusColor)
		list.SetBorderColor(statusColor)
		searchField.SetBorderColor(statusColor)
		searchField.SetLabelColor(statusColor)
		statusTextView.SetBorderColor(statusColor)
		formattedDetails := fmt.Sprintf(`[white]Key: [yellow]%s
[white]Summary: [yellow]%s
[white]Status: %s%s[-]
[white]Issue Type: %s%s[-]
[white]Assignee: [yellow]%s
[white]Created: [yellow]%s
[white]Updated: [yellow]%s

[white]Description:
[gray]%s
[white]Comments:
[gray]%s`,
			issue.Key,
			issue.Fields.Summary,
			getStatusColor(issue.Fields.Status.Name), issue.Fields.Status.Name,
			getIssueTypeColor(issue.Fields.IssueType.Name), issue.Fields.IssueType.Name,
			func() string {
				if issue.Fields.Assignee != nil {
					return issue.Fields.Assignee.DisplayName
				}
				return "Unassigned"
			}(),
			issue.Fields.Created.Format("2006-01-02 15:04"),
			issue.Fields.Updated.Format("2006-01-02 15:04"),
			func() string {
				if issue.Fields.Description != nil {
					return fmt.Sprintf("%v", issue.Fields.Description)
				}
				return "No description."
			}(),
			func() string {
				if issue.Fields.Comments != nil && issue.Fields.Comments.Total > 0 {
					var builder strings.Builder
					for _, comment := range issue.Fields.Comments.Comments {
						builder.WriteString(fmt.Sprintf("  - %s (%s): %s\n", comment.Author.DisplayName, comment.Created.Format("2006-01-02"), comment.Body))
					}
					return builder.String()
				}
				return "No comments."
			}(),
		)
		detailPane.SetText(formattedDetails)
	})
}

func setupInputCapture(app *tview.Application, searchField *tview.InputField, list *tview.List, modal *tview.Modal) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyDown {
			if app.GetFocus() == searchField {
				app.SetFocus(list)
				return nil
			}
		}
		if event.Key() == tcell.KeyBacktab || event.Key() == tcell.KeyUp {
			if app.GetFocus() == list && list.GetCurrentItem() == 0 {
				app.SetFocus(searchField)
				return nil
			}
		}
		if event.Rune() == '/' {
			app.SetFocus(searchField)
			return nil
		}
		if event.Key() == tcell.KeyEnter && app.GetFocus() == list {
			app.SetRoot(modal, false).SetFocus(modal)
			return nil
		}
		return event
	})
}

func main() {
	apiToken := os.Getenv("JIRA_API_TOKEN")
	if apiToken == "" {
		fmt.Println("Error: JIRA_API_TOKEN environment variable not set.")
		os.Exit(1)
	}
	email := os.Getenv("JIRA_EMAIL")
	if email == "" {
		fmt.Println("Error: JIRA_EMAIL environment variable not set.")
		os.Exit(1)
	}

	app := tview.NewApplication()

	mainFlex := setupMainApp(app, email, apiToken, "assignee = currentUser() ORDER BY created DESC")
	app.SetRoot(mainFlex, true).SetFocus(mainFlex)

	if err := app.Run(); err != nil {
		fmt.Println(err)
	}
}

func setupMainApp(app *tview.Application, email, apiToken string, initialJQL string) *tview.Flex {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlue
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorDarkBlue
	tview.Styles.BorderColor = tcell.ColorGray
	tview.Styles.TitleColor = tcell.ColorWhite
	tview.Styles.GraphicsColor = tcell.ColorWhite
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorLightGray
	tview.Styles.TertiaryTextColor = tcell.ColorDarkGray
	tview.Styles.InverseTextColor = tcell.ColorBlue
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorLightBlue

	list := createIssueList()
	searchField := createSearchField()
	statusTextView := createStatusTextView()
	detailPane := createDetailPane()

	var allIssues []Issue
	var displayedIssues []Issue
	var currentFilters map[string]string

	updateStatusFunc := func(message string, isError bool) {
		updateStatus(app, statusTextView, message, isError)
	}

	updateListFunc := func(searchTerm string) {
		list.Clear()
		displayedIssues = nil
		searchTerm = strings.ToLower(searchTerm)

		for _, issue := range allIssues {
			matchesSearch := searchTerm == "" ||
				strings.Contains(strings.ToLower(issue.Key), searchTerm) ||
				strings.Contains(strings.ToLower(issue.Fields.Summary), searchTerm)

			matchesFilters := true
			if len(currentFilters) > 0 {
				for key, value := range currentFilters {
					if value == "" {
						continue
					}
					switch key {
					case "status":
						if issue.Fields.Status.Name != value {
							matchesFilters = false
						}
					case "issueType":
						if issue.Fields.IssueType.Name != value {
							matchesFilters = false
						}
					case "assignee":
						assigneeName := "Unassigned"
						if issue.Fields.Assignee != nil {
							assigneeName = issue.Fields.Assignee.DisplayName
						}
						if assigneeName != value {
							matchesFilters = false
						}
					}
					if !matchesFilters {
						break
					}
				}
			}

			if matchesSearch && matchesFilters {
				displayedIssues = append(displayedIssues, issue)
			}
		}

		if len(displayedIssues) == 0 {
			detailPane.SetText("No tickets match your criteria.")
		}

		for _, issue := range displayedIssues {
			statusColor := getStatusColor(issue.Fields.Status.Name)
			mainText := fmt.Sprintf("%s%s: %s", statusColor, issue.Key, issue.Fields.Summary)
			list.AddItem(mainText, "", 0, nil)
		}
		if len(displayedIssues) > 0 {
			list.SetCurrentItem(0)
		}
	}

	searchField.SetChangedFunc(func(text string) {
		updateListFunc(text)
	})

	mainFlex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(searchField, 3, 1, true).
			AddItem(list, 0, 1, false).
			AddItem(statusTextView, 3, 0, false), 0, 1, true).
		AddItem(detailPane, 0, 1, false)

	modal := setupActionModal(app, mainFlex, list, &displayedIssues, updateStatusFunc)
	setupListChangedFunc(list, detailPane, searchField, statusTextView, &displayedIssues)
	list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		app.SetRoot(modal, false).SetFocus(modal)
	})
	setupInputCapture(app, searchField, list, modal)

	app.SetRoot(mainFlex, true).SetFocus(searchField)

	// Fetch issues using the provided JQL
	go func() {
		updateStatusFunc("Fetching Jira tickets...", false)
		issues, err := FetchJiraIssues(email, apiToken, initialJQL)
		if err != nil {
			app.QueueUpdateDraw(func() {
				updateStatusFunc(fmt.Sprintf("Error fetching tickets: %v", err), true)
			})
			return
		}

		app.QueueUpdateDraw(func() {
			allIssues = issues
			if len(allIssues) == 0 {
				updateStatusFunc("No tickets found for the provided JQL.", false)
				return
			}
			statusTextView.Clear()
			updateListFunc("")
			list.SetCurrentItem(0)
			selectedIssue := displayedIssues[0]
			formattedDetails := fmt.Sprintf(`[white]Key: [yellow]%s
[white]Summary: [yellow]%s
[white]Status: %s%s[-]
[white]Issue Type: %s%s[-]
[white]Assignee: [yellow]%s
[white]Created: [yellow]%s
[white]Updated: [yellow]%s

[white]Description:
[gray]%s`,
				selectedIssue.Key,
				selectedIssue.Fields.Summary,
				getStatusColor(selectedIssue.Fields.Status.Name), selectedIssue.Fields.Status.Name,
				getIssueTypeColor(selectedIssue.Fields.IssueType.Name), selectedIssue.Fields.IssueType.Name,
				func() string {
					if selectedIssue.Fields.Assignee != nil {
						return selectedIssue.Fields.Assignee.DisplayName
					}
					return "Unassigned"
				}(),
				selectedIssue.Fields.Created.Format("2006-01-02 15:04"),
				selectedIssue.Fields.Updated.Format("2006-01-02 15:04"),
				func() string {
					if selectedIssue.Fields.Description != nil {
						return fmt.Sprintf("%v", selectedIssue.Fields.Description)
					}
					return "No description."
				}(),
			)
			detailPane.SetText(formattedDetails)
		})
	}()
	return mainFlex
}
