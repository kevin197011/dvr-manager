package audit

import (
	"log"
	"os"
	"strconv"
	"time"
)

const DefaultRetentionMonths = 3

// RetentionMonths 审计日志保留月数，默认 3；可通过 AUDIT_RETENTION_MONTHS 配置。
func RetentionMonths() int {
	if s := os.Getenv("AUDIT_RETENTION_MONTHS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
		log.Printf("[Audit] invalid AUDIT_RETENTION_MONTHS=%q, using default %d", s, DefaultRetentionMonths)
	}
	return DefaultRetentionMonths
}

// RetentionCutoff 早于该时间的审计记录应被硬删除。
func RetentionCutoff() time.Time {
	return time.Now().AddDate(0, -RetentionMonths(), 0)
}
