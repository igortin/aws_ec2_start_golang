package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// ConsoleWriter custom Interface
type ConsoleWriter interface {
	String()
}

// ShowOutput support ConsoleWriter
func ShowOutput(c ConsoleWriter) {
	c.String()
}

// Ec2object structure
type Ec2object struct {
	InstanceID       string `json:"InstanceId"`
	InstanceState    string `json:"InstanceState"`
	PrivateDNSName   string `json:"PrivateDnsName"`
	PublicDNSame     string `json:"PublicDnsName"`
	BlockDevicesList []*BlockDevice
}

// BlockDevice structure
type BlockDevice struct {
	DeviceName string  `json:"DeviceName"`
	State      string  `json:"State"`
	VolumeID   string  `json:"VolumeId"`
	VolumeTag  ec2.Tag `json:"VolumeTag"`
}

// String method for Ec2object
func (eo *Ec2object) String() {
	fmt.Println("Instance ID:", eo.InstanceID)
	fmt.Println("Instance State", eo.InstanceState)
	// fmt.Println("Public Hosname:", eo.PublicDNSame)
	// fmt.Println("Private Hostname:", eo.PrivateDNSName)
	fmt.Println("Attached volumes number:", len(eo.BlockDevicesList))
	for _, vol := range eo.BlockDevicesList {
		fmt.Printf("\tVolume id: %s %v\n", vol.VolumeID, vol.VolumeTag)
	}
}

// ToJSON method for Ec2object
func (eo *Ec2object) ToJSON() []byte {
	b, err := json.Marshal(eo)
	if err != nil {
		log.Println("Error: Structure can not be Marshaled to JSON")
	}
	return b
}
