package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	everquest "github.com/Mortimus/goEverquest"
)

// Zevfeer's gear
// [Fri May 14 14:50:52 2021] Could not find player 006DF3000000000000000000000000000000000000000000000000000000000000000000000000000007216E420Zevfeer's
// 5 aa achiev
// [Fri May 14 14:51:14 2021] Could not find player 3Mortimus^100005^0^1^1612448610^0^'5
// Frostreaver's spell
// [Fri May 14 14:51:31 2021] Could not find player 63^1350^'Frostreaver's

// Page2Button10Name=auc2
// Page2Button10Color=0
// Page2Button10Line1=/gu 3Mortimus^11000071^1^0^22^'PORTS at tunnel entrance. Parcel all donations to Mortimus.

// Enter and exit links with 

// C:\Program Files (x86)\Steam\steamapps\common\Everquest F2P\Mortimus_aradune.ini

var link binding.String
var couldntfind = "Could not find player "

const secretSauce = ""

func main() {
	// 	logPath := `C:/Program Files (x86)/Steam/steamapps/common/Everquest F2P/Logs/eqlog_Mortimus_aradune.txt`
	var basePath string
	// hotKeyPath := "Mortimus_aradune.ini"
	// hotkeys := readHotkeys(hotKeyPath)
	link = binding.NewString()
	link.Set("63^1350^'Frostreaver's")
	// updateHotkey("Mortimus_aradune.ini", "Page2Button9Line1", "/g Charming %t")
	myApp := app.New()
	myWindow := myApp.NewWindow("Everquest Hotkeys")

	help := widget.NewTextGridFromString("How to\nStart everquest, and login to your character.\nHit the locate everquest folder button and locate the base everquest folder (contains eqgame.exe).\nCommon locations\nC:\\Program Files(x86)\\Steam\\Steamapps\\common\\Everquest F2P\nC:\\Users\\Public\\Daybreak Game Company\\Installed Games\\Everquest")
	openBtn := widget.NewButton("Locate Everquest Folder", func() {
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if list == nil {
				log.Println("Cancelled")
				return
			}
			basePath = list.Path()
			characters := getCharacters(list.Path())
			// dialog.ShowInformation("Folder Open", list.Path(), myWindow)
			charSelect := myApp.NewWindow("Select the character to edit hotkeys for")
			charList := widget.NewList(
				func() int {
					return len(characters)
				},
				func() fyne.CanvasObject {
					wid := widget.NewButton("click me", func() {
						log.Println("tapped")
					})
					return wid
				},
				func(i widget.ListItemID, o fyne.CanvasObject) {
					o.(*widget.Button).SetText(characters[i])
					o.(*widget.Button).OnTapped = func() {
						fmt.Printf("Selected %s\n", characters[i])
						ChatLogs := make(chan everquest.EqLog)
						go everquest.BufferedLogRead(basePath+"/Logs/eqlog_"+characters[i]+".txt", false, 6, ChatLogs)
						go parseLogs(ChatLogs)
						hotKeyPath := basePath + "/" + characters[i] + ".ini"
						hotkeys := readHotkeys(hotKeyPath)
						list := widget.NewList(
							func() int {
								return len(hotkeys)
							},
							func() fyne.CanvasObject {
								wid := widget.NewButton("click me", func() {
									log.Println("tapped")
								})
								wid.Alignment = widget.ButtonAlign(fyne.TextAlignLeading)
								return wid
							},
							func(i widget.ListItemID, o fyne.CanvasObject) {
								o.(*widget.Button).SetText(lookuphotkey(i, hotkeys))
								o.(*widget.Button).OnTapped = func() {
									win := myApp.NewWindow("Edit Hotkey")
									help := widget.NewTextGridFromString("Send a CROSS-SERVER tell to the link you want to use\n;t ITEM_LINK\n;t ACHIEVMENT_LINK\n;t SPELL_LINK\nThis will update the box below shortly after Everquest says it cannot find the player.")
									help2 := widget.NewTextGridFromString("After inserting the link, make sure to scroll right and change the EDIT_THIS_TEXT to what you want it to say.\nAfter saving the hotkey CAMP to DESKTOP\n/camp desktop\nAnd log back in, it will NOT reload unless you do this.")
									input := widget.NewEntry()
									input.SetText(hotkeys[i].value) // Max size is 255 characters
									linkInput := widget.NewEntryWithData(link)
									linkInput.Disable()
									linkInput.Resize(fyne.NewSize(400, 50))
									content := container.NewVBox(
										widget.NewLabel(hotkeys[i].key),
										help,
										linkInput,
										widget.NewButton("Insert Link", func() {
											input.Text += secretSauce + linkInput.Text + "EDIT_THIS_TEXT" + secretSauce
											input.Refresh()
										}),
										input,
										widget.NewButton("Save", func() {
											log.Printf("Saving %s\n", hotkeys[i].key)
											hotkeys[i].value = input.Text
											updateHotkey(hotKeyPath, hotkeys[i].key, hotkeys[i].value)
										}),
										help2,
									)
									win.SetContent(content)
									win.Resize(fyne.NewSize(500, 200))
									win.Show()
								}
							})
						charHotkeys := myApp.NewWindow("Select social to edit")
						charHotkeys.Resize(fyne.NewSize(500, 600))
						charHotkeys.SetContent(list)
						charHotkeys.Show()
						charSelect.Hide()
						charHotkeys.SetOnClosed(func() {
							os.Exit(0)
						})
					}
				})
			charSelect.Resize(fyne.NewSize(300, 600))
			charSelect.SetContent(charList)
			charSelect.Show()
			myWindow.Hide()

		}, myWindow)
	})

	splashContainer := container.NewVBox(help, openBtn)

	myWindow.SetContent(splashContainer)
	myWindow.Resize(fyne.NewSize(500, 400))
	myWindow.ShowAndRun()
}

func lookuphotkey(order int, hotkeys []HotKey) string {
	for i, val := range hotkeys {
		if i == order {
			return val.key + "=" + val.value
		}
	}
	return "" // this should NEVER happen
}

func parseLogs(logs chan everquest.EqLog) {
	for l := range logs { // handle all logs in the channel
		if l.Channel == "system" && strings.Contains(l.Msg, couldntfind) { // Failed messages show as system, and must contain could not find player
			processLog(l)
		}
	}
}

func processLog(l everquest.EqLog) {
	charname := "Mortimus"
	prefixLength := len(couldntfind)
	link.Set(l.Msg[prefixLength:])
	if strings.Contains(l.Msg, charname) { // They are linking an achievement
		// we need to return so we don't process this for spells
		// remove string after final ^
		pos := strings.LastIndex(l.Msg, "^")
		link.Set(l.Msg[prefixLength : pos+2])
		return
	}
	if strings.Contains(l.Msg, "^") { // They are linking a spell
		// remove string after final ^
		pos := strings.LastIndex(l.Msg, "^")
		link.Set(l.Msg[prefixLength : pos+2])
		return
	}
	// They are linking an item
	// 007A070000000000157120000000000000000000000000000000000000000000000000000011BD00000EBFA1C4B
	// 91 chars long
	link.Set(l.Msg[prefixLength : 91+prefixLength])
}

type HotKey struct {
	key   string
	value string
}

func readHotkeys(path string) []HotKey {
	var hotkeys []HotKey
	hotKeyOpen := "[Socials]"
	var inHotbuttons bool
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, hotKeyOpen) {
			inHotbuttons = true
			continue
		} else if text[0] == '[' { // we found a section we don't care about
			inHotbuttons = false
			continue
		}
		if inHotbuttons && strings.Contains(text, "Line") {
			split := strings.Split(text, "=")
			hotkey := HotKey{
				key:   split[0],
				value: split[1],
			}
			hotkeys = append(hotkeys, hotkey)
			// hotbuttons[split[0]] = split[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return hotkeys
}

func updateHotkey(path, key, value string) error {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, key) {
			lines[i] = key + "=" + value
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(path, []byte(output), 0644)
	if err != nil {
		return err
	}
	return nil
}

func getCharacters(basePath string) []string {
	fmt.Printf("Looking for characters in %s\n", basePath+"/Logs")
	var files []string

	err := filepath.Walk(basePath+"/Logs", func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	var results []string
	for _, file := range files {
		file = filepath.Base(file)
		if strings.HasPrefix(file, "eqlog_") {
			// fmt.Printf("%s\n", file[6:len(file)-4])
			results = append(results, file[6:len(file)-4])
		}
	}
	return results
}
