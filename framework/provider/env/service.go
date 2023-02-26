package env

import (
	"bufio"
	"bytes"
	"github.com/hiholder/geex/framework/contract"
	gerrors "github.com/pkg/errors"
	"io"
	"os"
	"path"
	"strings"
)

type GeexEnv struct {
	folder string
	maps map[string]string
}

func (g *GeexEnv) AppEnv() string {
	return g.maps["APP_ENV"]
}

func (g *GeexEnv) IsExist(s string) bool {
	_, ok := g.maps[s]
	return ok
}

func (g *GeexEnv) Get(s string) string {
	if val, ok := g.maps[s]; ok {
		return val
	}
	return ""
}

func (g *GeexEnv) All() map[string]string {
	return g.maps
}

func NewGeexEnv(params ...interface{}) (interface{}, error) {
	folder, ok := params[0].(string)
	if !ok {
		return nil, gerrors.New("GeexEnv params error")
	}
	geexEnv := &GeexEnv{
		folder: folder,
		maps: make(map[string]string),
	}
	geexEnv.maps["APP_ENV"] = contract.EnvDevelopment
	file := path.Join(folder, ".env")
	fi, err := os.Open(file)
	if err == nil {
		defer fi.Close()
		br := bufio.NewReader(fi)
		for {
			line, _, err := br.ReadLine()
			if err == io.EOF {
				break
			}
			env := bytes.SplitN(line, []byte{'='}, 2)
			if len(env) < 2 {
				continue
			}
			geexEnv.maps[string(env[0])] = string(env[1])
		}
	}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) < 2 {
			continue
		}
		geexEnv.maps[pair[0]] = pair[1]
	}

	return geexEnv, nil
}
