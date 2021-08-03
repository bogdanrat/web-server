package models

type KeyValuePair struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type GetPairRequest struct {
	Key string `form:"key" json:"key"`
}

type GetPairResponse struct {
	Key   string `form:"key" json:"key"`
	Value string `form:"value" json:"value"`
}

type PostPairRequest struct {
	KeyValuePair
}

type DeletePairRequest struct {
	Key string `form:"key"`
}

type GetAllPairsRequest struct {
}
