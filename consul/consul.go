package consul

import (
	"fmt"

	consulapi "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

// Service define struct use for parse httprequest
type Service struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Port    int               `json:"port"`
	Tags    map[string]string `json:"tags"`
	Address string            `json:"address"`
}

// InitClient use for init consulagent client
func InitClient(consuladdress string) (client *consulapi.Client, err error) {
	config := consulapi.DefaultConfig()
	config.Address = consuladdress
	client, err = consulapi.NewClient(config)
	if err != nil {
		log.Errorf("Connect consul service failed errinfo: %v", err)
		return nil, fmt.Errorf("Connect consul service failed errinfo: %v", err)
	}
	return client, nil
}

// ConsulRegister use for register service to consul
func ConsulRegister(service *Service, client *consulapi.Client) error {
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = service.ID
	registration.Name = service.Name
	registration.Port = service.Port
	var tags []string
	if len(service.Tags) != 0 {
		for key, value := range service.Tags {
			tag := fmt.Sprintf("%s=%s", key, value)
			tags = append(tags, tag)
		}
	}
	registration.Tags = tags
	registration.Address = service.Address
	registration.EnableTagOverride = true
	//log.Infoln(registration)
	err := client.Agent().ServiceRegisterOpts(registration, consulapi.ServiceRegisterOpts{
		ReplaceExistingChecks: true,
	})
	if err != nil {
		log.Errorf(" %v failed,errinfo %v", registration, err)
		return err
	}
	return nil
}

// ConsulDeRegister use for DeRegister service
func ConsulDeRegister(service string, client *consulapi.Client) error {
	for i := 1; i < 4; i++ {
		err := client.Agent().ServiceDeregister(service)
		if err != nil {
			log.Errorf("Deregistry service %s failed", service)
			return err
		}
	}
	return nil
}

// ConsulFindServer use for query service in consul
func ConsulFindServer(service string, client *consulapi.Client) error {
	// 获取指定service
	serviceinfo, meta, err := client.Agent().Service(service, nil)
	if err != nil {
		log.Errorf("Query specified service %s error : %v ", service, err)
		return fmt.Errorf("Query specified service %s error : %v ", service, err)
	}
	log.Infof("Service Info %v", *serviceinfo)
	log.Infof("Meta %v", meta)
	return nil
}
