package sdtdclient

import "fmt"

// Returns a location as a string coordinate.
func (l Location) GetCoordinates() string {
	return fmt.Sprintf("(%d, %d, %d)", l.X, l.Y, l.Z)
}
