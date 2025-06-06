package main

import (
	"bufio"
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
	"unicode/utf8"

	twitterscraper "github.com/imperatrona/twitter-scraper"
	"github.com/mmpx12/optionparser"
)

var (
	usr     string
	format  string
	proxy   string
	update  bool
	onlyrtw bool
	onlymtw bool
	vidz    bool
	gifs    bool
	imgs    bool
	urlOnly bool
	version = "1.14.2"
	scraper *twitterscraper.Scraper
	client  *http.Client
	size    = "orig"
	datefmt = "2006-01-02"
)

func download(wg *sync.WaitGroup, tweet interface{}, url string, filetype string, output string, dwn_type string) {
	defer wg.Done()
	segments := strings.Split(url, "/")
	name := segments[len(segments)-1]
	re := regexp.MustCompile(`name=`)
	if re.MatchString(name) {
		segments := strings.Split(name, "?")
		name = segments[len(segments)-2]
	}
	if format != "" {
		name = getFormat(tweet) + "_" + name
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
		} else if filetype == "rtgif" {
			f, _ = os.Create(output + "/gif/RE-" + name)
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
	defer f.Close()
	io.Copy(f, resp.Body)
	fmt.Println("Downloaded " + name)
}

func videoUser(wait *sync.WaitGroup, tweet *twitterscraper.TweetResult, output string, rt bool) {
	defer wait.Done()
	wg := sync.WaitGroup{}
	if len(tweet.Videos) > 0 {
		for _, i := range tweet.Videos {
			url := strings.Split(i.URL, "?")[0]
			if tweet.IsRetweet {
				if rt || onlyrtw {
					wg.Add(1)
					go download(&wg, tweet, url, "video", output, "user")
					continue
				} else {
					continue
				}
			} else if onlyrtw {
				continue
			}
			wg.Add(1)
			go download(&wg, tweet, url, "video", output, "user")
		}
		wg.Wait()
	}
}

func gifUser(wait *sync.WaitGroup, tweet *twitterscraper.TweetResult, output string, rt bool) {
	defer wait.Done()
	wg := sync.WaitGroup{}
	if len(tweet.GIFs) > 0 {
		for _, i := range tweet.GIFs {
			url := i.URL
			if tweet.IsRetweet {
				if rt || onlyrtw {
					wg.Add(1)
					go download(&wg, tweet, url, "video", output, "user")
					continue
				} else {
					continue
				}
			} else if onlyrtw {
				continue
			}
			wg.Add(1)
			go download(&wg, tweet, url, "video", output, "user")
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
				go download(&wg, tweet, url, "img", output, "user")
			}
		}
		wg.Wait()
	}
}

func videoSingle(tweet *twitterscraper.Tweet, output string) {
	if tweet == nil {
		return
	}
	if len(tweet.Videos) > 0 {
		wg := sync.WaitGroup{}
		for _, i := range tweet.Videos {
			url := strings.Split(i.URL, "?")[0]
			if usr != "" {
				wg.Add(1)
				go download(&wg, tweet, url, "rtvideo", output, "user")
			} else {
				wg.Add(1)
				go download(&wg, tweet, url, "tweet", output, "tweet")
			}
		}
		wg.Wait()
	}
}

func gifSingle(tweet *twitterscraper.Tweet, output string) {
	if tweet == nil {
		return
	}
	if len(tweet.GIFs) > 0 {
		wg := sync.WaitGroup{}
		for _, i := range tweet.GIFs {
			url := i.URL
			if usr != "" {
				wg.Add(1)
				go download(&wg, tweet, url, "rtgif", output, "user")
			} else {
				wg.Add(1)
				go download(&wg, tweet, url, "tweet", output, "tweet")
			}
		}
		wg.Wait()
	}
}

func photoSingle(tweet *twitterscraper.Tweet, output string) {
	if tweet == nil {
		return
	}
	if len(tweet.Photos) > 0 {
		wg := sync.WaitGroup{}
		for _, i := range tweet.Photos {
			var url string
			if !strings.Contains(i.URL, "video_thumb/") {
				if size == "orig" || size == "small" {
					url = i.URL + "?name=" + size
				} else {
					url = i.URL
				}
				if usr != "" {
					wg.Add(1)
					go download(&wg, tweet, url, "rtimg", output, "user")
				} else {
					wg.Add(1)
					go download(&wg, tweet, url, "tweet", output, "tweet")
				}
			}
		}
		wg.Wait()
	}
}

func processCookieString(cookieStr string) []*http.Cookie {
	cookiePairs := strings.Split(cookieStr, "; ")
	cookies := make([]*http.Cookie, 0)
	expiresTime := time.Now().AddDate(1, 0, 0)

	for _, pair := range cookiePairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		value := parts[1]
		value = strings.Trim(value, "\"")

		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Path:     "/",
			Domain:   ".x.com",
			Expires:  expiresTime,
			HttpOnly: true,
			Secure:   true,
		}

		cookies = append(cookies, cookie)
	}
	return cookies
}

func askPass() {
	for {
		var auth_token, ct0 string
		fmt.Println(`  ╔═══════════════════════════════════════════════════════════════╗
  ║                                                               ║
  ║  User/pass login is no longer supported,                      ║
  ║  Log in using a browser and find auth_token and ct0 cookies.  ║
  ║  (via Inspect → Storage → Cookies).                           ║
  ║                                                               ║
  ╚═══════════════════════════════════════════════════════════════╝`)
		fmt.Println()
		fmt.Printf("auth_token cookie: ")
		fmt.Scanln(&auth_token)
		fmt.Printf("ct0 cookie: ")
		fmt.Scanln(&ct0)
		scraper.SetAuthToken(twitterscraper.AuthToken{Token: auth_token, CSRFToken: ct0})
		if !scraper.IsLoggedIn() {
			fmt.Println("Bad Cookies.")
			askPass()
		}
		cookies := scraper.GetCookies()
		js, _ := json.Marshal(cookies)
		f, _ := os.OpenFile("twmd_cookies.json", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		defer f.Close()
		f.Write(js)
		break
	}
}

func Login(useCookies bool) {
	if useCookies {
		if _, err := os.Stat("twmd_cookies.json"); errors.Is(err, fs.ErrNotExist) {
			fmt.Print("Enter cookies string: ")
			var cookieStr string
			cookieStr, _ = bufio.NewReader(os.Stdin).ReadString('\n')
			cookieStr = strings.TrimSpace(cookieStr)

			cookies := processCookieString(cookieStr)
			scraper.SetCookies(cookies)

			// Save cookies to file
			js, _ := json.MarshalIndent(cookies, "", "  ")
			f, _ := os.OpenFile("twmd_cookies.json", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			defer f.Close()
			f.Write(js)
		} else {
			f, _ := os.Open("twmd_cookies.json")
			var cookies []*http.Cookie
			json.NewDecoder(f).Decode(&cookies)
			scraper.SetCookies(cookies)
			fmt.Println(scraper.IsLoggedIn())
		}
	} else {
		if _, err := os.Stat("twmd_cookies.json"); errors.Is(err, fs.ErrNotExist) {
			askPass()
		} else {
			f, _ := os.Open("twmd_cookies.json")
			var cookies []*http.Cookie
			json.NewDecoder(f).Decode(&cookies)
			scraper.SetCookies(cookies)
		}
	}

	if !scraper.IsLoggedIn() {
		if useCookies {
			fmt.Println("Invalid cookies. Please try again.")
			os.Remove("twmd_cookies.json")
			Login(useCookies)
		} else {
			askPass()
		}
	} else {
		fmt.Println("Logged in.")
	}
}

func singleTweet(output string, id string) {
	tweet, err := scraper.GetTweet(id)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if tweet == nil {
		fmt.Println("Error retrieve tweet")
		return
	}
	if usr != "" {
		if vidz {
			videoSingle(tweet, output)
		}
		if gifs {
			gifSingle(tweet, output)
		}
		if imgs {
			photoSingle(tweet, output)
		}
	} else {
		videoSingle(tweet, output)
		gifSingle(tweet, output)
		photoSingle(tweet, output)
	}
}

func getFormat(tweet interface{}) string {
	var formatNew string
	var tweetResult *twitterscraper.TweetResult
	var tweetObj *twitterscraper.Tweet

	switch t := tweet.(type) {
	case *twitterscraper.TweetResult:
		tweetResult = t
	case *twitterscraper.Tweet:
		tweetObj = t
	default:
		fmt.Println("Invalid tweet type")
		return ""
	}

	pattern := `[/\\:*?"<>|]`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Error compiling regular expression:", err)
		return ""
	}

	replacer := map[string]string{}

	if tweetResult != nil {
		replacer["{DATE}"] = time.Unix(tweetResult.Timestamp, 0).Format(datefmt)
		replacer["{NAME}"] = tweetResult.Name
		replacer["{USERNAME}"] = tweetResult.Username
		replacer["{TITLE}"] = sanitizeText(tweetResult.Text, regex, 255)
		replacer["{ID}"] = tweetResult.ID
	} else if tweetObj != nil {
		replacer["{DATE}"] = time.Unix(tweetObj.Timestamp, 0).Format(datefmt)
		replacer["{NAME}"] = tweetObj.Name
		replacer["{USERNAME}"] = tweetObj.Username
		replacer["{TITLE}"] = sanitizeText(tweetObj.Text, regex, 255)
		replacer["{ID}"] = tweetObj.ID
	}

	formatNew = format

	for key, val := range replacer {
		formatNew = strings.ReplaceAll(formatNew, key, val)
	}

	return formatNew
}

func sanitizeText(text string, regex *regexp.Regexp, maxLen int) string {
	cleaned := ""
	remaining := maxLen
	for _, char := range text {
		charStr := string(char)
		if regex.MatchString(charStr) {
			charStr = "_"
		}
		if utf8.RuneCountInString(cleaned)+utf8.RuneCountInString(charStr) > remaining {
			break
		}
		cleaned += charStr
	}
	return cleaned
}

func main() {
	var nbr, single, output string
	var retweet, all, printversion, nologo, login, useCookies bool
	op := optionparser.NewOptionParser()
	op.Banner = "twmd: Apiless twitter media downloader\n\nUsage:"
	op.On("-u", "--user USERNAME", "User you want to download", &usr)
	op.On("-t", "--tweet TWEET_ID", "Single tweet to download", &single)
	op.On("-n", "--nbr NBR", "Number of tweets to download", &nbr)
	op.On("-i", "--img", "Download images only", &imgs)
	op.On("-v", "--video", "Download videos only", &vidz)
	op.On("-g", "--gif", "Download gifs only", &gifs)
	op.On("-a", "--all", "Download images, videos and gifs", &all)
	op.On("-r", "--retweet", "Download retweet too", &retweet)
	op.On("-z", "--url", "Print media url without download it", &urlOnly)
	op.On("-R", "--retweet-only", "Download only retweet", &onlyrtw)
	op.On("-M", "--mediatweet-only", "Download only media tweet", &onlymtw)
	op.On("-s", "--size SIZE", "Choose size between small|normal|large (default large)", &size)
	op.On("-U", "--update", "Download missing tweet only", &update)
	op.On("-o", "--output DIR", "Output directory", &output)
	op.On("-f", "--file-format FORMAT", "Formatted name for the downloaded file, {DATE} {USERNAME} {NAME} {TITLE} {ID}", &format)
	op.On("-d", "--date-format FORMAT", "Apply custom date format. (https://go.dev/src/time/format.go)", &datefmt)
	op.On("-L", "--login", "Login (needed for NSFW tweets)", &login)
	op.On("-C", "--cookies", "Use cookies for authentication", &useCookies)
	op.On("-p", "--proxy PROXY", "Use proxy (proto://ip:port)", &proxy)
	op.On("-V", "--version", "Print version and exit", &printversion)
	op.On("-B", "--no-banner", "Don't print banner", &nologo)
	op.Exemple("twmd -u Spraytrains -o ~/Downloads -a -r -n 300")
	op.Exemple("twmd -u Spraytrains -o ~/Downloads -R -U -n 300")
	op.Exemple("twmd --proxy socks5://127.0.0.1:9050 -t 156170319961391104")
	op.Exemple("twmd -t 156170319961391104")
	op.Exemple("twmd -t 156170319961391104 -f \"{DATE} {ID}\"")
	op.Exemple("twmd -t 156170319961391104 -f \"{DATE} {ID}\" -d \"2006-01-02_15-04-05\"")
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
		gifs = true
		imgs = true
	}
	if !vidz && !gifs && !imgs && single == "" {
		fmt.Println("You must specify what to download. (-i --img) for images, (-v --video) for videos, (-g --gif) for gifs or (-a --all) for both")
		op.Help()
		os.Exit(1)
	}
	var re = regexp.MustCompile(`{ID}|{DATE}|{NAME}|{USERNAME}|{TITLE}`)
	if format != "" && !re.MatchString(format) {
		fmt.Println("You must specify a format (-f --format)")
		op.Help()
		os.Exit(1)
	}

	re = regexp.MustCompile("small|normal|large")
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

	// Modified login handling
	if login || useCookies {
		Login(useCookies)
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
	if gifs {
		os.MkdirAll(output+"/gif", os.ModePerm)
	}
	if imgs {
		os.MkdirAll(output+"/img", os.ModePerm)
	}
	nbrs, _ := strconv.Atoi(nbr)
	wg := sync.WaitGroup{}

	var tweets <-chan *twitterscraper.TweetResult
	if onlymtw {
		tweets = scraper.GetMediaTweets(context.Background(), usr, nbrs)
	} else {
		tweets = scraper.GetTweets(context.Background(), usr, nbrs)
	}

	for tweet := range tweets {
		if tweet.Error != nil {
			fmt.Println(tweet.Error)
			os.Exit(1)
		}
		if vidz {
			wg.Add(1)
			go videoUser(&wg, tweet, output, retweet)
		}
		if gifs {
			wg.Add(1)
			go gifUser(&wg, tweet, output, retweet)
		}
		if imgs {
			wg.Add(1)
			go photoUser(&wg, tweet, output, retweet)
		}
	}
	wg.Wait()
}
