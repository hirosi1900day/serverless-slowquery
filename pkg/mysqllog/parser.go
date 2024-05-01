package mysqllog

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/percona/go-mysql/query"
)

type Query struct {
	Time         string
	User         string
	Host         string
	ID           string
	QueryTime    string
	LockTime     string
	RowsSent     string
	RowsExamined string
	RowsAffected string
	Query        string
	Fingerprint  string
}

func Parser(r io.Reader) Query {
	scanner := bufio.NewScanner(r)
	var queries []string
	var q Query
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if parts[0] != "#" {
			queries = append(queries, line)
		} else {
			for i, part := range parts {
				part = strings.ToLower(part)
				if strings.Contains(part, "query_time:") {
					q.QueryTime = parts[i+1]
				} else if strings.Contains(part, "lock_time:") {
					q.LockTime = parts[i+1]
				} else if strings.Contains(part, "time:") {
					q.Time = parts[i+1]
				} else if strings.Contains(part, "user@host:") {
					items := regexp.MustCompile(`\[(.*?)\]`).FindAllString(line, -1)
					q.User = items[0]
					q.Host = items[1]
				} else if strings.Contains(part, "id:") {
					q.ID = parts[i+1]
				} else if strings.Contains(part, "rows_sent:") {
					q.RowsSent = parts[i+1]
				} else if strings.Contains(part, "rows_examined:") {
					q.RowsExamined = parts[i+1]
				} else if strings.Contains(part, "rows_affected:") {
					q.RowsAffected = parts[i+1]
				}
			}
		}
	}
	q.Query = strings.Join(queries, " ")

	// NOTE: クエリを Fingerprint でマスクする
	var fingerprints []string
	t := strings.Split(q.Query, ";")
	for _, qq := range t {
		// use <database> の <database> をマスクしてしまうので除外
		// SET timestamp=<timestamp> の <timestamp> をマスクしてしまうので除外
		if !strings.HasPrefix(qq, "use ") && !strings.HasPrefix(qq, "SET") {
			qq = query.Fingerprint(qq)
		}
		fingerprints = append(fingerprints, qq)
	}
	q.Fingerprint = strings.Join(fingerprints, ";\n")
	return q
}
