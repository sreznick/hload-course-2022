package services

import (
    global "main/global"
    "crypto/sha256"
    "encoding/hex"
    base62 "github.com/jxskiss/base62"
    pq "github.com/lib/pq"
)

const TINYURL_LENGTH = 7

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

  for offset = 0; offset + TINYURL_LENGTH < len([]rune(tinyurl)); offset += 1 {
    if _, insert_err = global.Connection.Exec(statement, longurl, tinyurl[offset : offset + TINYURL_LENGTH]); insert_err == nil {
      break
    } else {
      if pq_err, ok := insert_err.(*pq.Error); ok {
        // Workaround for case when another request with the same longurl succeed.
        if pq_err.Constraint == "url_mappings_longurl_idx" {
          return FetchLongURL(longurl)
        }
      }
    }
  }

  if insert_err != nil {
    return "", ErrInternal
  }

  return tinyurl[offset : offset + TINYURL_LENGTH], nil
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
