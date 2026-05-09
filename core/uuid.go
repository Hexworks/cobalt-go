package core

import "github.com/google/uuid"

// UUID is a 128-bit unique identifier. Aliased to google/uuid so the
// canonical Go UUID type flows through the entire library without
// wrapping overhead.
type UUID = uuid.UUID

// NilUUID is the all-zeros UUID.
var NilUUID = uuid.Nil

// RandomUUID returns a freshly generated v4 UUID.
func RandomUUID() UUID {
	return uuid.New()
}

// UUIDFromString parses a UUID from its canonical 36-character
// representation. Returns an error if the input is malformed.
func UUIDFromString(s string) (UUID, error) {
	return uuid.Parse(s)
}

// Abbreviate returns the first four characters of u's string form. The
// Kotlin original used this for compact log/debug output.
func Abbreviate(u UUID) string {
	return u.String()[:4]
}
