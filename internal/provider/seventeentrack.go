package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/channinghe/ems2sns/internal/model"
)

const apiBase = "https://api.17track.net/track/v2.2"

// SeventeenTrackProvider queries 17track API for China EMS events
type SeventeenTrackProvider struct {
	client *http.Client
	token  string
}

func NewSeventeenTrackProvider(token string) *SeventeenTrackProvider {
	return &SeventeenTrackProvider{
		client: &http.Client{Timeout: 30 * time.Second},
		token:  token,
	}
}

func (p *SeventeenTrackProvider) Name() string                { return "17track" }
func (p *SeventeenTrackProvider) Source() model.TrackingSource { return model.SourceChinaEMS }
func (p *SeventeenTrackProvider) NeedsRegistration() bool     { return true }

func (p *SeventeenTrackProvider) IsConfigured() bool {
	return p.token != ""
}

func (p *SeventeenTrackProvider) Register(ctx context.Context, trackingNumber string) error {
	if !p.IsConfigured() {
		return fmt.Errorf("17track API not configured")
	}

	payload := []map[string]string{{"number": trackingNumber}}
	body, err := p.doRequest(ctx, "/register", payload)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parsing register response: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("17track register error (code %d): %s", resp.Code, string(body))
	}
	if len(resp.Data.Rejected) > 0 {
		raw := string(resp.Data.Rejected[0])
		// -18010012 = already registered
		if !strings.Contains(raw, "-18010012") {
			return fmt.Errorf("17track rejected: %s", raw)
		}
	}
	return nil
}

func (p *SeventeenTrackProvider) FetchTrackingInfo(ctx context.Context, trackingNumber string) (*model.TrackingInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("17track API not configured")
	}

	payload := []map[string]string{{"number": trackingNumber}}
	body, err := p.doRequest(ctx, "/gettrackinfo", payload)
	if err != nil {
		return nil, fmt.Errorf("gettrackinfo: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("17track error (code %d): %s", resp.Code, string(body))
	}

	if len(resp.Data.Accepted) == 0 {
		if len(resp.Data.Rejected) > 0 {
			return nil, fmt.Errorf("17track rejected: %s", string(resp.Data.Rejected[0]))
		}
		return nil, fmt.Errorf("no tracking data returned")
	}

	var trackInfo trackInfoResult
	if err := json.Unmarshal(resp.Data.Accepted[0], &trackInfo); err != nil {
		return nil, fmt.Errorf("parsing tracking info: %w", err)
	}

	return p.toTrackingInfo(trackingNumber, &trackInfo), nil
}

func (p *SeventeenTrackProvider) toTrackingInfo(trackingNumber string, info *trackInfoResult) *model.TrackingInfo {
	result := &model.TrackingInfo{
		TrackingNumber: trackingNumber,
		Details:        []model.TrackingDetail{},
		LastUpdate:     time.Now(),
		Source:         model.SourceChinaEMS,
		Status:         info.TrackInfo.LatestStatus.Status,
	}

	for _, prov := range info.TrackInfo.Tracking.Providers {
		if !isChinaProvider(prov.Provider.Name) {
			continue
		}
		for _, ev := range prov.Events {
			location := ev.Location
			if location == "None" || location == "" {
				location = ""
			}
			result.Details = append(result.Details, model.TrackingDetail{
				DateTime:    formatEventTime(ev.TimeISO),
				Description: ev.Description,
				Office:      location,
				Source:      model.SourceChinaEMS,
			})
		}
	}

	reverseDetails(result.Details)
	return result
}

func (p *SeventeenTrackProvider) doRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+endpoint, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("17token", p.token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// --- API response types ---

type apiResponse struct {
	Code int         `json:"code"`
	Data apiRespData `json:"data"`
}

type apiRespData struct {
	Accepted []json.RawMessage `json:"accepted"`
	Rejected []json.RawMessage `json:"rejected"`
}

type trackInfoResult struct {
	Number    string           `json:"number"`
	Carrier   int              `json:"carrier"`
	TrackInfo trackInfoPayload `json:"track_info"`
}

type trackInfoPayload struct {
	LatestStatus struct {
		Status    string `json:"status"`
		SubStatus string `json:"sub_status"`
	} `json:"latest_status"`
	Tracking struct {
		Providers []trackProvider `json:"providers"`
	} `json:"tracking"`
}

type trackProvider struct {
	Provider struct {
		Key  int    `json:"key"`
		Name string `json:"name"`
	} `json:"provider"`
	Events []trackEvent `json:"events"`
}

type trackEvent struct {
	TimeISO     string `json:"time_iso"`
	TimeUTC     string `json:"time_utc"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Stage       string `json:"stage"`
}

// --- helpers ---

func isChinaProvider(name string) bool {
	return strings.Contains(strings.ToLower(name), "china")
}

func formatEventTime(isoTime string) string {
	if t, err := time.Parse(time.RFC3339, isoTime); err == nil {
		return t.Format("2006-01-02 15:04:05")
	}
	return isoTime
}

func reverseDetails(details []model.TrackingDetail) {
	for i, j := 0, len(details)-1; i < j; i, j = i+1, j-1 {
		details[i], details[j] = details[j], details[i]
	}
}
