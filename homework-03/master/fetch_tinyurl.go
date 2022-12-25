package master

import (
    "fmt"
    "errors"
    "database/sql"
)

var ErrNotFound = errors.New("Not found")
var ErrInternal = errors.New("Internal server error")

func FetchLongURL(tinyurl string) (string, error) {
  var longurl string

  statement := "SELECT longurl FROM url_mappings WHERE tinyurl = $1;"
  if err := PGConnection.QueryRow(statement, tinyurl).Scan(&longurl); err != nil {
    if err == sql.ErrNoRows {
      return "", ErrNotFound
    } else {
      fmt.Println(err)
      return "", ErrInternal
    }
  }

  return longurl, nil
}
