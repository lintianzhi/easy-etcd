// +build ignore

package main

import (
	"fmt"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/lintianzhi/easy-etcd"
)

type A struct {
	A int `json:"a"`
}

func (a *A) ReloadA(i int) {
	println("in ReloadA:", i)
	a.A = i
}

type Config struct {
	KeyNil *string `json:"key_nil"`
	Key0   *string `json:"key0"`
	Key1   string  `json:"key1"`
	Struct *A      `json:"struct"`

	Slice       []int               `json:"slice"`
	SliceMap    []map[string]string `json:"slice_map"`
	SliceStruct []A                 `json:"slice_struct"`
	SliceSlice  [][]string          `json:"slice_slice"`

	Map       map[string]string            `json:"map"`
	MapMap    map[string]map[string]string `json:"map_map"`
	MapStruct map[string]A                 `json:"map_struct"`
}

func (conf *Config) ReloadKey1(s string) {
	println("in ReloadKey1:", s)
	conf.Key1 = s
}
func (conf *Config) ReloadKey0(s *string) {
	println("in ReloadKey0:", *s)
	conf.Key0 = s
}

func (conf *Config) ReloadStruct(a *A) {
	fmt.Println(a)
}

func main() {

	client := etcd.NewClient([]string{})
	var conf Config
	err := etcdcfg.LoadWithEtcdClient(client, "test", &conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	//	for {
	fmt.Printf("conf: %#v %v\n", conf, conf.Struct.A)
	fmt.Println("Key1:", conf.Key1)
	fmt.Println("Key0:", *conf.Key0)
	time.Sleep(100 * time.Second)
	//	}
}
