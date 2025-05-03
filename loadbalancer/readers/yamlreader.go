package readers

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Default path for config path
// Use env variable BALANCER_CONFIG_PATH to redeclare it
const DEFAULT_CONFIG_FILE_PATH = "../config/config.yaml"

type Healthcheck struct {
	Period                 *int    `yaml:"period"`
	ResponseTimeoutSeconds *int    `yaml:"responseTimeoutSeconds"`
	MaxRetries             *int    `yaml:"maxRetries"`
	TimeOutStep            *int    `yaml:"timeOutStep"`
	Endpoint               *string `yaml:"endpoint"`
}

type Balancer struct {
	ResponseHeaderTimeout *int `yaml:"responseHeaderTimeout"`
}

type Config struct {
	Listen      string       `yaml:"listen"`
	Balance     string       `yaml:"balance"`
	Healthcheck *Healthcheck `yaml:"healthcheck"`
	Upstreams   []string     `yaml:"upstreams"` // Теперь это срез структур Upstream
	Balancer    *Balancer    `yaml:"balancer"`
}

func ReadConfig(config *Config) error {
	config_file_path, err := validateConfigPath()
	if err != nil {
		return err
	}
	file, err := os.ReadFile(config_file_path)
	if err != nil {
		return fmt.Errorf(SomethingWentWrong, config_file_path)
	}

	if err := yaml.Unmarshal(file, config); err != nil {
		return err
	}
	return nil
}

const (
	ErrEnvFileNotExists   = "error occured during BALANCER_CONFIG_PATH reading: %s"
	DefautlFileNotExists  = "please specify config.yaml or use BALANCER_CONFIG_PATH env var: %s"
	ProvidedFileNotExists = "file %s does not exist"
	CanNotAccessFile      = "there are not enougth permissions to access %s"
	SomethingWentWrong    = "something went wrong during reading %s"
	WrongFileExtention    = "config should be yaml or yml file"
)

func validateConfigPath() (string, error) {
	env_config_file_path := os.Getenv("BALANCER_CONFIG_PATH")
	if env_config_file_path != "" {
		if err := checkExistAndReadable(env_config_file_path); err != nil {
			return "", errors.New(ErrEnvFileNotExists)
		}
		if err := checkExtention(env_config_file_path); err != nil {
			return "", err
		}
		return env_config_file_path, nil

	}
	if err := checkExistAndReadable(DEFAULT_CONFIG_FILE_PATH); err != nil {
		return "", errors.New(DefautlFileNotExists)
	}
	return DEFAULT_CONFIG_FILE_PATH, nil

}

func checkExtention(filePath string) error {
	splited := strings.Split(filePath, "/")
	file_ex := strings.Split(splited[len(splited)-1], ".")
	if file_ex[len(file_ex)-1] != "yaml" && file_ex[len(file_ex)-1] != "yml" {
		return errors.New(WrongFileExtention)
	}
	return nil

}

func checkExistAndReadable(filePath string) error {

	file, err := os.Open(filePath)

	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(ProvidedFileNotExists, filePath)
		} else if os.IsPermission(err) {
			return fmt.Errorf(CanNotAccessFile, filePath)
		}
		return fmt.Errorf(SomethingWentWrong, filePath)
	}
	defer file.Close()

	return nil
}
