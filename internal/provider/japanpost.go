package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/channinghe/ems2sns/internal/model"
)

const japanPostBaseURL = "https://trackings.post.japanpost.jp/services/srv/search/direct"

// JapanPostProvider scrapes Japan Post tracking page
type JapanPostProvider struct {
	client *http.Client
	locale string
	source model.TrackingSource
}

func NewJapanPostProvider(locale string, source model.TrackingSource) *JapanPostProvider {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &JapanPostProvider{
		client: &http.Client{Transport: tr, Timeout: 30 * time.Second},
		locale: locale,
		source: source,
	}
}

func (p *JapanPostProvider) Name() string {
	return fmt.Sprintf("JapanPost(%s)", p.locale)
}

func (p *JapanPostProvider) Source() model.TrackingSource {
	return p.source
}

func (p *JapanPostProvider) NeedsRegistration() bool { return false }

func (p *JapanPostProvider) Register(_ context.Context, _ string) error { return nil }

func (p *JapanPostProvider) FetchTrackingInfo(ctx context.Context, trackingNumber string) (*model.TrackingInfo, error) {
	params := url.Values{}
	params.Set("searchKind", "S004")
	params.Set("locale", p.locale)
	params.Set("reqCodeNo1", trackingNumber)
	params.Set("x", "0")
	params.Set("y", "0")

	fullURL := fmt.Sprintf("%s?%s", japanPostBaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ja,en-US;q=0.7,en;q=0.3")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching tracking page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return p.parseHTML(trackingNumber, string(body))
}

func (p *JapanPostProvider) parseHTML(trackingNumber, htmlContent string) (*model.TrackingInfo, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	info := &model.TrackingInfo{
		TrackingNumber: trackingNumber,
		Details:        []model.TrackingDetail{},
		LastUpdate:     time.Now(),
		Source:         p.source,
	}

	errorMsg := doc.Find(".error").Text()
	if errorMsg != "" {
		return nil, fmt.Errorf("tracking error: %s", strings.TrimSpace(errorMsg))
	}

	phonePattern := regexp.MustCompile(`^0\d{1,4}-\d{2,4}-\d{2,4}$`)
	dateTimePattern := regexp.MustCompile(`^\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}$`)
	dateOnlyPattern := regexp.MustCompile(`^\d{4}/\d{2}/\d{2}$`)
	trackingNumPattern := regexp.MustCompile(`^[A-Z]{2}\s+\d{3}\s+\d{3}\s+\d{3}\s+[A-Z]{2}$`)
	postalCodePattern := regexp.MustCompile(`^\d{3}-\d{4}$`)

	doc.Find("table.tableType01 tr, table.tableType02 tr, table tr").Each(func(_ int, row *goquery.Selection) {
		if row.Find("th").Length() > 0 {
			return
		}

		cells := row.Find("td")
		if cells.Length() < 2 {
			return
		}

		dateTime := strings.TrimSpace(cells.Eq(0).Text())
		description := strings.TrimSpace(cells.Eq(1).Text())

		if phonePattern.MatchString(dateTime) || phonePattern.MatchString(description) {
			return
		}
		if trackingNumPattern.MatchString(dateTime) || postalCodePattern.MatchString(dateTime) {
			return
		}
		if !dateTimePattern.MatchString(dateTime) && !dateOnlyPattern.MatchString(dateTime) {
			return
		}
		if description == "" {
			return
		}

		detail := model.TrackingDetail{
			DateTime:    dateTime,
			Description: description,
			Source:      p.source,
		}

		if cells.Length() >= 3 {
			detail.Details = strings.TrimSpace(cells.Eq(2).Text())
		}
		if cells.Length() >= 4 {
			detail.Office = strings.TrimSpace(cells.Eq(3).Text())
		}
		if cells.Length() >= 5 {
			detail.Region = strings.TrimSpace(cells.Eq(4).Text())
		}

		info.Details = append(info.Details, detail)
	})

	statusText := doc.Find(".status, .statusText, h2, h3").First().Text()
	info.Status = strings.TrimSpace(statusText)

	if len(info.Details) == 0 {
		doc.Find("div, p, span").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if len(text) > 10 && len(text) < 500 {
				if containsDeliveryKeyword(text) {
					info.Status = text
				}
			}
		})
	}

	return info, nil
}

func containsDeliveryKeyword(text string) bool {
	keywords := []string{
		"配達", "受付", "通過", "到着",
		"Delivered", "Posting", "In transit", "Arrival",
	}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
