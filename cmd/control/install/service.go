package install

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/util"
	"github.com/maxive/os/pkg/util/network"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
)

type ImageConfig struct {
	Image string `yaml:"image,omitempty"`
}

func GetCacheImageList(cloudconfig string, oldcfg *config.CloudConfig) []string {
	savedImages := make([]string, 0)
	bytes, err := readConfigFile(cloudconfig)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to read cloud-config")
		return savedImages
	}
	r := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(bytes, &r); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to unmarshal cloud-config")
		return savedImages
	}
	newcfg := &config.CloudConfig{}
	if err := util.Convert(r, newcfg); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to convert cloud-config")
		return savedImages
	}

	// services_include
	for key, value := range newcfg.Maxive.ServicesInclude {
		if value {
			serviceImage := getServiceImage(key, "", oldcfg, newcfg)
			if serviceImage != "" {
				savedImages = append(savedImages, serviceImage)
			}
		}
	}

	// console
	newConsole := newcfg.Maxive.Console
	if newConsole != "" && newConsole != "default" {
		consoleImage := getServiceImage(newConsole, "console", oldcfg, newcfg)
		if consoleImage != "" {
			savedImages = append(savedImages, consoleImage)
		}
	}

	// docker engine
	newEngine := newcfg.Maxive.Docker.Engine
	if newEngine != "" && newEngine != oldcfg.Maxive.Docker.Engine {
		engineImage := getServiceImage(newEngine, "docker", oldcfg, newcfg)
		if engineImage != "" {
			savedImages = append(savedImages, engineImage)
		}

	}

	return savedImages
}

func getServiceImage(service, svctype string, oldcfg, newcfg *config.CloudConfig) string {
	var (
		serviceImage string
		bytes        []byte
		err          error
	)
	if len(newcfg.Maxive.Repositories.ToArray()) > 0 {
		bytes, err = network.LoadServiceResource(service, true, newcfg)
	} else {
		bytes, err = network.LoadServiceResource(service, true, oldcfg)
	}
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to load service resource")
		return serviceImage
	}
	imageConfig := map[interface{}]ImageConfig{}
	if err = yaml.Unmarshal(bytes, &imageConfig); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to unmarshal service")
		return serviceImage
	}
	switch svctype {
	case "console":
		serviceImage = formatImage(imageConfig["console"].Image, oldcfg, newcfg)
	case "docker":
		serviceImage = formatImage(imageConfig["docker"].Image, oldcfg, newcfg)
	default:
		serviceImage = formatImage(imageConfig[service].Image, oldcfg, newcfg)
	}

	return serviceImage
}

func RunCacheScript(partition string, images []string) error {
	return util.RunScript("/scripts/cache-services.sh", partition, strings.Join(images, " "))
}

func readConfigFile(file string) ([]byte, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			content = []byte{}
		} else {
			return nil, err
		}
	}
	return content, err
}

func formatImage(image string, oldcfg, newcfg *config.CloudConfig) string {
	registryDomain := newcfg.Maxive.Environment["REGISTRY_DOMAIN"]
	if registryDomain == "" {
		registryDomain = oldcfg.Maxive.Environment["REGISTRY_DOMAIN"]
	}
	image = strings.Replace(image, "${REGISTRY_DOMAIN}", registryDomain, -1)

	image = strings.Replace(image, "${SUFFIX}", config.Suffix, -1)

	return image
}
