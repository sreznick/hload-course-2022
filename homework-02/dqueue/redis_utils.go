package dqueue

import (
  "strings"
  "regexp"
)

// seems like redis client does not provide easy way to check if operation error
// is a MOVED error thus this ugly hack is used as a workaround
func isMovedError(error string) bool {
  return strings.Contains(error, "MOVED")
}

// yet another workaround
func extractRedisHostFromMovedError(value string) string {
  // stolen from stackoverflow
  ipAddrPartRegex := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
  regexPattern := ipAddrPartRegex + "\\." + ipAddrPartRegex + "\\." + ipAddrPartRegex + "\\." + ipAddrPartRegex + ":" + "[0-9]+"
  regEx := regexp.MustCompile(regexPattern)

  return regEx.FindString(value)
}
