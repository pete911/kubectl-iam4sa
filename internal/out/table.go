package out

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

type Table struct {
	writer *tabwriter.Writer
}

func NewTable() Table {
	return Table{writer: tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)}
}

func (t Table) AddRow(columns ...string) error {
	_, err := fmt.Fprintln(t.writer, strings.Join(columns, "\t"))
	return err
}

func (t Table) Print() error {
	return t.writer.Flush()
}
