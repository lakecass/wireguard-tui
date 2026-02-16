package ui

type Mascot struct {
	Frames []string
	Index  int
}

var (
	MascotSleep = []string{
		" ( -.-) Zzz ",
		" ( -.-) zZz ",
	}

	MascotActive = []string{
		" (/^▽^)/ ",
		" \\(^▽^\\) ",
		" (/^▽^)/ ",
		" \\(^▽^\\) ",
	}

	// A simple Gopher-ish ASCII
	GopherIdle = []string{
		`
   ,-.
   | |
   | |
  /| |\
 (_| |_)
`,
	}

	GopherRun = []string{
		`
    ,-.
    | |
   /| |
  (_| |_
`,
		`
   ,-.
   | |
   | |\
  _| |_)
`,
	}
)

func GetMascotFrame(tick int, active bool) string {
	if !active {
		// Sleep mode
		idx := (tick / 5) % len(MascotSleep)
		return MascotSleep[idx]
	}
	// Active mode
	idx := tick % len(MascotActive)
	return MascotActive[idx]
}

// Helper to overlay mascot on the view string?
// Or better, just render it in the View() function positioned correctly.
