package mysqllog

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/percona/go-mysql/query"
)

// Query 構造体の定義は変更なし
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

// メタデータの解析を別関数に
func parseMetadata(line string, q *Query) {
	parts := strings.Fields(line)
	for i, part := range parts {
		part = strings.ToLower(part)
		switch {
		case strings.Contains(part, "query_time:"):
			q.QueryTime = parts[i+1]
		case strings.Contains(part, "lock_time:"):
			q.LockTime = parts[i+1]
		case strings.Contains(part, "time:"):
			q.Time = parts[i+1]
		case strings.Contains(part, "user@host:"):
			items := regexp.MustCompile(`\[(.*?)\]`).FindAllString(line, -1)
			if len(items) >= 2 {
				q.User = items[0]
				q.Host = items[1]
			}
		case strings.Contains(part, "id:"):
			q.ID = parts[i+1]
		case strings.Contains(part, "rows_sent:"):
			q.RowsSent = parts[i+1]
		case strings.Contains(part, "rows_examined:"):
			q.RowsExamined = parts[i+1]
		case strings.Contains(part, "rows_affected:"):
			q.RowsAffected = parts[i+1]
		}
	}
}

// クエリの正規化を別関数に
func normalizeQuery(queryText string) string {
	var fingerprints []string
	queries := strings.Split(queryText, ";")
	for _, qq := range queries {
		if !strings.HasPrefix(qq, "use ") && !strings.HasPrefix(qq, "SET") {
			qq = query.Fingerprint(qq)
		}
		fingerprints = append(fingerprints, qq)
	}
	return strings.Join(fingerprints, ";\n")
}

// Parser 関数のリファクタリング
func Parser(r io.Reader) Query {
	scanner := bufio.NewScanner(r)
	var queryText strings.Builder
	var q Query

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			parseMetadata(line, &q)
		} else {
			queryText.WriteString(line + " ")
		}
	}

	q.Query = queryText.String()
	q.Fingerprint = normalizeQuery(q.Query)

	return q
}
