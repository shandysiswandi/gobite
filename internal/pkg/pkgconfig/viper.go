package pkgconfig

import (
	"encoding/base64"
	"path"
	"strings"

	"github.com/spf13/viper"
)

// Viper is a Config implementation backed by github.com/spf13/viper.
type Viper struct {
	v *viper.Viper
}

// NewViper loads configuration from the given file path and returns a Viper-backed Config.
//
// The config file type is inferred by Viper from the filename extension.
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

// GetInt returns the value for key as int64.
func (vc *Viper) GetInt(key string) int64 {
	return vc.v.GetInt64(key)
}

// GetBool returns the value for key as bool.
func (vc *Viper) GetBool(key string) bool {
	return vc.v.GetBool(key)
}

// GetFloat returns the value for key as float64.
func (vc *Viper) GetFloat(key string) float64 {
	return vc.v.GetFloat64(key)
}

// GetString returns the value for key as string.
func (vc *Viper) GetString(key string) string {
	return vc.v.GetString(key)
}

// GetBinary returns the value for key decoded from base64.
func (vc *Viper) GetBinary(key string) []byte {
	data, err := base64.StdEncoding.DecodeString(vc.v.GetString(key))
	if err != nil {
		return nil
	}

	return data
}

// GetArray returns the value for key split by commas.
func (vc *Viper) GetArray(key string) []string {
	return strings.Split(vc.v.GetString(key), ",")
}

// GetMap returns the value for key parsed from "k:v,k:v" pairs.
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

// Close implements io.Closer for interface compatibility.
func (vc *Viper) Close() error {
	// No resources to close for ViperConfig; this is just for interface completeness.
	return nil
}
