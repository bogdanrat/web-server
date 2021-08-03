package models

type KeyValuePair struct {
	Key   string
	Value interface{}
}

type GetPairRequest struct {
	Key string `form:"key"`
}

type PutPairRequest struct {
	Key   string `form:"key"`
	Value string `form:"value"`
}

type DeletePairRequest struct {
	Key string `form:"key"`
}

type GetAllPairsRequest struct {
}
