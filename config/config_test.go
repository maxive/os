package config

import (
	"testing"

	"github.com/maxive/os/config/cmdline"
	"github.com/maxive/os/pkg/util"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/stretchr/testify/require"
)

func TestFilterKey(t *testing.T) {
	assert := require.New(t)
	data := map[interface{}]interface{}{
		"ssh_authorized_keys": []string{"pubk1", "pubk2"},
		"hostname":            "ros-test",
		"maxive": map[interface{}]interface{}{
			"ssh": map[interface{}]interface{}{
				"keys": map[interface{}]interface{}{
					"dsa":     "dsa-test1",
					"dsa-pub": "dsa-test2",
				},
			},
			"docker": map[interface{}]interface{}{
				"ca_key":  "ca_key-test3",
				"ca_cert": "ca_cert-test4",
				"args":    []string{"args_test5"},
			},
		},
	}
	expectedFiltered := map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"ssh": map[interface{}]interface{}{
				"keys": map[interface{}]interface{}{
					"dsa":     "dsa-test1",
					"dsa-pub": "dsa-test2",
				},
			},
		},
	}
	expectedRest := map[interface{}]interface{}{
		"ssh_authorized_keys": []string{"pubk1", "pubk2"},
		"hostname":            "ros-test",
		"maxive": map[interface{}]interface{}{
			"docker": map[interface{}]interface{}{
				"ca_key":  "ca_key-test3",
				"ca_cert": "ca_cert-test4",
				"args":    []string{"args_test5"},
			},
		},
	}
	filtered, rest := filterKey(data, []string{"maxive", "ssh"})
	assert.Equal(expectedFiltered, filtered)
	assert.Equal(expectedRest, rest)
}

func TestUnmarshalOrReturnString(t *testing.T) {
	assert := require.New(t)

	assert.Equal("ab", cmdline.UnmarshalOrReturnString("ab"))
	assert.Equal("a\nb", cmdline.UnmarshalOrReturnString("a\nb"))
	assert.Equal("a\n", cmdline.UnmarshalOrReturnString("a\n"))
	assert.Equal("\nb", cmdline.UnmarshalOrReturnString("\nb"))
	assert.Equal("a,b", cmdline.UnmarshalOrReturnString("a,b"))
	assert.Equal("a,", cmdline.UnmarshalOrReturnString("a,"))
	assert.Equal(",b", cmdline.UnmarshalOrReturnString(",b"))

	assert.Equal(int64(10), cmdline.UnmarshalOrReturnString("10"))
	assert.Equal(true, cmdline.UnmarshalOrReturnString("true"))
	assert.Equal(false, cmdline.UnmarshalOrReturnString("false"))

	assert.Equal([]interface{}{"a"}, cmdline.UnmarshalOrReturnString("[a]"))
	assert.Equal([]interface{}{"a"}, cmdline.UnmarshalOrReturnString("[\"a\"]"))

	assert.Equal([]interface{}{"a,"}, cmdline.UnmarshalOrReturnString("[\"a,\"]"))
	assert.Equal([]interface{}{" a, "}, cmdline.UnmarshalOrReturnString("[\" a, \"]"))
	assert.Equal([]interface{}{",a"}, cmdline.UnmarshalOrReturnString("[\",a\"]"))
	assert.Equal([]interface{}{" ,a "}, cmdline.UnmarshalOrReturnString("[\" ,a \"]"))

	assert.Equal([]interface{}{"a\n"}, cmdline.UnmarshalOrReturnString("[\"a\n\"]"))
	assert.Equal([]interface{}{" a\n "}, cmdline.UnmarshalOrReturnString("[\" a\n \"]"))
	assert.Equal([]interface{}{"\na"}, cmdline.UnmarshalOrReturnString("[\"\na\"]"))
	assert.Equal([]interface{}{" \na "}, cmdline.UnmarshalOrReturnString("[\" \na \"]"))

	assert.Equal([]interface{}{"a", "b"}, cmdline.UnmarshalOrReturnString("[a,b]"))
	assert.Equal([]interface{}{"a", "b"}, cmdline.UnmarshalOrReturnString("[\"a\",\"b\"]"))

	assert.Equal([]interface{}{"a,", "b"}, cmdline.UnmarshalOrReturnString("[\"a,\",b]"))
	assert.Equal([]interface{}{"a", ",b"}, cmdline.UnmarshalOrReturnString("[a,\",b\"]"))
	assert.Equal([]interface{}{" a, ", " ,b "}, cmdline.UnmarshalOrReturnString("[\" a, \",\" ,b \"]"))

	assert.Equal([]interface{}{"a\n", "b"}, cmdline.UnmarshalOrReturnString("[\"a\n\",b]"))
	assert.Equal([]interface{}{"a", "\nb"}, cmdline.UnmarshalOrReturnString("[a,\"\nb\"]"))
	assert.Equal([]interface{}{" a\n ", " \nb "}, cmdline.UnmarshalOrReturnString("[\" a\n \",\" \nb \"]"))

	assert.Equal([]interface{}{"a", int64(10)}, cmdline.UnmarshalOrReturnString("[a,10]"))
	assert.Equal([]interface{}{int64(10), "a"}, cmdline.UnmarshalOrReturnString("[10,a]"))

	assert.Equal([]interface{}{"a", true}, cmdline.UnmarshalOrReturnString("[a,true]"))
	assert.Equal([]interface{}{false, "a"}, cmdline.UnmarshalOrReturnString("[false,a]"))
}

func TestCmdlineParse(t *testing.T) {
	assert := require.New(t)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, cmdline.Parse("a b maxive.key1=value1 c maxive.key2=value2", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"key": "a,b",
		},
	}, cmdline.Parse("maxive.key=a,b", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"key": "a\nb",
		},
	}, cmdline.Parse("maxive.key=a\nb", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"key": "a b",
		},
	}, cmdline.Parse("maxive.key='a b'", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"key": "a:b",
		},
	}, cmdline.Parse("maxive.key=a:b", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"key": int64(5),
		},
	}, cmdline.Parse("maxive.key=5", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"rescue": true,
		},
	}, cmdline.Parse("maxive.rescue", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"keyArray": []interface{}{int64(1), int64(2)},
		},
	}, cmdline.Parse("maxive.keyArray=[1,2]", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"strArray": []interface{}{"url:http://192.168.1.100/cloud-config?a=b"},
		},
	}, cmdline.Parse("maxive.strArray=[\"url:http://192.168.1.100/cloud-config?a=b\"]", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"strArray": []interface{}{"url:http://192.168.1.100/cloud-config?a=b"},
		},
	}, cmdline.Parse("maxive.strArray=[url:http://192.168.1.100/cloud-config?a=b]", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"strArray": []interface{}{"part1 part2", "part3"},
		},
	}, cmdline.Parse("maxive.strArray=['part1 part2',part3]", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"strArray": []interface{}{"part1 part2", "part3"},
		},
	}, cmdline.Parse("maxive.strArray=[\"part1 part2\",part3]", false), false)

	assert.Equal(map[interface{}]interface{}{
		"maxive": map[interface{}]interface{}{
			"strArray": []interface{}{"part1 part2", "part3"},
		},
	}, cmdline.Parse("maxive.strArray=[ \"part1 part2\", part3 ]", false), false)
}

func TestGet(t *testing.T) {
	assert := require.New(t)

	data := map[interface{}]interface{}{
		"key": "value",
		"maxive": map[interface{}]interface{}{
			"key2": map[interface{}]interface{}{
				"subkey": "subvalue",
				"subnum": 42,
			},
		},
	}

	tests := map[string]interface{}{
		"key": "value",
		"maxive.key2.subkey":  "subvalue",
		"maxive.key2.subnum":  42,
		"maxive.key2.subkey2": "",
		"foo": "",
	}

	for k, v := range tests {
		val, _ := cmdline.GetOrSetVal(k, data, nil)
		assert.Equal(v, val)
	}
}

func TestSet(t *testing.T) {
	assert := require.New(t)

	data := map[interface{}]interface{}{
		"key": "value",
		"maxive": map[interface{}]interface{}{
			"key2": map[interface{}]interface{}{
				"subkey": "subvalue",
				"subnum": 42,
			},
		},
	}

	expected := map[interface{}]interface{}{
		"key": "value2",
		"maxive": map[interface{}]interface{}{
			"key2": map[interface{}]interface{}{
				"subkey":  "subvalue2",
				"subkey2": "value",
				"subkey3": 43,
				"subnum":  42,
			},
			"key3": map[interface{}]interface{}{
				"subkey3": 44,
			},
		},
		"key4": "value4",
	}

	tests := map[string]interface{}{
		"key": "value2",
		"maxive.key2.subkey":  "subvalue2",
		"maxive.key2.subkey2": "value",
		"maxive.key2.subkey3": 43,
		"maxive.key3.subkey3": 44,
		"key4":                 "value4",
	}

	for k, v := range tests {
		_, tData := cmdline.GetOrSetVal(k, data, v)
		val, _ := cmdline.GetOrSetVal(k, tData, nil)
		data = tData
		assert.Equal(v, val)
	}

	assert.Equal(expected, data)
}

type OuterData struct {
	One Data `yaml:"one"`
}

type Data struct {
	Two   bool `yaml:"two"`
	Three bool `yaml:"three"`
}

func TestMapMerge(t *testing.T) {
	assert := require.New(t)
	one := `
one:
  two: true`
	two := `
one:
  three: true`

	data := map[string]map[string]bool{}
	yaml.Unmarshal([]byte(one), &data)
	yaml.Unmarshal([]byte(two), &data)

	assert.NotNil(data["one"])
	assert.True(data["one"]["three"])
	assert.False(data["one"]["two"])

	data2 := &OuterData{}
	yaml.Unmarshal([]byte(one), data2)
	yaml.Unmarshal([]byte(two), data2)

	assert.True(data2.One.Three)
	assert.True(data2.One.Two)
}

func TestUserDocker(t *testing.T) {
	assert := require.New(t)

	config := &CloudConfig{
		Maxive: RancherConfig{
			Docker: DockerConfig{
				TLS: true,
			},
		},
	}

	bytes, err := yaml.Marshal(config)
	assert.Nil(err)

	config = &CloudConfig{}
	assert.False(config.Maxive.Docker.TLS)
	err = yaml.Unmarshal(bytes, config)
	assert.Nil(err)
	assert.True(config.Maxive.Docker.TLS)

	data := map[interface{}]interface{}{}
	err = util.Convert(config, &data)
	assert.Nil(err)

	val, ok := data["maxive"].(map[interface{}]interface{})["docker"]
	assert.True(ok)

	m, ok := val.(map[interface{}]interface{})
	assert.True(ok)
	v, ok := m["tls"]
	assert.True(ok)
	assert.True(v.(bool))

}
