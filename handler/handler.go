package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gsakun/consul-sync/consul"
	consulapi "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

// Syncdata use for sync mysql machine table data to consul
func Syncdata(db *sql.DB, client *consulapi.Client) (errnum int, err error) {
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
				errnum++
				continue
			}
			if monitor == 0 {
				err := register(hostname, ip, labels, monitor, db, client)
				if err != nil {
					log.Errorf("Register %s failed errinfo %v", err)
					errnum++
				}
			} else {
				serviceid := fmt.Sprintf("%s-%s", hostname, ip)
				err := consul.ConsulFindServer(serviceid, client)
				if err != nil {
					err := register(hostname, ip, labels, monitor, db, client)
					if err != nil {
						log.Errorf("Register %s failed errinfo %v", err)
						errnum++
					}
				} else {
					err := consul.ConsulDeRegister(serviceid, client)
					if err != nil {
						log.Errorf("Deregister service %s failed", serviceid)
						errnum++
					} else {
						err := register(hostname, ip, labels, monitor, db, client)
						if err != nil {
							log.Errorf("Register %s failed errinfo %v", err)
						}
					}
				}
			}
		}
		log.Infof("sync machineinfo success")
	}
	return errnum, err
}

func register(hostname, ip, labels string, monitor int, db *sql.DB, client *consulapi.Client) error {
	service := new(consul.Service)
	service.ID = fmt.Sprintf("%s-%s", hostname, ip)
	var m map[string]interface{}
	err := json.Unmarshal([]byte(labels), &m)
	port, ok := m["port"]
	if ok {
		service.Port = port.(int)
	} else {
		service.Port = 9100
	}
	service.Tags = m
	service.Address = ip
	err = consul.ConsulRegister(service, client)
	if err != nil {
		return err
	}
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
	return nil
}
