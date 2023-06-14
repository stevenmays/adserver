package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Campaign struct {
	ID              int
	StartTimestamp  int64    `json:"start_timestamp"`
	EndTimestamp    int64    `json:"end_timestamp"`
	TargetKeywords  []string `json:"target_keywords"`
	MaxImpression   int      `json:"max_impression"`
	CPM             float64  `json:"cpm"`
	ImpressionCount int
	ImpressionIds    []string
}

type AdRequest struct {
	Keywords []string
}

// Might be better as env variable - but want this app to be simpler to run
var BASE_URL = "http://localhost:8000"

// Variable which exists
var campaigns []Campaign

/**
* When working with a Go API endpoint that interacts with a database, you generally don't need a mutex to manage concurrency, as the database itself handles concurrent access and transactions. Mutexes are primarily used for synchronizing access to in-memory data structures and shared resources in your Go program.
 */
var campaignMutex sync.Mutex

/** main function */
func main() {
	http.HandleFunc("/campaign", campaignHandler)
	http.HandleFunc("/addecision", adDecisionHandler)
	http.HandleFunc("/", impressionHandler)

	fmt.Println("Server is up")

	log.Fatal(http.ListenAndServe(":8000", nil))
}

/**
 * Creates a new campaign.
 */
func campaignHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var campaign Campaign
	err := json.NewDecoder(r.Body).Decode(&campaign)
	if err != nil || campaign.StartTimestamp == 0 || campaign.EndTimestamp == 0 || len(campaign.TargetKeywords) == 0 || campaign.MaxImpression == 0 || campaign.CPM <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	campaignMutex.Lock()
	campaign.ID = len(campaigns) + 1000 + 1
	campaigns = append(campaigns, campaign)
	campaignMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		CampaignID int `json:"campaign_id"`
	}{CampaignID: campaign.ID})
}

/**
* Creates an ad decision
*/
func adDecisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var adRequest AdRequest
	err := json.NewDecoder(r.Body).Decode(&adRequest)
	if err != nil || len(adRequest.Keywords) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	campaignMutex.Lock()
	defer campaignMutex.Unlock()

	now := time.Now().Unix()
	var selectedCampaign *Campaign
	for i, campaign := range campaigns {
		if now < campaign.StartTimestamp || now >= campaign.EndTimestamp {
			continue
		}

		if campaign.ImpressionCount >= campaign.MaxImpression {
			continue
		}

		if !hasCommonKeyword(campaign.TargetKeywords, adRequest.Keywords) {
			continue
		}

		// This can be a one liner, but breaking it up for readability
		if selectedCampaign == nil {
			selectedCampaign = &campaigns[i]
		} else if campaign.CPM > selectedCampaign.CPM {
			selectedCampaign = &campaigns[i]
		} else if (campaign.CPM == selectedCampaign.CPM && campaign.EndTimestamp < selectedCampaign.EndTimestamp) || (campaign.CPM == selectedCampaign.CPM && campaign.EndTimestamp == selectedCampaign.EndTimestamp && campaign.ID < selectedCampaign.ID) {
			selectedCampaign = &campaigns[i]
		}

	}

	if selectedCampaign == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	impressionID := generateUUID()

	selectedCampaign.ImpressionIds = append(selectedCampaign.ImpressionIds, impressionID)

	impressionURL := fmt.Sprintf("%s/%s", BASE_URL, impressionID)
	json.NewEncoder(w).Encode(struct {
		CampaignID    int    `json:"campaign_id"`
		ImpressionURL string `json:"impression_url"`
	}{CampaignID: selectedCampaign.ID, ImpressionURL: impressionURL})
}

/**
* Handles an impression callback
*/
func impressionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	impressionID := strings.TrimPrefix(r.URL.Path, "/")

	if len(impressionID) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	campaignMutex.Lock()
	defer campaignMutex.Unlock()

	found := false
	for i, campaign := range campaigns {
		if (containsImpression(campaign.ImpressionIds, impressionID)) {
			campaigns[i].ImpressionCount++
			found = true
			break
		}
	}

	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/**
* Evaluates campaign keywords and adRequestKeywords and determines if there's a match
 */
func hasCommonKeyword(campaignKeywords, adRequestKeywords []string) bool {
	for _, campaignKeyword := range campaignKeywords {
		for _, adRequestKeyword := range adRequestKeywords {
			if campaignKeyword == adRequestKeyword {
				return true
			}
		}
	}
	return false
}

/**
* Generate a uuid and throw if the package fails to generate
 */
func generateUUID() string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("Failed to generate UUID: %v", err)
	}
	return uuid.String()
}

func containsImpression(impressionIDs []string, impressionID string) bool {
    for _, a := range impressionIDs {
        if a == impressionID {
            return true
        }
    }
    return false
}
