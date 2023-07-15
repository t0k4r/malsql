package qb

import (
	"fmt"
	"strings"
)

type pair struct {
	col string
	val string
}

type query struct {
	table string
	pairs []pair
}

func Insert(table string) *query {
	return &query{
		table: table,
	}
}

func (q *query) Int(col string, val int) *query {
	q.pairs = append(q.pairs, pair{
		col: col,
		val: fmt.Sprint(val),
	})
	return q
}
func (q *query) Str(col string, val string) *query {
	if val != "" {
		q.pairs = append(q.pairs, pair{
			col: col,
			val: `'` + strings.ReplaceAll(val, "'", "''") + `'`,
		})
	}

	return q
}

func (q *query) SubQ(col string, query string, param string) *query {
	if param != "" {
		subQ := fmt.Sprintf(query, strings.ReplaceAll(param, "'", "''"))
		q.pairs = append(q.pairs, pair{
			col: col,
			val: `(` + subQ + `)`,
		})
	}
	return q
}
func (q *query) SubQRaw(col string, sql string) *query {
	q.pairs = append(q.pairs, pair{
		col: col,
		val: `(` + sql + `)`,
	})
	return q
}

func (q *query) Sql() string {
	if len(q.pairs) == 0 {
		return ""
	}
	insertInto := `INSERT INTO ` + q.table + ` (`
	values := `VALUES (`
	for _, p := range q.pairs {
		insertInto += p.col + `, `
		values += p.val + `, `
	}
	insertInto = insertInto[:len(insertInto)-2] + `) `
	values = values[:len(values)-2] + `) `
	return insertInto + values + `ON CONFLICT DO NOTHING`
}
