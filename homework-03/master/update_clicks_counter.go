package master

import (
  "fmt"
  "strconv"
  "log"
  "context"
  "database/sql"
)

const RETRY_COUNT = 5

func UpdateClicksCounter(tinyurl string, delta int) {
  // pg-specific errors are possible so we ret for RETRY_COUNT times
  var err error
  for i := 0; i < RETRY_COUNT; i++ {
    err = UpdateClicksCounterImpl(tinyurl, delta)
    if err == nil {
      return
    } else {
      fmt.Println("[master] pgretry " + strconv.Itoa(i))
    }
  }

  fmt.Println("[master] failed to update counter, error %s", err)
}

func UpdateClicksCounterImpl(tinyurl string, delta int) error {
  statement := "SELECT clicks_count FROM counters WHERE tinyurl = $1"
  var ctx = context.Background()

  var counter int64

  tx, err := PGConnection.BeginTx(ctx, nil); if err != nil {
    return err
  }

  // within transaction check if record exists and create if it doesn't
  if err := tx.QueryRow(statement, tinyurl).Scan(&counter); err != nil {
    if err != sql.ErrNoRows {
      tx.Rollback()
      return err
    }

    statement = "INSERT INTO counters(tinyurl, clicks_count) VALUES ($1, $2)"
    if _, e := tx.ExecContext(ctx, statement, tinyurl, delta); e != nil {
      tx.Rollback()
      return e
    }

    err = tx.Commit()
    if err != nil {
      log.Fatal(err)
      return err
    }

    return nil
  }

  statement = "UPDATE counters SET clicks_count = clicks_count + $1 WHERE tinyurl = $2;"
  if _, e := tx.ExecContext(ctx, statement, delta, tinyurl); e != nil {
    tx.Rollback()
    return e
  }

  err = tx.Commit()
	if err != nil {
		log.Fatal(err)
    return err
	}

  return nil
}
