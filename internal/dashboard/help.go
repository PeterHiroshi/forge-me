package dashboard

func (m Model) renderHelp() string {
	title := helpTitleStyle.Render("Keyboard Shortcuts")

	sections := "" +
		"Navigation\n" +
		"  Tab / Shift-Tab   Switch tabs\n" +
		"  1 / 2 / 3 / 4    Jump to tab\n" +
		"  j/k / up/down     Move cursor / scroll\n" +
		"\n" +
		"Actions\n" +
		"  r                 Refresh data\n" +
		"  q                 Quit\n" +
		"  ?                 Toggle help\n" +
		"\n" +
		"Interactive\n" +
		"  /                 Filter (workers/containers/alerts)\n" +
		"  Enter             Open detail view\n" +
		"  Esc               Back / dismiss\n"

	return helpOverlayStyle.Render(title + "\n" + sections)
}
