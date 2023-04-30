package formatter

import (
	"fmt"
	"github.com/hiholder/geex/framework/contract"
	"strings"
	"time"
)

func TextFormatter(level contract.LogLevel, t time.Time, msg string, fields map[string]interface{}) ([]byte, error) {
	var bs  strings.Builder
	sep := "\t"
	prefix := Prefix(level)
	bs.WriteString(prefix)
	bs.WriteString(sep)
	// 打印时间
	bs.WriteString(t.Format(time.RFC3339))
	bs.WriteString(sep)

	bs.WriteString("\"")
	bs.WriteString(msg)
	bs.WriteString("\"")
	bs.WriteString(sep)
	bs.WriteString(fmt.Sprint(fields))
	return []byte(bs.String()), nil
}
