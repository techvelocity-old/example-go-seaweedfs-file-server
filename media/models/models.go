package models

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
