# twmd: twitter media downlaoder (without api key)

This twitter downlaoader will not require Creds or api key. Its based on [twitter-scrapper](https://github.com/n0madic/twitter-scraper). 

Unfortunately you will not be able to download more than 3200 tweets.


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
-U, --update                 Downlaod missing tweet only
-o, --output   DIR           Output direcory
-B, --no-banner              Don't print banner
-V, --version                Print version and exit
```


### Exemples:

#### Download 300 tweets from @Spraytrains.

If tweet don't contains photo or video nothing will be download but it will be count in the 300.

```sh
twmd -u Spraytrains -o ~/Downlaods -a -n 3000
```

You can use `-r|--retweet` for download retweets too or `-R|--retweet-only` for downoad retweet only

`-U|--update` will only download missing media.

#### Download a single tweet:

```sh
twmd -t 156170319961391104
```

### Insallation:

**Note:** If you don't want to build it you can download prebuild binaries [here](https://github.com/mmpx12/twitter-media-downloader/releases/latest).


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

You may also want to add stuff in ~/bin/termux-url-opener for automaticly download profile or post when share with termux.

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


Chech [here](https://gist.github.com/mmpx12/f0741d40909ed3f182fd6f9b33b580d7) for a full termux-url-opener exemple.


#### Gif aren't support for the moment.
