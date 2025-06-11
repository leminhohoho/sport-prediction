package helpers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Generate a random time duration between min and max
func GetRandomTime(min time.Duration, max time.Duration) time.Duration {
	timeRange := max - min

	return min + time.Duration(rand.Intn(int(timeRange)))
}

// Generate a randomized order of an array of integer values from 0 to modulo-1
func RandomizeCyclicGroup(modulo int) []int {
	var arr []int

	for i := 0; i < modulo; i++ {
		arr = append(arr, i)
	}

	for i := 0; i < len(arr); i++ {
		newRandomIndex := rand.Intn(len(arr))
		arr[i], arr[newRandomIndex] = arr[newRandomIndex], arr[i]
	}

	return arr
}

// Run a callback on a pair of match element from 2 parallel goquery doc
// provided that the structure till the selector from both docs are the same
func Find2Docs(
	selector string,
	doc1 *goquery.Document,
	doc2 *goquery.Document,
	f func(node1 *goquery.Selection, node2 *goquery.Selection),
) {
	var nodes1 []*goquery.Selection
	var nodes2 []*goquery.Selection

	doc1.Find(selector).Each(func(_ int, node *goquery.Selection) {
		nodes1 = append(nodes1, node)
	})
	doc2.Find(selector).Each(func(_ int, node *goquery.Selection) {
		nodes2 = append(nodes2, node)
	})

	if len(nodes1) == 0 {
		fmt.Println("No selector found here")
		return
	}

	for i := range nodes1 {
		f(nodes1[i], nodes2[i])
	}
}

// Classify tournament tier
func IsBigEvent(tournamentName string, prizepool int) bool {
	tier1Phrases := []string{
		"Valorant Master",
		"Valorant Champion",
		"Champion Tour",
	}

	for _, phrase := range tier1Phrases {
		if strings.Contains(tournamentName, phrase) {
			return true
		}
	}

	if prizepool > 100000 {
		return true
	}

	return false
}

func GetPlayerName(url string) (string, error) {
	res, err := http.Get("https://www.vlr.gg" + url)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	name := strings.TrimSpace(doc.Find(`#wrapper > div.col-container > div > div.wf-card.mod-header.mod-full > div.player-header > div:nth-child(2) > div:nth-child(1) > h1`).Text())
	name = strings.ToLower(name)
	name = strings.Replace(name, " ", "-", -1)

	return name, nil
}

func GetTeamName(url string) (string, error) {
	res, err := http.Get("https://www.vlr.gg" + url)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	name := strings.TrimSpace(doc.Find(`#wrapper > div.col-container > div > div.wf-card.mod-header.mod-full > div.team-header > div.team-header-desc > div > div.team-header-name > h1`).Text())
	name = strings.ToLower(name)
	name = strings.Replace(name, " ", "-", -1)

	return name, nil
}
