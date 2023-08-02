module main

go 1.19

require (
	github.com/integrii/flaggy v1.5.2
	github.com/tadhunt/jenkinstool v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/text v0.11.0 // indirect
	gopkg.in/vansante/go-dl-stream.v2 v2.0.1 // indirect
)

replace github.com/tadhunt/jenkinstool => ./..
