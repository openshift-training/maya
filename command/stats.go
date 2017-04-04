package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Volume struct {
	Spec struct {
		AccessModes interface{} `json:"AccessModes"`
		Capacity    interface{} `json:"Capacity"`
		ClaimRef    interface{} `json:"ClaimRef"`
		OpenEBS     struct {
			VolumeID string `json:"volumeID"`
		} `json:"OpenEBS"`
		PersistentVolumeReclaimPolicy string `json:"PersistentVolumeReclaimPolicy"`
		StorageClassName              string `json:"StorageClassName"`
	} `json:"Spec"`

	Status struct {
		Message string `json:"Message"`
		Phase   string `json:"Phase"`
		Reason  string `json:"Reason"`
	} `json:"Status"`
	Annotations       interface{} `json:"annotations"`
	CreationTimestamp interface{} `json:"creationTimestamp"`
	Name              string      `json:"name"`
}

type Annotations struct {
	VolSize      string   `json:"be.jiva.volume.openebs.io/vol-size"`
	VolAddr      string   `json:"fe.jiva.volume.openebs.io/ip"`
	Iqn          string   `json:"iqn"`
	Targetportal string   `json:"targetportal"`
	Replicas     []string `json:"JIVA_REP_IP_*"`
	ReplicaCount string   `json:"be.jiva.volume.openebs.io/count"`
}

const (
	timeout = 2 * time.Second
)

func getVolDetails(volName string, obj interface{}) error {
	addr := os.Getenv("MAPI_ADDR")
	fmt.Println("ADDR =", addr)
	url := addr + "/latest/volume/info/" + volName
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj)
}

func GetVolAnnotations(volName string) (*Annotations, error) {
	var volume Volume
	var annotations Annotations
	err := getVolDetails(volName, &volume)
	if err != nil {
		return nil, err
	}
	for key, value := range volume.Annotations.(map[string]interface{}) {
		switch key {
		case "be.jiva.volume.openebs.io/vol-size":
			annotations.VolSize = value.(string)
		case "fe.jiva.volume.openebs.io/ip":
			annotations.VolAddr = value.(string)
		case "iqn":
			annotations.Iqn = value.(string)
		case "be.jiva.volume.openebs.io/count":
			annotations.ReplicaCount = value.(string)
		case "JIVA_REP_IP_0":
			annotations.Replicas = append(annotations.Replicas, value.(string))
		case "JIVA_REP_IP_1":
			annotations.Replicas = append(annotations.Replicas, value.(string))

		}
	}
	return &annotations, nil
}
