# twmd: twitter media downloader (without api key)

This twitter downloader doesn't require Credentials or an api key. It's based on [twitter-scrapper](https://github.com/n0madic/twitter-scraper). 

Unfortunately, you will not be able to download more than 3200 tweets.


## usage: 

```
-h, --help                   Show this help
-u, --user     USERNAME      User you want to download
-t, --tweet    TWEET_ID      Single tweet to download
-n, --nbr      NBR           Number of tweet to download (default 3000)
-i, --img                    Download images only
-v, --video                  Download video only
-a, --all                    Download video and img
-r, --retweet                Download retweet
-R, --only-retweet           Download retweet only
-z, --url                    Print media url without download it
-U, --update                 Download missing tweet only
-o, --output   DIR           Output directory
-B, --no-banner              Don't print banner
-V, --version                Print version and exit
```


### Examples:

#### Download 300 tweets from @Spraytrains.

If the tweet doesn't contain a photo or video nothing will be downloaded but it will count towards the 300.

```sh
twmd -u Spraytrains -o ~/Downlaods -a -n 3000
```

You can use `-r|--retweet` to download retweets as well, or `-R|--retweet-only` to download retweet only

`-U|--update` will only download missing media.

#### Download a single tweet:

```sh
twmd -t 156170319961391104
```

### Installation:

**Note:** If you don't want to build it you can download prebuilt binaries [here](https://github.com/mmpx12/twitter-media-downloader/releases/latest).


```sh
git clone https://github.com/mmpx12/twitter-media-downloader.git
cd twitter-media-downloader
make
sudo make install
# OR
sudo make all
# Clean
sudo make clean
```

#### Termux (no root):

installation: 

```sh
git clone https://github.com/mmpx12/twitter-media-downloader.git
cd twitter-media-downloader
make
make termux-install
# OR
make termux-all
# Clean
make termux-clean
```

You may also want to add stuff in ~/bin/termux-url-opener to automatically download profile or post when share with termux.

```sh
cd ~/storage/downlaods
if grep twitter <<< "$1" >/dev/null; then
  if [[ $(tr -cd '/' <<< "$1" | wc -c) -eq 3 ]]; then
    userid=$(cut -d '/' -f 4 <<< "$1" |  cut -d '?' -f 1)
    echo "$userid"
    twmd -B -u "$userid" -o twitter -i -v -n 3000
  else 
    postid=$(cut -d '/' -f 6 <<< "$1" |  cut -d '?' -f 1)
    twmd -B -t "$postid" -o twitter
  fi
fi
```


Check [here](https://gist.github.com/mmpx12/f0741d40909ed3f182fd6f9b33b580d7) for a full termux-url-opener example.


#### Gifs are not supported at the moment.
