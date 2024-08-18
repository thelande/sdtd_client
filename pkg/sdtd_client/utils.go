/*
Copyright © 2024 Tom Helander thomas.helander@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package sdtdclient

import "fmt"

func SecondsToDaysHoursMinutesSeconds(playtime int) string {
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
