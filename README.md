## etcdcfg

可以方便地从etcd里面读取配置

- 支持方便的Reload，监控配置变化的时候自动调用Reload函数
 - 只需要实现 ReloadXXX
 - 还支持子结构体精细化的Reload
- 支持完全的json到etcd的映射
- 支持心跳（etcd的配置加TTL


如果配置不是写在本地的配置文件里面，那么所有的升级只需要改动一次配置，就可以做到程序之间是统一的。
因为需要灰度，那么一个程序会有两套的配置，一个作为灰度配置，一个作为正式配置。


### 使用

```
type A struct {
    ABC int `json:"abc"`
}

// 只要配置文件里面含有A的结构体，当配置ABC变化的时候会自动调用这个函数
func (a *A) ReloadABC(i int) {
    println("in ReloadABC:", i)
    a.ABC = i
}

type Config struct {
    KeyNil *string `json:"key_nil"`
    Key0   *string `json:"key0"`
    Key1   string  `json:"key1"`
    Struct *A           `json:"struct"`

    Slice       []int               `json:"slice"`
    SliceMap    []map[string]string `json:"slice_map"`
    SliceStruct []A                     `json:"slice_struct"`
    SliceSlice  [][]string          `json:"slice_slice"`

    Map       map[string]string            `json:"map"`
    MapMap    map[string]map[string]string `json:"map_map"`
    MapStruct map[string]A                 `json:"map_struct"`
}

// Key1的Reload函数
func (conf *Config) ReloadKey1(s string) {
    println("in ReloadKey1:", s)
    conf.Key1 = s
}

```

PS. 这才完成一半啊不要吐槽文档，单元测试什么的

