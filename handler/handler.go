package handler

import (
	"github.com/gsakun/consul-sync/db"
	"github.com/gsakun/consul-sync/consul"
	log "github.com/sirupsen/logrus"
)

func syncregister() {
	sql := "select hostname, ip, labels from machine"
	rows, err := db.DB.Query(sql)
	if err != nil {
		log.Errorf("Query data_center table Failed")
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				hostname string
				ip string
				labels string
			)

			err = rows.Scan(&hostname, &ip,&labels)
			if err != nil {
				log.Errorf("ERROR: %v", err)
				continue
			}
			
		}
	}
	log.Infof("sync datacenter table success %v", DataCentermap)
}

func syncderegister() {

}

func syncupdatelabels() {
	sql := "select hostname, ip, labels from machine"
	rows, err := db.DB.Query(sql)
	if err != nil {
		log.Errorf("Query data_center table Failed")
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				hostname string
				ip string
				labels string
			)

			err = rows.Scan(&hostname, &ip,&labels)
			if err != nil {
				log.Errorf("ERROR: %v", err)
				continue
			}
			DataCentermap[dataCenterName] = id
		}
	}
	log.Infof("sync datacenter table success %v", DataCentermap)
}
