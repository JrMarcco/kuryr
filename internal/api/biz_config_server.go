package api

import configv1 "github.com/JrMarcco/kuryr-api/api/config/v1"

type BizConfigServer struct {
	configv1.UnimplementedBizConfigServiceServer
}

func NewBizConfigServer() *BizConfigServer {
	return &BizConfigServer{}
}
