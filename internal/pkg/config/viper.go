package config

import (
	"bytes"
	"encoding/base64"
	"errors"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
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

	v.OnConfigChange(func(_ fsnotify.Event) {
		if err := v.ReadInConfig(); err != nil {
			slog.Error("config reload failed", "path", pathFile, "err", err)
			return
		}
		slog.Info("config success reloaded", "path", pathFile)
	})
	v.WatchConfig()

	return &Viper{v: v}, nil
}

// NewViperFromBytes loads configuration from memory and returns a Viper-backed Config.
// configType should be a format supported by Viper (e.g. "yaml", "json", "toml").
func NewViperFromBytes(configType string, data []byte) (*Viper, error) {
	if strings.TrimSpace(configType) == "" {
		return nil, errors.New("config type is required")
	}

	v := viper.New()
	v.SetConfigType(configType)

	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, err
	}

	return &Viper{v: v}, nil
}

// GetInt returns the value for key as int.
func (vc *Viper) GetInt(key string) int {
	return vc.v.GetInt(key)
}

// GetInt32 returns the value for key as int32.
func (vc *Viper) GetInt32(key string) int32 {
	return vc.v.GetInt32(key)
}

// GetInt64 returns the value for key as int64.
func (vc *Viper) GetInt64(key string) int64 {
	return vc.v.GetInt64(key)
}

// GetUint returns the value for key as uint.
func (vc *Viper) GetUint(key string) uint {
	return vc.v.GetUint(key)
}

// GetUint16 returns the value for key as uint16.
func (vc *Viper) GetUint16(key string) uint16 {
	return uint16(vc.v.GetUint(key))
}

// GetUint32 returns the value for key as uint32.
func (vc *Viper) GetUint32(key string) uint32 {
	return vc.v.GetUint32(key)
}

// GetUint64 returns the value for key as uint64.
func (vc *Viper) GetUint64(key string) uint64 {
	return vc.v.GetUint64(key)
}

// GetBool returns the value for key as bool.
func (vc *Viper) GetBool(key string) bool {
	return vc.v.GetBool(key)
}

// GetFloat32 returns the value for key as float32.
func (vc *Viper) GetFloat32(key string) float32 {
	return float32(vc.v.GetFloat64(key))
}

// GetFloat64 returns the value for key as float64.
func (vc *Viper) GetFloat64(key string) float64 {
	return vc.v.GetFloat64(key)
}

// GetSecond returns the value for key as seconds.
func (vc *Viper) GetSecond(key string) time.Duration {
	return time.Duration(vc.v.GetInt64(key)) * time.Second
}

// GetMinute returns the value for key as minutes.
func (vc *Viper) GetMinute(key string) time.Duration {
	return time.Duration(vc.v.GetInt64(key)) * time.Minute
}

// GetHour returns the value for key as hours.
func (vc *Viper) GetHour(key string) time.Duration {
	return time.Duration(vc.v.GetInt64(key)) * time.Hour
}

// GetDay returns the value for key as days (24h).
func (vc *Viper) GetDay(key string) time.Duration {
	return time.Duration(vc.v.GetInt64(key)) * 24 * time.Hour
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
