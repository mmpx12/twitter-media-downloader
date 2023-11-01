package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	URL "net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mmpx12/optionparser"
	twitterscraper "github.com/n0madic/twitter-scraper"
	"golang.org/x/term"
)

var (
	usr     string
	proxy   string
	update  bool
	onlyrtw bool
	vidz    bool
	imgs    bool
	urlOnly bool
	version = "1.10.3"
	scraper *twitterscraper.Scraper
	client  *http.Client
	size    = "orig"
)

func download(wg *sync.WaitGroup, url string, filetype string, output string, dwn_type string) {
	defer wg.Done()
	segments := strings.Split(url, "/")
	name := segments[len(segments)-1]
	re := regexp.MustCompile(`name=`)
	if re.MatchString(name) {
		segments := strings.Split(name, "?")
		name = segments[len(segments)-2]
	}
	if urlOnly {
		fmt.Println(url)
		time.Sleep(2 * time.Millisecond)
		return
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")
	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		fmt.Println("error")
		return
	}

	if resp.StatusCode != 200 {
		fmt.Println("error")
		return
	}

	var f *os.File
	defer f.Close()
	if dwn_type == "user" {
		if update {
			if _, err := os.Stat(output + "/" + filetype + "/" + name); !errors.Is(err, os.ErrNotExist) {
				fmt.Println(name + ": already exists")
				return
			}
		}
		if filetype == "rtimg" {
			f, _ = os.Create(output + "/img/RE-" + name)
		} else if filetype == "rtvideo" {
			f, _ = os.Create(output + "/video/RE-" + name)
		} else {
			f, _ = os.Create(output + "/" + filetype + "/" + name)
		}
	} else {
		if update {
			if _, err := os.Stat(output + "/" + name); !errors.Is(err, os.ErrNotExist) {
				fmt.Println("File exist")
				return
			}
		}
		f, _ = os.Create(output + "/" + name)
	}
	io.Copy(f, resp.Body)
	fmt.Println("Downloaded " + name)
}

func vidUrl(video string) string {
	vid := strings.Split(string(video), " ")
	v := vid[len(vid)-1]
	v = strings.TrimSuffix(v, "}")
	vid = strings.Split(v, "?")
	return vid[0]
}

func videoUser(wait *sync.WaitGroup, tweet *twitterscraper.TweetResult, output string, rt bool) {
	defer wait.Done()
	wg := sync.WaitGroup{}
	if len(tweet.Videos) > 0 {
		for _, i := range tweet.Videos {
			j := fmt.Sprintf("%s", i)
			if tweet.IsRetweet {
				if rt || onlyrtw {
					v := vidUrl(j)
					wg.Add(1)
					go download(&wg, v, "video", output, "user")
				} else {
					continue
				}
			} else if onlyrtw {
				continue
			}
			v := vidUrl(j)
			wg.Add(1)
			go download(&wg, v, "video", output, "user")
		}
		wg.Wait()
	}
}

func photoUser(wait *sync.WaitGroup, tweet *twitterscraper.TweetResult, output string, rt bool) {
	defer wait.Done()
	wg := sync.WaitGroup{}
	if len(tweet.Photos) > 0 || tweet.IsRetweet {
		if tweet.IsRetweet && (rt || onlyrtw) {
			singleTweet(output, tweet.ID)
		}
		for _, i := range tweet.Photos {
			if onlyrtw || tweet.IsRetweet {
				continue
			}
			var url string
			if !strings.Contains(i.URL, "video_thumb/") {
				if size == "orig" || size == "small" {
					url = i.URL + "?name=" + size
				} else {
					url = i.URL
				}
				wg.Add(1)
				go download(&wg, url, "img", output, "user")
			}
		}
		wg.Wait()
	}
}

func videoSingle(tweet *twitterscraper.Tweet, output string) {
	if len(tweet.Videos) > 0 {
		wg := sync.WaitGroup{}
		for _, i := range tweet.Videos {
			j := fmt.Sprintf("%s", i)
			v := vidUrl(j)
			if usr != "" {
				wg.Add(1)
				go download(&wg, v, "rtvideo", output, "user")
			} else {
				wg.Add(1)
				go download(&wg, v, "tweet", output, "tweet")
			}
		}
		wg.Wait()
	}
}

func photoSingle(tweet *twitterscraper.Tweet, output string) {
	if len(tweet.Photos) > 0 {
		wg := sync.WaitGroup{}
		for _, i := range tweet.Photos {
			fmt.Println(i.URL)
			var url string
			if !strings.Contains(i.URL, "video_thumb/") {
				if size == "orig" || size == "small" {
					url = i.URL + "?name=" + size
				} else {
					url = i.URL
				}
				if usr != "" {
					wg.Add(1)
					go download(&wg, url, "rtimg", output, "user")
				} else {
					wg.Add(1)
					go download(&wg, url, "tweet", output, "tweet")
				}
			}
		}
		wg.Wait()
	}
}

func askPass() {
	for {
		var username string
		fmt.Printf("username: ")
		fmt.Scanln(&username)
		fmt.Printf("password: ")
		pass, _ := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		scraper.Login(username, string(pass))
		if !scraper.IsLoggedIn() {
			var code string
			fmt.Printf("two-factor: ")
			fmt.Scanln(&code)
			fmt.Println()
			scraper.Login(username, string(pass), code)
		}
		if !scraper.IsLoggedIn() {
			fmt.Println("Bad user/pass")
			continue
		}
		cookies := scraper.GetCookies()
		js, _ := json.Marshal(cookies)
		f, _ := os.OpenFile("twmd_cookies.json", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		defer f.Close()
		f.Write(js)
		break
	}
}

func Login() {
	if _, err := os.Stat("twmd_cookies.json"); errors.Is(err, fs.ErrNotExist) {
		askPass()
	} else {
		f, _ := os.Open("twmd_cookies.json")
		var cookies []*http.Cookie
		json.NewDecoder(f).Decode(&cookies)
		scraper.SetCookies(cookies)
	}
	if !scraper.IsLoggedIn() {
		askPass()
	}

}

func singleTweet(output string, id string) {
	tweet, err := scraper.GetTweet(id)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if usr != "" {
		if vidz {
			videoSingle(tweet, output)
		}
		if imgs {
			photoSingle(tweet, output)
		}
	} else {
		videoSingle(tweet, output)
		photoSingle(tweet, output)
	}
}

func main() {
	var nbr, single, output string
	var retweet, all, printversion, nologo, login bool
	op := optionparser.NewOptionParser()
	op.Banner = "twmd: Apiless twitter media downloader\n\nUsage:"
	op.On("-u", "--user USERNAME", "User you want to download", &usr)
	op.On("-t", "--tweet TWEET_ID", "Single tweet to download", &single)
	op.On("-n", "--nbr NBR", "Number of tweets to download", &nbr)
	op.On("-i", "--img", "Download images only", &imgs)
	op.On("-v", "--video", "Download videos only", &vidz)
	op.On("-a", "--all", "Download images and videos", &all)
	op.On("-r", "--retweet", "Download retweet too", &retweet)
	op.On("-z", "--url", "Print media url without download it", &urlOnly)
	op.On("-R", "--retweet-only", "Download only retweet", &onlyrtw)
	op.On("-s", "--size SIZE", "Choose size between small|normal|large (default large)", &size)
	op.On("-U", "--update", "Download missing tweet only", &update)
	op.On("-o", "--output DIR", "Output directory", &output)
	op.On("-L", "--login", "Login (needed for NSFW tweets)", &login)
	op.On("-p", "--proxy PROXY", "Use proxy (proto://ip:port)", &proxy)
	op.On("-V", "--version", "Print version and exit", &printversion)
	op.On("-B", "--no-banner", "Don't print banner", &nologo)
	op.Exemple("twmd -u Spraytrains -o ~/Downlaods -a -r -n 300")
	op.Exemple("twmd -u Spraytrains -o ~/Downlaods -R -U -n 300")
	op.Exemple("twmd --proxy socks5://127.0.0.1:9050 -t 156170319961391104")
	op.Exemple("twmd -t 156170319961391104")
	op.Parse()

	if printversion {
		fmt.Println("version:", version)
		os.Exit(1)
	}

	op.Logo("twmd", "elite", nologo)
	if usr == "" && single == "" {
		fmt.Println("You must specify an user (-u --user) or a tweet (-t --tweet)")
		op.Help()
		os.Exit(1)
	}
	if all {
		vidz = true
		imgs = true
	}
	if !vidz && !imgs && single == "" {
		fmt.Println("You must specify what to download. (-i --img) for images, (-v --video) for videos or (-a --all) for both")
		op.Help()
		os.Exit(1)
	}

	re := regexp.MustCompile("small|normal|large")
	if !re.MatchString(size) && size != "orig" {
		print("Error in size, setting up to normal\n")
		size = ""
	}
	if size == "large" {
		size = "orig"
	}

	client = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: time.Duration(5) * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   time.Duration(5) * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			DisableKeepAlives:     true,
		},
	}
	if proxy != "" {
		proxyURL, _ := URL.Parse(proxy)
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}

	scraper = twitterscraper.New()
	scraper.WithReplies(true)
	scraper.SetProxy(proxy)
	if login {
		Login()
	}

	if single != "" {
		if output == "" {
			output = "./"
		} else {
			os.MkdirAll(output, os.ModePerm)
		}
		singleTweet(output, single)
		os.Exit(0)
	}
	if nbr == "" {
		nbr = "3000"
	}
	if output != "" {
		output = output + "/" + usr
	} else {
		output = usr
	}
	if vidz {
		os.MkdirAll(output+"/video", os.ModePerm)
	}
	if imgs {
		os.MkdirAll(output+"/img", os.ModePerm)
	}
	nbrs, _ := strconv.Atoi(nbr)
	wg := sync.WaitGroup{}
	for tweet := range scraper.GetTweets(context.Background(), usr, nbrs) {
		if tweet.Error != nil {
			fmt.Println(tweet.Error)
			os.Exit(1)
		}
		if vidz {
			wg.Add(1)
			go videoUser(&wg, tweet, output, retweet)
		}
		if imgs {
			wg.Add(1)
			go photoUser(&wg, tweet, output, retweet)
		}
	}
	wg.Wait()
}
