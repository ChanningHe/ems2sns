package tracker

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/channinghe/ems2sns/internal/model"
)

func computeHash(info *model.TrackingInfo) string {
	data := fmt.Sprintf("%s|%s|%d", info.TrackingNumber, info.Status, len(info.Details))
	for _, d := range info.Details {
		data += fmt.Sprintf("|%s|%s|%s|%s|%s", d.DateTime, d.Description, d.Details, d.Office, d.Region)
	}
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func isDelivered(info *model.TrackingInfo) bool {
	keywords := []string{
		"お届け済み", "配達完了", "delivered", "Delivered",
		"已签收", "已妥投",
	}

	for _, kw := range keywords {
		if strings.Contains(info.Status, kw) {
			return true
		}
	}
	for _, d := range info.Details {
		for _, kw := range keywords {
			if strings.Contains(d.Description, kw) {
				return true
			}
		}
	}
	return false
}
