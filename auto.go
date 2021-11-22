package autogcp

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/api/cloudresourcemanager/v1"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Metadata struct {
	rc api.Config

	projectNumber      string
	projectNumberOnce  sync.Once
	projectNumberError error
}

func NewMetadata() (*Metadata, error) {
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), nil)
	rc, err := loader.RawConfig()
	if err != nil {
		return nil, err
	}
	return &Metadata{rc: rc}, nil
}

func (m *Metadata) ClusterName() string {
	return m.rc.Contexts[m.rc.CurrentContext].Cluster
}

func (m *Metadata) ProjectName() string {
	p, _ := parseClusterName(m.ClusterName())
	return p
}

func (m *Metadata) ProjectNumber() (string, error) {
	m.projectNumberOnce.Do(func() {
		name := m.ProjectName()
		ctx := context.Background()
		cloudresourcemanagerService, err := cloudresourcemanager.NewService(ctx)
		if err != nil {
			m.projectNumberError = err
			return
		}
		res, err := cloudresourcemanagerService.Projects.Get(name).Do()
		if err != nil {
			m.projectNumberError = err
			return
		}
		m.projectNumber = fmt.Sprint(res.ProjectNumber)
	})
	return m.projectNumber, m.projectNumberError
}

func (m *Metadata) Location() string {
	_, l := parseClusterName(m.ClusterName())
	return l
}

func parseClusterName(c string) (project, location string) {
	if !strings.HasPrefix(c, "gke_") {
		return
	}
	parts := strings.Split(c, "_")
	if len(parts) != 4 {
		return
	}
	return parts[1], parts[2]
}
