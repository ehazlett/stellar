package application

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/runtime/restart"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/application/v1"
	ptypes "github.com/gogo/protobuf/types"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

const (
	defaultSnapshotter = "overlayfs"
)

func (s *service) Create(ctx context.Context, req *api.CreateRequest) (*ptypes.Empty, error) {
	// TODO: move this to the node service for creation of containers
	logrus.Debugf("creating application %s", req.Name)
	ctx = namespaces.WithNamespace(ctx, s.namespace)
	for _, service := range req.Services {
		if _, err := s.newContainer(ctx, req.Name, service); err != nil {
			return empty, err
		}
	}
	return empty, nil
}

func (s *service) newContainer(ctx context.Context, appName string, service *api.Service) (containerd.Container, error) {
	var (
		opts  []oci.SpecOpts
		cOpts []containerd.NewContainerOpts
	)

	client, err := s.containerd()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	id := fmt.Sprintf("%s.%s", appName, service.Name)
	opts = append(opts, oci.WithEnv(service.Process.Env), withMounts(service.Mounts))
	cOpts = append(cOpts, containerd.WithContainerLabels(convertLabels(service.Labels)))
	if service.Runtime != "" {
		cOpts = append(cOpts, containerd.WithRuntime(service.Runtime, nil))
	}
	snapshotter := defaultSnapshotter
	if service.Snapshotter != "" {
		snapshotter = service.Snapshotter
	}
	image, err := client.GetImage(ctx, service.Image)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, err
		}
		// pull
		img, err := client.Pull(ctx, service.Image, containerd.WithPullUnpack)
		if err != nil {
			return nil, err
		}
		image = img
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
		containerd.WithNewSnapshot(id, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
		containerd.WithContainerLabels(map[string]string{
			stellar.StellarApplicationLabel: appName,
			stellar.StellarNetworkLabel:     "true",
		}),
		restart.WithStatus(containerd.Running),
	)

	container, err := client.NewContainer(ctx, id, cOpts...)
	if err != nil {
		return nil, err
	}

	// TODO: redirect output somewhere
	task, err := container.NewTask(ctx, cio.NullIO)
	if err != nil {
		return nil, err
	}
	if err := task.Start(ctx); err != nil {
		return nil, err
	}

	// TODO: setup networking
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	subnetCIDR, err := c.Network().GetSubnet(s.nodeName())
	if err != nil {
		return nil, err
	}
	ip, err := c.Network().AllocateIP(id, s.nodeName(), subnetCIDR)
	if err != nil {
		return nil, err
	}
	gw, err := gateway(subnetCIDR)
	if err != nil {
		return nil, err
	}
	if err := c.Node().SetupContainerNetwork(id, ip.String(), subnetCIDR, gw); err != nil {
		return nil, err
	}
	return container, nil

}

func withMounts(mounts []*api.Mount) oci.SpecOpts {
	return func(ctx context.Context, _ oci.Client, c *containers.Container, s *oci.Spec) error {
		for _, cm := range mounts {
			if cm.Type == "bind" {
				// create source dir if it does not exist
				if err := os.MkdirAll(filepath.Dir(cm.Source), 0755); err != nil {
					return err
				}
				if err := os.Mkdir(cm.Source, 0755); err != nil {
					if !os.IsExist(err) {
						return err
					}
				} else {
					if err := os.Chown(cm.Source, int(s.Process.User.UID), int(s.Process.User.GID)); err != nil {
						return err
					}
				}
			}
			s.Mounts = append(s.Mounts, specs.Mount{
				Type:        cm.Type,
				Source:      cm.Source,
				Destination: cm.Destination,
				Options:     cm.Options,
			})
		}
		return nil
	}
}

func gateway(subnetCIDR string) (string, error) {
	ip, ipnet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return "", err
	}

	gw := ip.Mask(ipnet.Mask)
	gw[3]++

	return gw.String(), nil
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
