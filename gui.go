package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	. "twmd/utils"

	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
	dg "github.com/sqweek/dialog"
)

const version = "beta"

var (
	windows    *ui.Window
	win        *ui.Window
	quit       = make(chan bool)
	ubox       *ui.Box
	pb         *ui.ProgressBar
	media_type = map[int]string{
		0: "all",
		1: "videos",
		2: "pictures",
	}
)

func LaunchDownload(box *ui.Box, button *ui.Button, stop *ui.Button, opt Opts) {
	ui.QueueMain(func() {
		stop.Enable()
		button.Disable()
		ip := ui.NewProgressBar()
		ip.SetValue(-1)
		box.Append(ip, false)
	})
	switch opt.Dtype {
	case "single":
		var wg sync.WaitGroup
		SingleTDownload(&wg, opt, false, false)
		ui.QueueMain(func() {
			box.Delete(3)
			button.Enable()
			stop.Disable()
		})

	case "user":
		UserTDownload(opt)
		time.Sleep(1 * time.Second)
		ui.QueueMain(func() {
			box.Delete(3)
			button.Enable()
			stop.Disable()
		})

	case "batch":
		BatchTDownload(opt)
		time.Sleep(1 * time.Second)
		ui.QueueMain(func() {
			box.Delete(3)
			button.Enable()
			stop.Disable()
		})
		fmt.Println("Finish batch")

	}
	fmt.Println("Finish")
}

///////////////
//
//  SINGLE TWEET
//
///////////////

func SingleTweet() ui.Control {
	box := ui.NewVerticalBox()
	box.SetPadded(true)
	group := ui.NewGroup(" ")
	group.SetMargined(true)
	box.Append(group, true)
	Form := ui.NewForm()
	Form.SetPadded(true)
	group.SetChild(Form)

	tweet_id := ui.NewEntry()
	Form.Append("Tweet ID: ", tweet_id, false)

	size := ui.NewCombobox()
	size.Append("Large")
	size.Append("normal")
	size.Append("small")
	size.SetSelected(0)
	Form.Append("Picture size: ", size, false)

	folder, _ := os.Getwd()
	str := "Choose (default: " + folder + ")"
	Output := ui.NewButton(str)
	Form.Append("Output Folder", Output, false)
	Output.OnClicked(func(button *ui.Button) {
		folder, err := dg.Directory().Title("Output Folder").Browse()
		if err != nil {
			dg.Message(err.Error()).Title("exception !").Info()
			return
		}
		if len(folder) > 75 {
			path := "...." + folder[len(folder)-75:]
			Output.SetText(path)
		} else {
			Output.SetText(folder)
		}
	})

	proxy := ui.NewEntry()
	Form.Append("Proxy: ", proxy, false)

	LogSingle = ui.NewNonWrappingMultilineEntry()
	LogSingle.SetReadOnly(true)
	Form.Append("Downloads: ", LogSingle, true)

	download := ui.NewButton("Download")
	box.Append(download, false)

	grid := ui.NewGrid()
	grid.SetPadded(true)
	exit := ui.NewButton("Exit")
	exit.OnClicked(func(button *ui.Button) {
		os.Exit(0)
	})

	stop := ui.NewButton("Stop")
	stop.Disable()
	stop.OnClicked(func(button *ui.Button) {
		Stop <- true
	})

	grid.Append(exit,
		0, 0, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(stop,
		1, 0, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)
	box.Append(grid, false)

	download.OnClicked(func(button *ui.Button) {
		if tweet_id.Text() == "" {
			ui.MsgBoxError(windows,
				"Empty Tweet ID.",
				"Fill the tweet id field before click on download.")
			return
		}

		var opt Opts
		opt.Tweet_id = tweet_id.Text()
		opt.Size = size.Selected()
		opt.Proxy = proxy.Text()
		opt.Dtype = "single"
		opt.Output = folder

		Stop = make(chan bool)
		go LaunchDownload(box, download, stop, opt)
	})

	return box

}

//////////////
//
//  USER DOWNLOAD
//
////////////

func UserDownload() ui.Control {
	box := ui.NewVerticalBox()
	box.SetPadded(true)
	group := ui.NewGroup(" ")
	group.SetMargined(true)
	box.Append(group, true)
	Form := ui.NewForm()
	Form.SetPadded(true)
	group.SetChild(Form)

	Username := ui.NewEntry()
	Username.OnChanged(func(entry *ui.Entry) {
		fmt.Scanln(Username.Text())
	})
	Form.Append("Username: ", Username, false)

	media := ui.NewCombobox()
	media.Append("pictures & videos")
	media.Append("videos only")
	media.Append("pictures only")
	media.SetSelected(0)
	Form.Append("Media to download: ", media, false)

	hbox := ui.NewHorizontalBox()

	retweet := ui.NewCheckbox("")
	retweet.SetText("Download retweet:")
	retweet_only := ui.NewCheckbox("")
	retweet_only.SetText("Download retweet only:")
	hbox.Append(retweet, false)
	hbox.Append(retweet_only, false)
	Form.Append("", hbox, false)

	size := ui.NewCombobox()
	size.Append("Large")
	size.Append("normal")
	size.Append("small")
	size.SetSelected(0)
	Form.Append("Picture size: ", size, false)

	nbr_tweet := ui.NewSpinbox(1, 3200)
	nbr_tweet.SetValue(1000)
	Form.Append("Max tweets:", nbr_tweet, false)

	folder, _ := os.Getwd()
	str := "Choose (default: " + folder + ")"
	Output := ui.NewButton(str)
	Form.Append("Output Folder", Output, false)
	Output.OnClicked(func(button *ui.Button) {
		folder, err := dg.Directory().Title("Output Folder").Browse()
		if err != nil {
			dg.Message(err.Error()).Title("exception !").Info()
			return
		}
		if len(folder) > 75 {
			path := "...." + folder[len(folder)-75:]
			Output.SetText(path)
		} else {
			Output.SetText(folder)
		}
	})

	proxy := ui.NewEntry()
	Form.Append("Proxy: ", proxy, false)

	download := ui.NewButton("Download")
	box.Append(download, false)

	grid := ui.NewGrid()
	grid.SetPadded(true)
	exit := ui.NewButton("Exit")
	exit.OnClicked(func(button *ui.Button) {
		os.Exit(0)
	})

	stop := ui.NewButton("Stop")
	stop.OnClicked(func(button *ui.Button) {
		Stop <- true
		close(Stop)
		stop.Disable()
	})
	stop.Disable()
	grid.Append(exit,
		0, 0, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(stop,
		1, 0, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)
	box.Append(grid, false)

	download.OnClicked(func(button *ui.Button) {
		var opt Opts
		opt.Media = media_type[media.Selected()]
		opt.Username = Username.Text()
		opt.Size = size.Selected()
		opt.Proxy = proxy.Text()
		opt.Output = folder
		opt.Nbr = nbr_tweet.Value()
		opt.Retweet = retweet.Checked()
		opt.Retweet_only = retweet_only.Checked()
		opt.Dtype = "user"

		if Username.Text() == "" {
			ui.MsgBoxError(windows,
				"Empty Username.",
				"Fill the username field before click on download.")
			return
		}

		Stop = make(chan bool)
		go LaunchDownload(box, download, stop, opt)
	})

	return box
}

/////////////////
//
// BATCH DOWNLOAD
//
/////////////////

func BatchTweet() ui.Control {
	box := ui.NewVerticalBox()
	box.SetPadded(true)
	group := ui.NewGroup(" ")
	group.SetMargined(true)
	box.Append(group, true)

	Form := ui.NewForm()
	Form.SetPadded(true)
	group.SetChild(Form)

	batch := ui.NewNonWrappingMultilineEntry()
	Form.Append("Tweets IDs: ", batch, true)

	size := ui.NewCombobox()
	size.Append("Large")
	size.Append("normal")
	size.Append("small")
	size.SetSelected(0)
	Form.Append("Picture size: ", size, false)

	folder, _ := os.Getwd()
	str := "Choose (default: " + folder + ")"
	Output := ui.NewButton(str)
	Form.Append("Output Folder", Output, false)
	Output.OnClicked(func(button *ui.Button) {
		folder, err := dg.Directory().Title("Output Folder").Browse()
		if err != nil {
			dg.Message(err.Error()).Title("exception !").Info()
			return
		}
		if len(folder) > 75 {
			path := "...." + folder[len(folder)-75:]
			Output.SetText(path)
		} else {
			Output.SetText(folder)
		}
	})

	proxy := ui.NewEntry()
	Form.Append("Proxy: ", proxy, false)

	download := ui.NewButton("Download")
	box.Append(download, false)
	grid := ui.NewGrid()
	grid.SetPadded(true)

	exit := ui.NewButton("Exit")
	exit.OnClicked(func(button *ui.Button) {
		os.Exit(0)
	})

	stop := ui.NewButton("Stop")
	stop.OnClicked(func(button *ui.Button) {
		Stop <- true
		close(Stop)
		stop.Disable()
	})
	stop.Disable()

	grid.Append(exit,
		0, 0, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(stop,
		1, 0, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)
	box.Append(grid, false)

	download.OnClicked(func(button *ui.Button) {
		var opt Opts
		opt.Size = size.Selected()
		opt.Proxy = proxy.Text()
		opt.Output = folder
		opt.Dtype = "batch"
		opt.Batch = batch.Text()

		if batch.Text() == "" {
			ui.MsgBoxError(windows,
				"Empty tweets.",
				"Fill the tweet ids field before click on download.")
			return
		}

		Stop = make(chan bool)
		go LaunchDownload(box, download, stop, opt)
	})

	return box

}

func AboutForm() ui.Control {
	box := ui.NewVerticalBox()
	box.SetPadded(true)

	Label0 := ui.NewLabel("")
	box.Append(Label0, false)

	Label1 := ui.NewLabel("Twmd: Apiless twitter media downloader")
	box.Append(Label1, false)

	Label3 := ui.NewLabel(fmt.Sprintf("Version: %s", version))
	box.Append(Label3, false)

	l1 := ui.NewLabel(fmt.Sprintf("Repo url: https://github.com/mmpx12/twitter-media-downloader"))
	box.Append(l1, false)

	return box
}

func LogPage() ui.Control {
	box := ui.NewVerticalBox()
	box.SetPadded(true)
	Log = ui.NewMultilineEntry()
	Log.SetReadOnly(true)
	box.Append(Log, true)
	return box
}

func twmd() {
	windows = ui.NewWindow("Twmd: apiless twitter media downloader", 700, 600, true)

	windows.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})

	ui.OnShouldQuit(func() bool {
		windows.Destroy()
		return true
	})

	tabs := ui.NewTab()
	windows.SetChild(tabs)

	windows.SetMargined(true)

	tabs.Append("Single tweet", SingleTweet())
	tabs.SetMargined(0, true)

	tabs.Append("User Download", UserDownload())
	tabs.SetMargined(0, true)

	tabs.Append("Batch tweet", BatchTweet())
	tabs.SetMargined(0, true)

	tabs.Append("Errors Log", LogPage())
	tabs.SetMargined(0, true)

	tabs.Append("About", AboutForm())
	tabs.SetMargined(0, true)

	windows.Show()

}

func main() {
	GUI = true
	os.Setenv("GTK_THEME", "Adwaita:dark")
	ui.Main(twmd)
}
