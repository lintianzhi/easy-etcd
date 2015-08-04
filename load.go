package etcdcfg

import (
	"flag"
	"log"
	"path"
	"reflect"
	"strings"

	"github.com/coreos/go-etcd/etcd"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/misc/strconv"
)

var (
	confName *string
)

func Init(cFlag, default_conf string) {

	confName = flag.String(cFlag, default_conf, "the etcd config file")
}

func Load(conf interface{}) error {

	if !flag.Parsed() {
		flag.Parse()
	}

	log.Println("use the etcd file of:", *confName)
	client, err := etcd.NewClientFromFile(*confName)
	if err != nil {
		return err
	}
	return LoadWithEtcdClient(client, "appName", conf)
}

var (
	defaultTag = "json"
)

func LoadWithEtcdClient(client *etcd.Client, app string, conf interface{}) (err error) {

	v := reflect.ValueOf(conf)
	if v.Kind() != reflect.Ptr {
		err = errors.New("etcdcfg.LoadWithEtcdClient: ret.type != pointer")
		return
	}

	if v.Elem().Kind() != reflect.Struct {
		err = errors.New("etcdcfg.LoadWithEtcdClient: ret.type != struct")
		return
	}

	resp, err := client.Get(app, true, true)
	if err != nil {
		err = errors.Info(err, "etcd.Get").Detail(err)
		return
	}

	err = parseNode(client, "/"+app+"/", &v, resp.Node)
	if err != nil {
	}

	return
}

func parseNode(client *etcd.Client, name string, v *reflect.Value, root *etcd.Node) (err error) {

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			nv := reflect.New(v.Type().Elem())
			v.Set(nv)
		}
		*v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		if !root.Dir {
			err = errors.New("node of " + name + "is not dir")
			return
		}
		vt := v.Type()
		for i := 0; i < v.NumField(); i++ {
			sv := v.Field(i)
			sf := vt.Field(i)
			tag := sf.Tag.Get(defaultTag)
			if tag == "" {
				// do not parse value without tag
				continue
			}
			tagv, ok := parseTag(tag)
			if !ok {
				log.Println("tag is empty:", sf.Name)
				continue
			}
			tagv = path.Join(name, tagv)
			for _, node := range root.Nodes {
				if node.Key == tagv {
					parseNode(client, tagv, &sv, node)
					break
				}
			}
			if client != nil && v.CanAddr() {
				vPtr := v.Addr()
				reloadMethodName := "Reload" + sf.Name
				reloadMethod := vPtr.MethodByName(reloadMethodName)
				if reloadMethod.IsValid() {
					previousValue := sv
					// check reloadMethod's arguments
					receiver := make(chan *etcd.Response)
					go func() {
						_, err := client.Watch(tagv, 0, true, receiver, nil)
						if err != nil {
							// deliver err to outside
							panic(err)
						}
					}()

					go func() {
						for {
							resp := <-receiver
							rv := reflect.New(sf.Type)
							parseNode(nil, tagv, &rv, resp.Node)
							if reflect.DeepEqual(previousValue.Interface(), rv.Interface()) {
								continue
							}
							previousValue = rv
							reloadMethod.Call([]reflect.Value{rv})
						}
					}()
				}
			}
		}
	case reflect.Slice:
		if !root.Dir {
			err = errors.New("node of " + name + "is not dir")
			return
		}
		n := len(root.Nodes)
		vt := v.Type()
		slice := reflect.MakeSlice(vt, n, n)
		for i := 0; i < n; i++ {
			sv := slice.Index(i)
			err = parseNode(client, root.Nodes[i].Key, &sv, root.Nodes[i])
			if err != nil {
				return
			}
		}
		v.Set(slice)
	case reflect.Map:
		if !root.Dir {
			err = errors.New("node of " + name + "is not dir")
			return
		}
		vt := v.Type()
		keyT, valueT := vt.Key(), vt.Elem()
		mp := reflect.MakeMap(vt)
		for _, node := range root.Nodes {
			key := reflect.New(keyT)
			key = key.Elem()
			err = strconv.ParseValue(key, node.Key[len(name)+1:])
			if err != nil {
				return
			}
			var value reflect.Value
			value = reflect.New(valueT)
			value = value.Elem()
			err = parseNode(client, node.Key, &value, node)
			if err != nil {
				return
			}
			mp.SetMapIndex(key, value)
		}
		v.Set(mp)
	default:
		err = strconv.ParseValue(*v, root.Value)
	}

	return
}

func parseTag(tag string) (v string, ok bool) {

	parts := strings.Split(tag, ",")
	v = parts[0]
	if v != "" {
		ok = true
	}

	// todo: add some opt

	return
}
