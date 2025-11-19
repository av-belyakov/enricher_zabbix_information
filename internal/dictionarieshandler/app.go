package dictionarieshandler

import (
	"errors"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

func New[T interfaces.StructConstraint](pathToFileDts, initialMarker string, dict T) (T, error) {
	if pathToFileDts == "" || initialMarker == "" {
		return dict, wrappers.WrapperError(errors.New("parameters 'pathToFileDts' and 'initialMarker' cannot be empty"))
	}

	rootPath, err := supportingfunctions.GetRootPath(constants.Root_Dir)
	if err != nil {
		return dict, wrappers.WrapperError(err)
	}

	dictionaryPath := filepath.Join(rootPath, pathToFileDts)
	viper.SetConfigFile(dictionaryPath)
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		return dict, wrappers.WrapperError(err)
	}

	if ok := viper.IsSet(initialMarker); ok {
		if err := viper.GetViper().Unmarshal(dict); err != nil {
			return dict, err
		}
	}

	return dict, nil
}
