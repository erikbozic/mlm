package views

import (
	"github.com/rivo/tview"
)

type LogsView struct {
	*tview.TextView
}

func NewLogsView() *LogsView {

	v := &LogsView{
		TextView: tview.NewTextView().SetWordWrap(true).SetWrap(true).
			SetRegions(false).SetDynamicColors(true),
	}
	v.SetTitle("logs")
	v.SetBorder(true)
	return v
}


