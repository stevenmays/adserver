package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCampaignHandler(t *testing.T) {
	// Set up test server
	ts := httptest.NewServer(http.HandlerFunc(campaignHandler))
	defer ts.Close()

	// Test cases
	tests := []struct {
		name           string
		campaign       Campaign
		expectedStatus int
	}{
		{
			name: "Valid Campaign",
			campaign: Campaign{
				StartTimestamp: time.Now().Add(time.Hour).Unix(),
				EndTimestamp:   time.Now().Add(2 * time.Hour).Unix(),
				TargetKeywords: []string{"shampoo"},
				MaxImpression:  100,
				CPM:            10,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Campaign - Missing StartTimestamp",
			campaign: Campaign{
				EndTimestamp:   time.Now().Add(2 * time.Hour).Unix(),
				TargetKeywords: []string{"shampoo"},
				MaxImpression:  100,
				CPM:            10,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Campaign - Missing EndTimestamp",
			campaign: Campaign{
				StartTimestamp: time.Now().Add(time.Hour).Unix(),
				TargetKeywords: []string{"shampoo"},
				MaxImpression:  100,
				CPM:            10,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Encode campaign struct as JSON
			jsonData, err := json.Marshal(tc.campaign)
			assert.NoError(t, err)

			// Send POST request to the server
			resp, err := http.Post(fmt.Sprintf("%s/campaign", ts.URL), "application/json", bytes.NewReader(jsonData))
			assert.NoError(t, err)

			// Check response status code
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			// Check response body for valid campaigns
			if tc.expectedStatus == http.StatusOK {
				var responseBody struct {
					CampaignID int `json:"campaign_id"`
				}
				err = json.NewDecoder(resp.Body).Decode(&responseBody)
				assert.NoError(t, err)
				assert.NotEqual(t, 0, responseBody.CampaignID)
			}
		})
	}
}

func TestAdDecisionHandler(t *testing.T) {
	// Set up test server
	ts := httptest.NewServer(http.HandlerFunc(adDecisionHandler))
	defer ts.Close()

	// Add a sample campaign
	campaign := Campaign{
		ID:             1001,
		StartTimestamp: time.Now().Add(-time.Hour).Unix(),
		EndTimestamp:   time.Now().Add(time.Hour).Unix(),
		TargetKeywords: []string{"shampoo"},
		MaxImpression:  100,
		CPM:            10,
	}
	campaignMutex.Lock()
	campaigns = append(campaigns, campaign)
	campaignMutex.Unlock()

	// Test cases
	tests := []struct {
		name           string
		adRequest      AdRequest
		expectedStatus int
	}{
		{
			name: "Valid Ad Request",
			adRequest: AdRequest{
				Keywords: []string{"shampoo"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Ad Request - Missing Keywords",
			adRequest: AdRequest{
				Keywords: []string{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "No Matching Campaign",
			adRequest: AdRequest{
				Keywords: []string{"nonexistent"},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Encode adRequest struct as JSON
			jsonData, err := json.Marshal(tc.adRequest)
			assert.NoError(t, err)

			// Send POST request to the server
			resp, err := http.Post(fmt.Sprintf("%s/addecision", ts.URL), "application/json", bytes.NewReader(jsonData))
			assert.NoError(t, err)

			// Check response status code
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			// Check response body for valid ad requests
			if tc.expectedStatus == http.StatusOK && strings.Contains(tc.name, "Valid Ad Request") {
				var responseBody struct {
					CampaignID    int    `json:"campaign_id"`
					ImpressionURL string `json:"impression_url"`
				}
				err = json.NewDecoder(resp.Body).Decode(&responseBody)
				fmt.Println(err)
				assert.NoError(t, err)
				assert.NotEqual(t, 0, responseBody.CampaignID)
				assert.NotEmpty(t, responseBody.ImpressionURL)
			}
		})
	}
}

func TestImpressionHandler(t *testing.T) {
	// Set up test server
	ts := httptest.NewServer(http.HandlerFunc(impressionHandler))
	defer ts.Close()

	// Add a sample campaign with an impression ID
	impressionID := generateUUID()
	campaign := Campaign{
		ID:              1001,
		StartTimestamp:  time.Now().Add(-time.Hour).Unix(),
		EndTimestamp:    time.Now().Add(time.Hour).Unix(),
		TargetKeywords:  []string{"shoes"},
		MaxImpression:   100,
		CPM:             10,
		ImpressionCount: 0,
		ImpressionIds:   []string{impressionID},
	}
	campaignMutex.Lock()
	campaigns = append(campaigns, campaign)
	campaignMutex.Unlock()

	// Test cases
	tests := []struct {
		name           string
		impressionID   string
		expectedStatus int
	}{
		{
			name:           "Valid Impression",
			impressionID:   impressionID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Impression - Nonexistent ID",
			impressionID:   "nonexistent",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Impression - Empty ID",
			impressionID:   "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Method - POST",
			impressionID:   impressionID,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid Method - PUT",
			impressionID:   impressionID,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		// More test cases ...
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare request
			var req *http.Request
			var err error

			if tc.name == "Invalid Method - POST" {
				req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", ts.URL, tc.impressionID), nil)
			} else if tc.name == "Invalid Method - PUT" {
				req, err = http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s", ts.URL, tc.impressionID), nil)
			} else {
				req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", ts.URL, tc.impressionID), nil)
			}
			assert.NoError(t, err)

			// Send request to the server
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			// Check response status code
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
