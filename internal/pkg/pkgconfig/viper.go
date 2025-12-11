package pkgconfig

import (
	"encoding/base64"
	"path"
	"strings"

	"github.com/spf13/viper"
)

type Viper struct {
	v *viper.Viper
}

func NewViper(pathFile string) (*Viper, error) {
	v := viper.New()

	filename := path.Base(pathFile)
	filePath := path.Dir(pathFile)

	configName := path.Base(filename[:len(filename)-len(path.Ext(filename))])

	v.AddConfigPath(filePath)
	v.SetConfigName(configName)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	v.WatchConfig()

	return &Viper{v: v}, nil
}

func (vc *Viper) GetInt(key string) int64 {
	return vc.v.GetInt64(key)
}

func (vc *Viper) GetBool(key string) bool {
	return vc.v.GetBool(key)
}

func (vc *Viper) GetFloat(key string) float64 {
	return vc.v.GetFloat64(key)
}

func (vc *Viper) GetString(key string) string {
	return vc.v.GetString(key)
}

func (vc *Viper) GetBinary(key string) []byte {
	data, err := base64.StdEncoding.DecodeString(vc.v.GetString(key))
	if err != nil {
		return nil
	}

	return data
}

func (vc *Viper) GetArray(key string) []string {
	return strings.Split(vc.v.GetString(key), ",")
}

func (vc *Viper) GetMap(key string) map[string]string {
	pairs := strings.Split(vc.v.GetString(key), ",")
	m := make(map[string]string)

	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}

	return m
}

// Close performs any necessary cleanup.
func (vc *Viper) Close() error {
	// No resources to close for ViperConfig; this is just for interface completeness.
	return nil
}
