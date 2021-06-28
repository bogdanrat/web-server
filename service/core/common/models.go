package common

type DatabaseSecrets struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DbName   string `json:"dbInstanceIdentifier"`
}
