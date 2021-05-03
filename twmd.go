package main

import (
  "context"
  "fmt"
  twitterscraper "github.com/n0madic/twitter-scraper"
  "github.com/speedata/optionparser"
  "io"
  "net/http"
  "os"
  "strconv"
  "strings"
  "sync"
)

var wg sync.WaitGroup
var mwg sync.WaitGroup

func download(url string, filetype string, output string, dwn_type string) {
  segments := strings.Split(url, "/")
  name := segments[len(segments)-1]
  resp, _ := http.Get(url)
  if resp.StatusCode != 200 {
    return
  }
  var f *os.File
  if dwn_type == "user" {
    f, _ = os.Create(output + "/" + filetype + "/" + name)
  } else {
    f, _ = os.Create(output + "/" + name)
  }
  defer f.Close()
  defer resp.Body.Close()
  io.Copy(f, resp.Body)
  fmt.Println("Downloaded " + name)
  wg.Done()
}

func vidUrl(video string) string {
  vid := strings.Split(string(video), " ")
  v := vid[len(vid)-1]
  v = strings.TrimSuffix(v, "}")
  vid = strings.Split(v, "?")
  return vid[0]
}

func videoUser(tweet *twitterscraper.TweetResult, output string, rt bool, dwn_tweet string) {
  if len(tweet.Videos) > 0 {
    for _, i := range tweet.Videos {
      j := fmt.Sprintf("%s", i)
      if tweet.IsRetweet {
        if rt {
          v := vidUrl(j)
          wg.Add(1)
          go download(v, "video", output, "user")
        } else {
          continue
        }
      }
      v := vidUrl(j)
      wg.Add(1)
      go download(v, "video", output, "user")
    }
    wg.Wait()
  }
  mwg.Done()
}

func photoUser(tweet *twitterscraper.TweetResult, output string, rt bool, dwn_type string) {
  if len(tweet.Photos) > 0 {
    for _, i := range tweet.Photos {
      i := i
      if !strings.Contains(i, "video_thumb/") {
        if tweet.IsRetweet {
          if rt {
            wg.Add(1)
            go download(i, "img", output, "user")
          } else {
            continue
          }
        }
        wg.Add(1)
        go download(i, "img", output, "user")
      }
    }
    wg.Wait()
  }
  mwg.Done()
}

func videoSingle(tweet *twitterscraper.Tweet, output string, dwn_tweet string) {
  if len(tweet.Videos) > 0 {
    for _, i := range tweet.Videos {
      j := fmt.Sprintf("%s", i)
      v := vidUrl(j)
      wg.Add(1)
      go download(v, "tweet", output, "tweet")
    }
    wg.Wait()
  }
}

func photoSingle(tweet *twitterscraper.Tweet, output string, dwn_type string) {
  if len(tweet.Photos) > 0 {
    for _, i := range tweet.Photos {
      if !strings.Contains(i, "video_thumb/") {
        wg.Add(1)
        download(i, "tweet", output, "tweet")
      }
    }
    wg.Wait()
  }
}

func singleTweet(output string, id string) {
  scraper := twitterscraper.New()
  tweet, err := scraper.GetTweet(id)
  if err != nil {
    fmt.Println(err)
  }
  videoSingle(tweet, output, "tweet")
  photoSingle(tweet, output, "tweet")
}

func help() {
  fmt.Println(`twd: Apiless twitter media downloader

usage:
-h, --help                   Show this help
-u, --user     USERNAME      User you want to download
-t, --tweet    TWEET_ID      Single tweet to download
-n, --nbr      NBR           Number of tweet to download
-i, --img                    Download images only
-v, --video                  Download video only
-r, --retweet                Download retweet
-o, --output   DIR           Output direcory

ex:
twmd -u Spraytrains -o ~/Downlaods -i -v -r -n 300
twmd -t 156170319961391104`)
  os.Exit(1)
}

func main() {
  var usr, nbr, single, output string
  var vidz, imgs, retweet bool
  op := optionparser.NewOptionParser()
  op.On("-u", "--user USERNAME", "User you want to download media", &usr)
  op.On("-t", "--tweet TWEET_ID", "Single tweet to download", &single)
  op.On("-n", "--nbr NBR", "Number of tweet to download", &nbr)
  op.On("-i", "--img", "Download images only", &imgs)
  op.On("-v", "--video", "Download video only", &vidz)
  op.On("-r", "--retweet", "Download retweet", &retweet)
  op.On("-o", "--output DIR", "Output direcory", &output)
  op.On("-h", "--help", "Show this help", help)
  op.Parse()
  if usr == "" && single == "" {
    fmt.Println("You must specify an user (-u --user) or a tweet (-t --tweet)\n")
    help()
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
    nbr = "50"
  }
  if output != "" {
    output = output + "/" + usr
  } else {
    output = usr
  }
  os.MkdirAll(output+"/video", os.ModePerm)
  os.MkdirAll(output+"/img", os.ModePerm)
  nbrs, _ := strconv.Atoi(nbr)
  scraper := twitterscraper.New()
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
