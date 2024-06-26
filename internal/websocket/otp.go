package websocket

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ! NOT USED
type OTP struct {
	IP      string
	Key     string
	Created time.Time
}

type RetentionMap map[string]OTP

func NewRetentionMap(ctx context.Context, retentionPeriod time.Duration) RetentionMap {
	rm := make(RetentionMap)

	go rm.Retention(ctx, retentionPeriod)

	return rm
}

func (rm RetentionMap) NewOTP(ip string) (bool, OTP) {
	otp := OTP{}
	if rm.checkExistingIP(ip) == true {
		return false, otp
	}

	otp.IP = ip
	otp.Key = uuid.New().String()
	otp.Created = time.Now()

	rm[otp.Key] = otp

	return true, otp
}

func (rm RetentionMap) VerifyOTP(otp string) bool {
	if _, ok := rm[otp]; !ok {
		return false
	}
	delete(rm, otp)
	return true
}

func (rm RetentionMap) Retention(ctx context.Context, retentionPeriod time.Duration) {
	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-ticker.C:
			for _, otp := range rm {
				if otp.Created.Add(retentionPeriod).Before(time.Now()) {
					delete(rm, otp.Key)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (rm RetentionMap) checkExistingIP(ip string) bool {
	for _, otp := range rm {
		if otp.IP == ip {
			return true
		}
	}

	return false
}
