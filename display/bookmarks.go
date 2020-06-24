package display

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/makeworld-the-better-one/amfora/bookmarks"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

// For adding and removing bookmarks, basically a clone of the input modal.
var bkmkModal = cview.NewModal().
	SetTextColor(tcell.ColorWhite)

// bkmkCh is for the user action
var bkmkCh = make(chan int) // 1, 0, -1 for add/update, cancel, and remove
var bkmkModalText string    // The current text of the input field in the modal

func bkmkInit() {
	if viper.GetBool("a-general.color") {
		bkmkModal.SetBackgroundColor(tcell.ColorTeal).
			SetButtonBackgroundColor(tcell.ColorNavy).
			SetButtonTextColor(tcell.ColorWhite)
	} else {
		bkmkModal.SetBackgroundColor(tcell.ColorBlack).
			SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack)
		bkmkModal.GetForm().
			SetLabelColor(tcell.ColorWhite).
			SetFieldBackgroundColor(tcell.ColorWhite).
			SetFieldTextColor(tcell.ColorBlack)
	}

	bkmkModal.SetBorder(true)
	bkmkModal.SetBorderColor(tcell.ColorWhite)
	bkmkModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonLabel {
		case "Add":
			bkmkCh <- 1
		case "Change":
			bkmkCh <- 1
		case "Remove":
			bkmkCh <- -1
		case "Cancel":
			bkmkCh <- 0
		}

		//tabPages.SwitchToPage(strconv.Itoa(curTab)) - handled in bkmk()
	})
	bkmkModal.GetFrame().SetTitleColor(tcell.ColorWhite)
	bkmkModal.GetFrame().SetTitleAlign(cview.AlignCenter)
	bkmkModal.GetFrame().SetTitle(" Add Bookmark ")
}

// Bkmk displays the "Add a bookmark" modal.
// It accepts the default value for the bookmark name that will be displayed, but can be changed by the user.
// It also accepts a bool indicating whether this page already has a bookmark.
// It returns the bookmark name and the bookmark action:
// 1, 0, -1 for add/update, cancel, and remove
func openBkmkModal(name string, exists bool) (string, int) {
	// Basically a copy of Input()

	// Remove and re-add input field - to clear the old text
	if bkmkModal.GetForm().GetFormItemCount() > 0 {
		bkmkModal.GetForm().RemoveFormItem(0)
	}
	bkmkModalText = ""
	bkmkModal.GetForm().AddInputField("Name: ", name, 0, nil,
		func(text string) {
			// Store for use later
			bkmkModalText = text
		})

	bkmkModal.ClearButtons()
	if exists {
		bkmkModal.SetText("Change or remove the bookmark for the current page?")
		bkmkModal.AddButtons([]string{"Change", "Remove", "Cancel"})
	} else {
		bkmkModal.SetText("Create a bookmark for the current page?")
		bkmkModal.AddButtons([]string{"Add", "Cancel"})
	}
	tabPages.ShowPage("bkmk")
	tabPages.SendToFront("bkmk")
	App.SetFocus(bkmkModal)
	App.Draw()

	action := <-bkmkCh
	tabPages.SwitchToPage(strconv.Itoa(curTab))

	return bkmkModalText, action
}

// Bookmarks displays the bookmarks page on the current tab.
func Bookmarks() {
	// Gather bookmarks
	rawContent := "# Bookmarks\r\n\r\n"
	m, keys := bookmarks.All()
	for i := range keys {
		rawContent += fmt.Sprintf("=> %s %s\r\n", keys[i], m[keys[i]])
	}
	// Render and display
	content, links := renderer.RenderGemini(rawContent, textWidth())
	page := structs.Page{Content: content, Links: links, Url: "about:bookmarks"}
	setPage(&page)
}

// addBookmark goes through the process of adding a bookmark for the current page.
// It is the high-level way of doing it. It should be called in a goroutine.
// It can also be called to edit an existing bookmark.
func addBookmark() {
	if !strings.HasPrefix(tabMap[curTab].Url, "gemini://") {
		// Can't make bookmarks for other kinds of URLs
		return
	}

	name, exists := bookmarks.Get(tabMap[curTab].Url)
	// Open a bookmark modal with the current name of the bookmark, if it exists
	newName, action := openBkmkModal(name, exists)
	switch action {
	case 1:
		// Add/change the bookmark
		bookmarks.Set(tabMap[curTab].Url, newName)
	case -1:
		bookmarks.Remove(tabMap[curTab].Url)
	}
	// Other case is action = 0, meaning "Cancel", so nothing needs to happen
}
