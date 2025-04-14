package tests

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Targets struct {
	data   *gmap.StrStrMap
	strict bool
	fatal  bool
}

func NewTargets(strict bool, fatal bool) *Targets {
	return &Targets{
		data:   gmap.NewStrStrMap(true),
		strict: strict,
		fatal:  fatal,
	}
}

func (target *Targets) Add(value string, key string) *Targets {
	if !target.data.Contains(key) {
		target.data.Set(key, value)
	} else {
		if target.strict && target.data.Get(key) == value || !target.strict {
			if target.fatal {
				panic(fmt.Sprintf("%s 已存在\n", key))
			} else {
				fmt.Printf("%s 已存在\n", key)
			}
		}
	}
	return target
}

func (target *Targets) Map() map[string]string {
	return target.data.Map()
}

func (target *Targets) Diff(other *Targets) *Targets {
	data := target.data.Clone()
	data.Removes(other.data.Keys())
	return &Targets{
		data: data,
	}
}

func (target *Targets) LoadFromCSV(name string) (err error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, os.ModePerm)
	if err != nil {
		err = gerror.Wrapf(err, "打开 csv 失败")
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		err = gerror.Wrapf(err, "读取 csv 失败")
	}

	for _, datum := range data {
		target.Add(datum[0], datum[1])
	}

	return
}

func (target *Targets) MustLoadFromCSV(name string) *Targets {
	err := target.LoadFromCSV(name)
	if err != nil {
		panic(err)
	}

	return target
}
