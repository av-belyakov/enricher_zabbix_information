package dictionarieshandler

import (
	"errors"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

func Read(pathToFileDts string) (*ListDictionaries, error) {
	ld := &ListDictionaries{}

	if pathToFileDts == "" {
		return ld, wrappers.WrapperError(errors.New("parameter 'pathToFileDts' cannot be empty"))
	}

	rootPath, err := supportingfunctions.GetRootPath(constants.Root_Dir)
	if err != nil {
		return ld, wrappers.WrapperError(err)
	}

	dictionaryPath := filepath.Join(rootPath, pathToFileDts)
	viper.SetConfigFile(dictionaryPath)
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		return ld, wrappers.WrapperError(err)
	}

	if ok := viper.IsSet("dictionaries"); ok {
		if err := viper.GetViper().Unmarshal(ld); err != nil {
			return ld, err
		}
	}

	return ld, nil
}
