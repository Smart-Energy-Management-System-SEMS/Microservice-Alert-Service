package services

import "time"

func IsInactive(lastActive time.Time, now time.Time, maxMinutes int) bool {
    if maxMinutes <= 0 {
        return false
    }

    return now.Sub(lastActive) >= time.Duration(maxMinutes)*time.Minute
}
