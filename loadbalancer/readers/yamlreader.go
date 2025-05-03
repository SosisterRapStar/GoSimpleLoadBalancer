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
const (
	BALANCER_CONFIG_PATH         = "BALANCER_CONFIG_PATH"
	ERR_ENV_FILE_NOT_EXISTS      = "error occured during BALANCER_CONFIG_PATH reading: %s"
	DEFAULT_FILE_NOT_EXISTS      = "please specify config.yaml in %s or use BALANCER_CONFIG_PATH env var: %s"
	DEFAULT_ETC_CONFIG_FILE_PATH = "/gosimplebalancer/config.yaml"
)

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
		return err
	}

	if err := yaml.Unmarshal(file, config); err != nil {
		return err
	}
	return nil
}

func validateConfigPath() (string, error) {
	env_config_file_path := os.Getenv(BALANCER_CONFIG_PATH)
	if env_config_file_path != "" {
		if err := checkExistAndReadable(env_config_file_path); err != nil {
			return "", err
		}
		if err := checkExtention(env_config_file_path); err != nil {
			return "", err
		}
		return env_config_file_path, nil
	}

	return DEFAULT_ETC_CONFIG_FILE_PATH, nil

}

func checkExtention(filePath string) error {
	splited := strings.Split(filePath, "/")
	file_ex := strings.Split(splited[len(splited)-1], ".")
	if file_ex[len(file_ex)-1] != "yaml" && file_ex[len(file_ex)-1] != "yml" {
		return fmt.Errorf("file %s should be yaml or yml file", filePath)
	}
	return nil
}

var ErrorSomethingWentWrong = errors.New("something went wrong during reading the file")

func checkExistAndReadable(filePath string) error {

	file, err := os.Open(filePath)

	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("provided file path %s does not exists", filePath)
		} else if os.IsPermission(err) {
			return fmt.Errorf("can not access the file path provided in env %s", filePath)
		}
		return errors.Join(ErrorSomethingWentWrong, err)
	}
	defer file.Close()

	return nil
}
