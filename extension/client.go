package extension

import (
	"strconv"
	"strings"
	"time"

	"github.com/mileusna/useragent"
)

func generateClientCorrelationId(d Device, ua useragent.UserAgent) string {
	extId := d.ExtensionId
	if extId == "" {
		extId = GenerateExtensionId()
	}

	limit := 30
	if len(extId) > limit {
		extId = extId[:limit]
	}

	browserName := ua.Name
	if len(browserName) > 3 {
		browserName = browserName[:3]
	}

	return "EXT-" + strings.ToUpper(browserName) + "-" + extId + "-" + strconv.Itoa(int(time.Now().UnixMilli()))
}
