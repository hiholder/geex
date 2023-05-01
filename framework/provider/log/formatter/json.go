package formatter

import (
	"github.com/hiholder/geex/framework/contract"
	"github.com/json-iterator/go"
	gerrors "github.com/pkg/errors"
	"strings"
	"time"
)

func JsonFormatter(level contract.LogLevel, t time.Time, msg string, fields map[string]interface{}) ([]byte, error) {
	bs := strings.Builder{}
	fields["msg"] = msg
	fields["level"] = level
	fields["time"] = t.Format(time.RFC3339)
	marshal, err := jsoniter.Marshal(fields)
	if err != nil {
		return []byte{}, gerrors.Wrap(err, "json error")
	}
	bs.Write(marshal)
	return []byte(bs.String()), err
}