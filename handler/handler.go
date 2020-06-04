package handler

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/gsakun/consul-sync/consul"
	consulapi "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

// Syncdata use for sync mysql machine table data to consul
func Syncdata(db *sql.DB, client *consulapi.Client) (errnum int, err error) {
	sql := "select hostname, ip, labels,monitor from machine where monitor != 1 and monitor !=4"
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
				log.Infof("Start register %s-%s", hostname, ip)
				err := register(hostname, ip, labels, monitor, db, client)
				if err != nil {
					log.Errorf("Register %s-%s failed errinfo %v", hostname, ip, err)
					errnum++
				}
			} else if monitor == 2 {
				serviceid := md5V3(fmt.Sprintf("%s-%s", hostname, ip))
				err := consul.ConsulFindServer(serviceid, client)
				if err != nil {
					err := register(hostname, ip, labels, monitor, db, client)
					if err != nil {
						log.Errorf("Register %s-%s failed errinfo %v", hostname, ip, err)
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
			} else if monitor == 3 {
				serviceid := md5V3(fmt.Sprintf("%s-%s", hostname, ip))
				log.Infof("Service id is %s", serviceid)
				err := consul.ConsulDeRegister(serviceid, client)
				if err != nil {
					log.Errorf("Deregister service %s failed", serviceid)
					errnum++
				}
				stmt, err := db.Prepare(`UPDATE machine set monitor=4 where ip=?`)
				if err != nil {
					log.Errorf("Update prepare err %v", err)
					errnum++
				}
				_, err = stmt.Exec(ip)
				if err != nil {
					log.Errorf("Update exec err %v", err)
					errnum++
				}
			} else {
				continue
			}
		}
		log.Infof("sync machineinfo success")
	}
	return errnum, err
}

func register(hostname, ip, labels string, monitor int, db *sql.DB, client *consulapi.Client) error {
	service := new(consul.Service)
	service.ID = md5V3(fmt.Sprintf("%s-%s", hostname, ip))
	service.Name = fmt.Sprintf("%s-%s", hostname, ip)
	var m map[string]interface{}
	err := json.Unmarshal([]byte(labels), &m)
	port, ok := m["port"]
	if ok {
		switch port.(type) {
		case string:
			num, err := strconv.Atoi(port.(string))
			if err != nil {
				service.Port = 9100
			} else {
				service.Port = num
			}
		case int:
			service.Port = port.(int)
		case float64:
			service.Port = int(port.(float64))
		}
	}
	var stringmap map[string]string = make(map[string]string)
	for key, value := range m {
		switch value.(type) {
		case string:
			stringmap[key] = value.(string)
		case int:
			stringmap[key] = strconv.Itoa(value.(int))
		default:
			continue
		}
	}
	service.Tags = stringmap
	service.Address = ip
	log.Infoln(*service)
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

func md5V3(str string) string {
	w := md5.New()
	io.WriteString(w, str)
	md5str := fmt.Sprintf("%x", w.Sum(nil))
	return md5str
}
