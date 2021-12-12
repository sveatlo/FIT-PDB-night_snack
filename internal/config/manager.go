package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Manager struct {
	files []string
	viper *viper.Viper

	conf Config
}

func NewManager(configPtr Config, files ...string) (*Manager, error) {
	m := &Manager{
		viper: viper.New(),
		files: files,

		conf: configPtr,
	}

	for _, file := range m.files {
		m.viper.SetConfigFile(file)
	}

	return m, nil
}

func (m *Manager) Load() (err error) {
	err = m.viper.ReadInConfig()
	if err != nil {
		err = fmt.Errorf("cannot read config: %v", err)
		return
	}

	err = m.viper.Unmarshal(m.conf)
	if err != nil {
		err = fmt.Errorf("config loading failed at unmarshal: %v", err)
		return err
	}

	err = m.conf.PostLoad()
	if err != nil {
		err = fmt.Errorf("config loading failed at post hook: %v", err)
		return
	}

	return nil
}
