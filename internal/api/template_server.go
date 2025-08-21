package api

import "github.com/JrMarcco/kuryr/internal/service/template"

type TemplateServer struct {
	svc template.Service
}

func NewTemplateServer(svc template.Service) *TemplateServer {
	return &TemplateServer{
		svc: svc,
	}
}
