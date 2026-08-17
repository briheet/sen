package trace

import "testing"

func statementRows(n int) []Statement {
	rows := make([]Statement, 0, n)
	for i := 0; i < n; i++ {
		rows = append(rows, Statement{
			QueryID:   int64(1000 + i),
			Query:     "SELECT * FROM users WHERE id = $1 AND status IN ($2, $3) ORDER BY id",
			Calls:     int64(50 + i),
			TotalExec: float64(20 + i),
			Rows:      int64(10 + i),
		})
	}
	return rows
}

func tableRows(n int) []Table {
	rows := make([]Table, 0, n)
	for i := 0; i < n; i++ {
		rows = append(rows, Table{Name: "table_" + string(rune('a'+i%26)), SeqScan: int64(i), LiveTuples: int64(1000 + i)})
	}
	return rows
}

func BenchmarkStatements(b *testing.B) {
	rows := statementRows(100)
	b.ReportAllocs()
	for b.Loop() {
		p := Statements(rows)
		_ = p
	}
}

func BenchmarkTables(b *testing.B) {
	rows := tableRows(100)
	b.ReportAllocs()
	for b.Loop() {
		p := Tables(rows)
		_ = p
	}
}

func BenchmarkLabel(b *testing.B) {
	q := "SELECT * FROM very_long_table_name WHERE some_column = $1 AND another_column = $2"
	b.ReportAllocs()
	for b.Loop() {
		_ = label(q)
	}
}
