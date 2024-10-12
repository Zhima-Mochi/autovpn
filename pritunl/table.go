package pritunl

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

var (
	table *tablewriter.Table
)

func createTable() *tablewriter.Table {
	if table != nil {
		return table
	}
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Server", "User", "Status", "Connected", "Client IP", "Server IP"})
	table.SetColWidth(1000)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(false)
	return table
}
