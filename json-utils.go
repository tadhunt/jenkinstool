package jenkinstool

import (
	"strconv"
	"strings"
	"time"
)

//
// From https://gist.github.com/alexmcroberts/219127816e7a16c7bd70
//

type JsonTime time.Time

func (t JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

func (t *JsonTime) UnmarshalJSON(s []byte) (err error) {
	r := strings.Replace(string(s), `"`, ``, -1)

	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q/1000, 0)
	return
}

func (t JsonTime) String() string {
	return time.Time(t).String()
}
