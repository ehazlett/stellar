package application

import (
	"context"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/oci"
	api "github.com/ehazlett/stellar/api/services/application/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Create(ctx context.Context, req *api.CreateRequest) (*ptypes.Empty, error) {
	logrus.Debugf("creating application %s", req.Name)
	for _, service := range req.Services {
		if _, err := s.newContainer(ctx, service); err != nil {
			return empty, err
		}
	}
	return empty, nil
}

func (s *service) newContainer(ctx context.Context, service *api.Service) (containerd.Container, error) {
	var (
		opts  []oci.SpecOpts
		cOpts []containerd.NewContainerOpts
		//spec  containerd.NewContainerOpts
	)

	client, err := s.containerd()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	opts = append(opts, oci.WithEnv(service.Process.Env))
	cOpts = append(cOpts, containerd.WithContainerLabels(convertLabels(service.Labels)))
	cOpts = append(cOpts, containerd.WithRuntime(service.Runtime, nil))
	snapshotter := service.Snapshotter
	image, err := client.GetImage(ctx, service.Image)
	if err != nil {
		return nil, err
	}
	unpacked, err := image.IsUnpacked(ctx, snapshotter)
	if err != nil {
		return nil, err
	}
	if !unpacked {
		if err := image.Unpack(ctx, snapshotter); err != nil {
			return nil, err
		}
	}
	opts = append(opts, oci.WithImageConfig(image))
	cOpts = append(cOpts,
		containerd.WithImage(image),
		containerd.WithSnapshotter(snapshotter),
		containerd.WithNewSnapshot(service.Name, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)

	container, err := client.NewContainer(ctx, service.Name, cOpts...)
	if err != nil {
		return nil, err
	}

	//// TODO: start
	//if _, err := container.NewTask(ctx, cio.Stdio); err != nil {
	//	return nil, err
	//}

	return container, nil

}

func convertLabels(values []string) map[string]string {
	labels := map[string]string{}
	for _, s := range values {
		p := strings.Split(s, "=")
		k := p[0]
		v := ""
		if len(p) > 1 {
			v = strings.Join(p[1:], "")
		}
		labels[k] = v
	}
	return labels
}
