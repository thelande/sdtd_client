package sdtdclient

import "fmt"

func secondsToDaysHoursMinutesSeconds(playtime int) string {
	var days, hours, minutes, seconds int
	seconds = playtime

	days = seconds / 86400
	seconds -= (days * 86400)

	hours = seconds / 3600
	seconds -= (hours * 3600)

	minutes = seconds / 60
	seconds -= (minutes * 60)

	return fmt.Sprintf("%d:%02d:%02d:%02d", days, hours, minutes, seconds)
}

// Returns the player's total playtime as a string in days, hours, minutes, and
// seconds.
func (p Player) GetPlaytime() string {
	return secondsToDaysHoursMinutesSeconds(p.TotalPlayTimeSeconds)
}

// Returns the player's total playtime as a string in days, hours, minutes, and
// seconds.
func (p PlayerM) GetPlaytime() string {
	return secondsToDaysHoursMinutesSeconds(p.TotalPlayTimeSeconds)
}
