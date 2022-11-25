package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	URL "net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
	twitterscraper "github.com/n0madic/twitter-scraper"
)

type Opts struct {
	Username     string
	Tweet_id     string
	Batch        string
	Output       string
	Media        string
	Nbr          int
	Dtype        string
	Size         int
	Retweet      bool
	Retweet_only bool
	Proxy        string
}

var (
	Log       *ui.MultilineEntry
	LogSingle *ui.MultilineEntry
	mu        = &sync.Mutex{}
	LogUser   *ui.MultilineEntry
	rt        bool
	GUI       bool
	Stop      = make(chan bool)
	quality   = map[int]string{
		0: "orig",
		1: "normal",
		2: "small",
	}
	client = &http.Client{
		Timeout: time.Second * 20,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: time.Duration(5) * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: time.Duration(5) * time.Second,
		},
	}
)

var download_id = make(chan string)

func LogErr(err string) {
	mu.Lock()
	ui.QueueMain(func() {
		Log.Append(err + "\n")
	})
	mu.Unlock()
}

func Name(s string) string {
	segments := strings.Split(s, "/")
	name := segments[len(segments)-1]

	re := regexp.MustCompile(`name=`)
	if re.MatchString(name) {
		segments := strings.Split(name, "?")
		name = segments[len(segments)-2]
	}
	return name
}

func UserTDownload(opt Opts) {
	var wg sync.WaitGroup
	if opt.Proxy != "" {
		proxyURL, _ := URL.Parse(opt.Proxy)
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}
	os.MkdirAll(opt.Output+"/"+opt.Username, os.ModePerm)
	scraper := twitterscraper.New()
	// do nothing if proxy = ""
	scraper.SetProxy(opt.Proxy)
	for tweet := range scraper.GetTweets(context.Background(), opt.Username, opt.Nbr) {
		select {
		case <-Stop:
			return
		default:
		}

		if tweet.Error != nil {
			fmt.Println(tweet.Error)
			return
		}
		if opt.Media == "videos" || opt.Media == "all" {
			go videoUser(&wg, tweet, opt)
		}
		if opt.Media == "pictures" || opt.Media == "all" {
			go photoUser(&wg, tweet, opt)
		}
	}
	wg.Wait()
	fmt.Printf("Download %d tweet from %s", opt.Nbr, opt.Username)
	time.Sleep(1 * time.Second)
}

func videoUser(wg *sync.WaitGroup, tweet *twitterscraper.TweetResult, opt Opts) {
	if len(tweet.Videos) > 0 {
		if tweet.IsRetweet && (opt.Retweet || opt.Retweet_only) {
			opt.Tweet_id = tweet.ID
			opt.Output = opt.Output + "/" + opt.Username
			SingleTDownload(wg, opt, true, false)
		}
		for _, i := range tweet.Videos {
			j := fmt.Sprintf("%s", i)
			v := vidUrl(j)
			wg.Add(1)
			go download(wg, v, opt.Output, opt.Username, GUI)
		}
	}
}

func photoUser(wg *sync.WaitGroup, tweet *twitterscraper.TweetResult, opt Opts) {
	if len(tweet.Photos) > 0 {
		if tweet.IsRetweet && (opt.Retweet || opt.Retweet_only) {
			opt.Tweet_id = tweet.ID
			opt.Output = opt.Output + "/" + opt.Username
			SingleTDownload(wg, opt, true, false)
		}
		for _, i := range tweet.Photos {
			select {
			case <-Stop:
				return
			default:
			}

			if opt.Retweet_only || tweet.IsRetweet {
				continue
			}
			if !strings.Contains(i, "video_thumb/") {
				if quality[opt.Size] == "orig" || quality[opt.Size] == "small" {
					i = i + "?name=" + quality[opt.Size]
				}
				wg.Add(1)
				go download(wg, i, opt.Output, opt.Username, GUI)
			}
		}
	}
}

func BatchTDownload(opt Opts) {
	var wg sync.WaitGroup
	scanner := bufio.NewScanner(strings.NewReader(opt.Batch))
	if opt.Proxy != "" {
		proxyURL, _ := URL.Parse(opt.Proxy)
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}

	for scanner.Scan() {
		opt.Tweet_id = scanner.Text()
		go SingleTDownload(&wg, opt, false, true)
	}
	// Ugly wait for waiting first wg.Add in singleTDownlaod
	time.Sleep(10 * time.Second)
	wg.Wait()
}

func SingleTDownload(wg *sync.WaitGroup, opt Opts, rt bool, batch bool) {
	if !rt || !batch {
		if opt.Proxy != "" {
			proxyURL, _ := URL.Parse(opt.Proxy)
			client = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
			}
		}
	}
	scraper := twitterscraper.New()
	scraper.SetProxy(opt.Proxy)
	tweet, err := scraper.GetTweet(opt.Tweet_id)
	if err != nil {
		LogErr("Scrap: " + err.Error())
		return
	}

	var gwg sync.WaitGroup
	gwg.Add(1)
	go func() {
		defer gwg.Done()
		if len(tweet.Videos) > 0 {
			for _, i := range tweet.Videos {
				select {
				case <-Stop:
					return
				default:
				}

				j := fmt.Sprintf("%s", i)
				v := vidUrl(j)
				if GUI {
					wg.Add(1)
					go download(wg, v, opt.Output, "", true)
					if !rt && !batch {
						n := Name(v)
						mu.Lock()
						ui.QueueMain(func() {
							LogSingle.Append("Downloaded vid: " + n + "\n")
						})
						mu.Unlock()
					}
				} else {
					wg.Add(1)
					go download(wg, v, opt.Output, "", false)
				}
			}
		}
	}()
	gwg.Add(1)
	go func() {
		defer gwg.Done()
		if len(tweet.Photos) > 0 {
			for _, i := range tweet.Photos {
				select {
				case <-Stop:
					return
				default:
				}

				if !strings.Contains(i, "video_thumb/") {
					if quality[opt.Size] == "orig" || quality[opt.Size] == "small" {
						i = i + "?name=" + quality[opt.Size]
					}
					if GUI {
						wg.Add(1)
						go download(wg, i, opt.Output, "", true)
						if !rt && !batch {
							n := Name(i)
							mu.Lock()
							ui.QueueMain(func() {
								LogSingle.Append("Downloaded img: " + n + "\n")
							})
							mu.Unlock()
						}
					} else {
						wg.Add(1)
						go download(wg, i, opt.Output, "", false)
					}
				}
			}
		}
	}()
	gwg.Wait()
	if GUI && !rt && !batch {
		fmt.Print("wait")
		wg.Wait()
		fmt.Print("Done")
		mu.Lock()
		ui.QueueMain(func() {
			LogSingle.Append("--------------------------\n")
		})
		mu.Unlock()
	}
}

func vidUrl(video string) string {
	vid := strings.Split(string(video), " ")
	v := vid[len(vid)-1]
	v = strings.TrimSuffix(v, "}")
	vid = strings.Split(v, "?")
	return vid[0]
}

func download(wg *sync.WaitGroup, url string, output string, user string, gui bool) {
	defer wg.Done()
	select {
	case <-Stop:
		wg.Done()
		return
	default:
	}

	name := Name(url)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		if gui {
			LogErr("http: " + err.Error())
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")
	var resp *http.Response

	finish := make(chan bool)
	var http_err error
	hwg := sync.WaitGroup{}
	hwg.Add(2)
	go func() {
		defer hwg.Done()
		resp, http_err = client.Do(req)
		if http_err != nil {
			return
		}
		finish <- true
	}()
	go func() {
		defer hwg.Done()
		for {
			select {
			case <-Stop:
				cancel()
				return
			case <-finish:
				return
			default:
			}
		}
	}()
	hwg.Wait()

	if http_err != nil {
		if gui {
			LogErr("http_err: " + http_err.Error())
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}
	var f *os.File
	defer f.Close()
	var ferr error
	if user != "" {
		f, ferr = os.Create(output + "/" + user + "/" + name)
	} else {
		f, ferr = os.Create(output + "/" + name)
	}
	if ferr != nil {
		if gui {
			LogErr("file: " + ferr.Error())
		} else {
			fmt.Println(err.Error())
		}
		return
	}
	var cerr error
	cwg := sync.WaitGroup{}
	cwg.Add(2)
	go func() {
		defer cwg.Done()
		_, cerr = io.Copy(f, resp.Body)
		if cerr != nil {
			finish <- true
			return
		}
		finish <- true
	}()
	go func() {
		defer cwg.Done()
		for {
			select {
			case <-Stop:
				cancel()
				return
			case <-finish:
				return
			default:
			}
		}
	}()
	cwg.Wait()

	if cerr != nil {
		if gui {
			LogErr("Copy: " + cerr.Error())
			return
		} else {
			fmt.Println("Copy: " + cerr.Error())
			return
		}
	}
	fmt.Println("Download: ", name)
}
