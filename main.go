package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/adrg/xdg"

	gobot "github.com/danrusei/gobot-bsky"
)

type lastfm struct {
	Key      string `json:"key"`
	Secret   string
	Username string
}

type blueskyConfig struct {
	handle string
	apikey string
	server string
}

type secrets struct {
	Lastfm lastfm
	Bsky   blueskyConfig
}

func getSecrets() secrets {
	configFilePath, err := xdg.ConfigFile("lastfmbluesky/secrets.json")
	if err != nil {
		fmt.Println("error")
	}
	settingsJson, err := os.Open(configFilePath)
	// if os.Open returns an error then handle it
	if err != nil {
		fmt.Println("Unable to open the config file. Did you place it in the right spot?")

	}
	defer func(settingsJson *os.File) {
		err := settingsJson.Close()
		if err != nil {
			errorString := fmt.Sprintf("Couldn't close the settings file. Error: %s", err)
			fmt.Println(errorString)

		}
	}(settingsJson)
	byteValue, _ := io.ReadAll(settingsJson)
	var settings *secrets
	err = json.Unmarshal(byteValue, &settings)
	if err != nil {
		fmt.Println("Check that you do not have errors in your JSON file.")
		errorString := fmt.Sprintf("Could not unmashal json: %s\n", err)
		fmt.Println(errorString)
		panic("AAAAAAH!")
	}
	return *settings
}

type attribute struct {
	Rank string
}

type overallAttribute struct {
	User       string
	totalPages string
	page       string
	perPage    string
	Total      string
}

type artist struct {
	Playcount string
	Attribute attribute `json:"@attr"`
	Name      string
}

type topArtists struct {
	Artist    []artist
	Attribute overallAttribute `json:"@attr"`
}

type topArtistsResult struct {
	Topartists topArtists
}

func submitLastfmCommand(period string, apiKey string, user string) (string, error) {
	apiURLBase := "https://ws.audioscrobbler.com/2.0/?"
	queryParameters := url.Values{}
	queryParameters.Set("method", "user.gettopartists")
	queryParameters.Set("user", user)
	switch period {
	case "weekly":
		queryParameters.Set("period", "7day")
	case "annual":
		queryParameters.Set("period", "12month")
	case "quarterly":
		queryParameters.Set("period", "3month")
	}
	queryParameters.Set("api_key", apiKey)
	queryParameters.Set("format", "json")
	fullURL := apiURLBase + queryParameters.Encode()
	lastfmResponse, statusCode, err := WebGet(fullURL)
	if err != nil {
		fmt.Println(statusCode)
		return lastfmResponse, err
	}
	return lastfmResponse, err
}

// webGet handles contacting a URL
func WebGet(url string) (string, int, error) {
	response, err := http.Get(url)
	if err != nil {
		return "Error accessing URL", 0, err
	}
	result, err := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode > 299 {
		statusCodeString := fmt.Sprintf("Response failed with status code: %d and \nbody: %s\n", response.StatusCode, result)
		fmt.Println(statusCodeString)
		panic("Invalid status, data will be garbage")
	}
	if err != nil {
		return "Error reading response", 0, err
	}
	return string(result), response.StatusCode, err

}

func assembleBskyPost(artists topArtistsResult, period string) string {
	var postString string
	switch period {
	case "weekly":
		postString = fmt.Sprintf("#music Out of %s songs, my top #lastfm artists for the past week: ", artists.Topartists.Attribute.Total)
	case "annual":
		postString = fmt.Sprintf("#music Out of %s songs, my top #lastfm artists for the past 12 months: ", artists.Topartists.Attribute.Total)
	case "quarterly":
		postString = fmt.Sprintf("#music Out of %s songs, my top #lastfm artists for the past 3 months: ", artists.Topartists.Attribute.Total)
	}
	for _, artist := range artists.Topartists.Artist {
		potentialString := fmt.Sprintf("%s.%s (%s), ", artist.Attribute.Rank, artist.Name, artist.Playcount)
		if len(postString)+len(potentialString) < 240 {
			postString += potentialString
		} else {
			return postString
		}
	}
	return postString
}

func main() {
	ourSecrets := getSecrets()
	// parse CLI flags
	period := flag.String("p", "weekly", "period to grab. Use: weekly, quarterly, or annual")
	debugMode := flag.Bool("d", false, "register the client")
	flag.Parse()

	weeklyArtistsJSON, err := submitLastfmCommand(*period, ourSecrets.Lastfm.Key, ourSecrets.Lastfm.Username)
	if err != nil {
		fmt.Println(err)
	}
	var weeklyArtsts topArtistsResult
	err = json.Unmarshal([]byte(weeklyArtistsJSON), &weeklyArtsts)
	if err != nil {
		fmt.Printf("Unable to marshall. %s", err)
	}
	postString := assembleBskyPost(weeklyArtsts, *period)
	fmt.Printf("Your toot will be: %s\n\n", postString)

	ctx := context.Background()

	agent := gobot.NewAgent(ctx, ourSecrets.Bsky.server, ourSecrets.Bsky.handle, ourSecrets.Bsky.apikey)
	agent.Connect(ctx)

	post, err := gobot.NewPostBuilder("%s", postString).
		Build()
	if err != nil {
		fmt.Printf("Got error: %v", err)
	}

	if *debugMode {
		fmt.Printf("post will be: %v", post)
	} else {

		cid1, uri1, err := agent.PostToFeed(ctx, post)
		if err != nil {
			fmt.Printf("Got error: %v", err)
		} else {
			fmt.Printf("Succes: Cid = %v , Uri = %v", cid1, uri1)
		}
	}
}
