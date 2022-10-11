package main

import (
	"context"
	"errors"
	"fmt"
	"io"
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
)

var (
	wg      sync.WaitGroup
	mwg     sync.WaitGroup
	usr     string
	proxy   string
	update  bool
	onlyrtw bool
	vidz    bool
	imgs    bool
	urlOnly bool
	version = "1.0.5"
	client  *http.Client
	size    = "orig"
)

func download(url string, filetype string, output string, dwn_type string) {
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

func videoUser(tweet *twitterscraper.TweetResult, output string, rt bool, dwn_tweet string) {
	defer mwg.Done()
	if len(tweet.Videos) > 0 {
		for _, i := range tweet.Videos {
			j := fmt.Sprintf("%s", i)
			if tweet.IsRetweet {
				if rt || onlyrtw {
					v := vidUrl(j)
					wg.Add(1)
					go download(v, "video", output, "user")
				} else {
					continue
				}
			} else if onlyrtw {
				continue
			}
			v := vidUrl(j)
			wg.Add(1)
			go download(v, "video", output, "user")
		}
		wg.Wait()
	}
}

func photoUser(tweet *twitterscraper.TweetResult, output string, rt bool, dwn_type string) {
	defer mwg.Done()
	if len(tweet.Photos) > 0 || tweet.IsRetweet {
		if tweet.IsRetweet && (rt || onlyrtw) {
			singleTweet(output, tweet.ID)
		}
		for _, i := range tweet.Photos {
			if onlyrtw || tweet.IsRetweet {
				continue
			}
			if !strings.Contains(i, "video_thumb/") {
				if size == "orig" || size == "small" {
					i = i + "?name=" + size
				}
				wg.Add(1)
				go download(i, "img", output, "user")
			}
		}
		wg.Wait()
	}
}

func videoSingle(tweet *twitterscraper.Tweet, output string, dwn_tweet string) {
	if len(tweet.Videos) > 0 {
		for _, i := range tweet.Videos {
			j := fmt.Sprintf("%s", i)
			v := vidUrl(j)
			wg.Add(1)
			if usr != "" {
				go download(v, "rtvideo", output, "user")
			} else {
				go download(v, "tweet", output, "tweet")
			}
		}
		wg.Wait()
	}
}

func photoSingle(tweet *twitterscraper.Tweet, output string, dwn_type string) {
	if len(tweet.Photos) > 0 {
		for _, i := range tweet.Photos {
			if !strings.Contains(i, "video_thumb/") {
				if size == "orig" || size == "small" {
					i = i + "?name=" + size
				}
				wg.Add(1)
				if usr != "" {
					go download(i, "rtimg", output, "user")
				} else {
					go download(i, "tweet", output, "tweet")
				}
			}
		}
		wg.Wait()
	}
}

func singleTweet(output string, id string) {
	scraper := twitterscraper.New()
	scraper.SetProxy(proxy)
	tweet, err := scraper.GetTweet(id)
	if err != nil {
		fmt.Println(err)
	}
	if usr != "" {
		if vidz {
			videoSingle(tweet, output, "video")
		}
		if imgs {
			photoSingle(tweet, output, "img")
		}
	} else {
		videoSingle(tweet, output, "tweet")
		photoSingle(tweet, output, "tweet")
	}
}

func main() {
	var nbr, single, output string
	var retweet, all, printversion, nologo bool
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
	scraper := twitterscraper.New()
	// do nothing if proxy = ""
	scraper.SetProxy(proxy)
	for tweet := range scraper.GetTweets(context.Background(), usr, nbrs) {
		if tweet.Error != nil {
			fmt.Println(tweet.Error)
			os.Exit(1)
		}
		if vidz {
			mwg.Add(1)
			go videoUser(tweet, output, retweet, "user")
		}
		if imgs {
			mwg.Add(1)
			go photoUser(tweet, output, retweet, "user")
		}
	}
	mwg.Wait()
}
