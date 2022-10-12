package services

import (
    "fmt"
    global "main/global"
    "crypto/sha256"
    "encoding/hex"
    base62 "github.com/jxskiss/base62"
    pq "github.com/lib/pq"
)

func CreateTinyURL(longurl string) (string, error) {
  if _, err := FetchLongURL(longurl); err != nil {
    if err != ErrNotFound {
      return "", err
    }
  }

  statement := "INSERT INTO url_mappings(longurl, tinyurl) VALUES ($1, $2);"
  tinyurl := transformLongURL(longurl)

  var insert_err error
  var offset = 0

  for offset = 0; offset + 7 < len([]rune(tinyurl)); offset += 1 {
    if _, insert_err = global.Connection.Exec(statement, longurl, tinyurl[offset : offset + 7]); insert_err == nil {
      break
    } else {
      if pq_err, ok := insert_err.(*pq.Error); ok {
        // Workaround for case when another request with the same longurl succeed after FetchLongUrl was
        // invoked and failed in this method.
        // In general, we prefer [C]onsistency over [A]vailabillity.
        if pq_err.Constraint == "url_mappings_longurl_idx" {
          return FetchLongURL(longurl)
        }
      }
    }
  }

  if insert_err != nil {
    fmt.Println("ERROR", tinyurl[0:7], longurl)
    return "", ErrInternal
  }

  return tinyurl[offset : offset + 7], nil
}

func transformLongURL(longurl string) string {
  return toBase62(getSHA256Hash(longurl))
}

func getSHA256Hash(text string) string {
   hash := sha256.Sum256([]byte(text))

   return hex.EncodeToString(hash[:])
}

func toBase62(text string) string {
  return base62.EncodeToString([]byte(text))
}
