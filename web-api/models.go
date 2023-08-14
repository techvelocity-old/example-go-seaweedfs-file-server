package main

import (
	"gorm.io/gorm"
)

type MasterResponse struct {
	Count     int    `json:"count"`
	FID       string `json:"fid"`
	URL       string `json:"url"`
	PublicURL string `json:"publicUrl"`
}

type Location struct {
	PublicURL string `json:"publicUrl"`
	URL       string `json:"url"`
}

type Volume struct {
	VolumeID  string     `json:"volumeId"`
	Locations []Location `json:"locations"`
}

type FileRecord struct {
	gorm.Model
	FID      string `json:"fid"`
	FileName string `json:"fileName"`
}
