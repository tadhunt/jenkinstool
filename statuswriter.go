package jenkinstool

import (
	"golang.org/x/text/message"
	"golang.org/x/text/number"
	"os"
	"time"
)

const (
	Esc       = "\u001B["
	EraseLine = Esc + "2K"
	SOL       = "\r"
)

type StatusWriter struct {
	p      *message.Printer
	format number.FormatFunc
	last   int64
	total  int64
	start  time.Time
	name   string
	quiet  bool
}

func (sw *StatusWriter) Write(data []byte) (int, error) {
	sw.total += int64(len(data))

	if !sw.quiet {
		if sw.total-sw.last >= 256*1000 {
			kb := float64(sw.total) / 1000.0
			elapsed := time.Now().Sub(sw.start)
			kbps := kb / elapsed.Seconds()
			sw.p.Fprintf(os.Stdout, "%s%sDownloading %s %v KB (%v KB/s)", EraseLine, SOL, sw.name, sw.format(kb), sw.format(kbps))
			os.Stdout.Sync()
			sw.last = sw.total
		}
	}

	return len(data), nil
}
