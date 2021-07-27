package models
type Output struct {
	Items []struct {
		Address struct 
		{
			State string `json:"state"`
			CountryName string `json:"countryName"`
		} `json:"address"`
	} `json:"items"`
}

type Entry struct {
	State string
	Cases string
	Last_Updated string
}
