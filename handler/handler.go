package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gsakun/consul-sync/consul"
	consulapi "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

func syncdata(db *sql.DB, client *consulapi.Client) {
	sql := "select hostname, ip, labels,monitor from machine where monitor != 1"
	rows, err := db.Query(sql)
	if err != nil {
		log.Errorf("Query machine table Failed")
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				hostname string
				ip       string
				labels   string
				monitor  int
			)
			err = rows.Scan(&hostname, &ip, &labels, &monitor)
			if err != nil {
				log.Errorf("ERROR: %v", err)
				continue
			}
			if monitor == 0 {
				service := new(consul.Service)
				service.ID = fmt.Sprintf("%s-%s", hostname, ip)
				var m map[string]interface{}
				err = json.Unmarshal([]byte(labels), &m)
				port, ok := m["port"]
				if ok {
					service.Port = port.(int)
				} else {
					service.Port = 9100
				}
				service.Tags = m
				service.Address = ip
				err := consul.ConsulRegister(service)
				if err != nil {
					return err
				} else {
					stmt, err := db.Prepare(`UPDATE machine set monitor=1 where ip=?`)
					if err != nil {
						log.Errorf("Update prepare err %v", err)
						return err
					}
					_, err = stmt.Exec(ip)
					if err != nil {
						log.Errorf("Update exec err %v", err)
						return err
					}
				}
			}else{
				serviceid := fmt.Sprintf("%s-%s", hostname, ip)
				err := consul.ConsulFindServer(serviceid)
				if err != nil {
					
				}else{
					err := consul.ConsulDeRegister(serviceid)
					if err != nil {
						log.Errorf("Deregister service %s failed",serviceid)
						continue
					}else{

					}
				}
			}

		}
	}
	log.Infof("sync datacenter table success %v", DataCentermap)
}

func 
