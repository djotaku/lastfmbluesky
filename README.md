# lastfmBluesky

## Usage 
Post your weekly and/or yearly last.fm stats to Mastodon


- For last.fm get your key and secret at: https://www.last.fm/api/account/create (more about their API at: https://www.last.fm/api)
- At $HOME/.config/lastfmbluesky you should have a secrets.json file that looks like:


```json

{
        "lastfm":
                {
                        "key": "last.fm key",
                        "secret": "last.fm secret",
                        "username": "last.fm username"
                },
        "bsky":
            {
                    "Handle": "username.bsky.social",
                    "Sever": "URL of your bluesky instance - bsky.social",
                    "APIkey": "This is your app password from from the bluesky website"
            }
}


```

## Changes coming

Will post your top listened artists to Bluesky. There's a lot of overlap in the last.fm code with [lastfmmastodon](https://github.com/djotaku/lastfmmastodon) as of version 1.0. I intend to extract the last.fm code into its own library eventually.