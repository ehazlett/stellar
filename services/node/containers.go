package node

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/runtime/restart"
	gocni "github.com/containerd/go-cni"
	"github.com/containerd/typeurl"
	"github.com/ehazlett/stellar"
	nameserverapi "github.com/ehazlett/stellar/api/services/nameserver/v1"
	api "github.com/ehazlett/stellar/api/services/node/v1"
	"github.com/ehazlett/stellar/api/types"
	ptypes "github.com/gogo/protobuf/types"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const (
	defaultSnapshotter = "overlayfs"
	defaultIfName      = "eth0"
	cniLoopbackConf    = `{
	"cniVersion": "0.4.0",
	"name": "loopback",
	"type": "loopback",
        "ipam": {
                "type": "static",
		"addresses": [
		    {
			"address": "127.0.0.1/8"
		    }
		]
        }
}
	`
	cniConfTemplate = `{
        "cniVersion": "0.4.0",
        "name": "stellar",
        "type": "bridge",
        "bridge": "{{.Bridge}}",
        "isGateway": true,
	"hairpinMode": true,
        "ipMasq": true,
        "ipam": {
                "type": "stellar-cni-ipam",
                "node_name": "{{.NodeName}}",
                "peer_addr": "{{.PeerAddr}}"
        }
}
`
	hostsTemplate = `127.0.0.1       localhost.localdomain   localhost {{.ID}}
::1             localhost6.localdomain6 localhost6
# The following lines are desirable for IPv6 capable hosts
::1     localhost ip6-localhost ip6-loopback
fe00::0 ip6-localnet
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
ff02::3 ip6-allhosts
`
)

type cniConf struct {
	Bridge   string
	NodeName string
	PeerAddr string
}

func (s *service) Containers(ctx context.Context, req *api.ContainersRequest) (*api.ContainersResponse, error) {
	c, err := s.containerd()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	containers, err := c.Containers(ctx, req.Filters...)
	if err != nil {
		return nil, err
	}

	conv, err := s.containersToProto(containers)
	if err != nil {
		return nil, err
	}

	return &api.ContainersResponse{
		Containers: conv,
	}, nil
}

func (s *service) Container(ctx context.Context, req *api.ContainerRequest) (*api.ContainerResponse, error) {
	c, err := s.containerd()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	container, err := c.LoadContainer(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	cont, err := s.containerToProto(container)
	if err != nil {
		return nil, err
	}

	return &api.ContainerResponse{
		Container: cont,
	}, nil
}

func (s *service) CreateContainer(ctx context.Context, req *api.CreateContainerRequest) (*ptypes.Empty, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var (
		opts  []oci.SpecOpts
		cOpts []containerd.NewContainerOpts
	)

	client, err := s.containerd()
	if err != nil {
		return empty, err
	}
	defer client.Close()

	ctx = namespaces.WithNamespace(ctx, s.namespace)

	service := req.Service
	id := req.Service.Name
	if appName := req.Application; appName != "" {
		id = fmt.Sprintf("%s.%s", appName, service.Name)
	}
	snapshotter := defaultSnapshotter
	if service.Snapshotter != "" {
		snapshotter = service.Snapshotter
	}
	image, err := client.GetImage(ctx, service.Image)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return empty, err
		}
		// pull
		img, err := client.Pull(ctx, service.Image, containerd.WithPullUnpack)
		if err != nil {
			return empty, err
		}
		image = img
	}
	unpacked, err := image.IsUnpacked(ctx, snapshotter)
	if err != nil {
		return empty, err
	}
	if !unpacked {
		if err := image.Unpack(ctx, snapshotter); err != nil {
			return empty, err
		}
	}
	opts = append(opts,
		oci.WithImageConfig(image),
		oci.WithHostname(id),
		s.withStellarHosts,
		s.withStellarResolvConf,
		withMounts(service.Mounts),
	)
	if service.Process != nil && service.Process.Env != nil {
		opts = append(opts, oci.WithEnv(service.Process.Env))
	}

	defer func() {
		if err != nil {
			netPath, err := s.getNetPath(id)
			if err != nil {
				logrus.Error(err)
			}
			if netPath != "" {
				if err := deleteNetNS(netPath); err != nil {
					logrus.Errorf("error deleting network namespace (%s) for %s: %s", netPath, id, err)
				}
			}
		}
	}()

	// netns
	netPath, err := s.getNetPath(id)
	if err := createNetNS(netPath); err != nil {
		return empty, err
	}
	opts = append(opts, oci.WithLinuxNamespace(specs.LinuxNamespace{
		Type: specs.NetworkNamespace,
		Path: netPath,
	}))

	netConfData, err := s.getNetworkConf()
	if err != nil {
		return empty, err
	}
	network, err := s.getNetwork(
		gocni.WithConf([]byte(cniLoopbackConf)),
		gocni.WithConf(netConfData),
	)
	if err != nil {
		return empty, err
	}

	netResult, err := network.Setup(id, netPath)
	if err != nil {
		return empty, err
	}

	logrus.Debugf("node.createcontainer: cni result %+v", netResult)

	// TODO check result for length
	ip := netResult.Interfaces[defaultIfName].IPConfigs[0].IP.String()

	cOpts = append(cOpts,
		containerd.WithContainerLabels(convertLabels(service.Labels)),
		containerd.WithContainerExtension(stellar.StellarServiceExtension, req.Service),
	)
	if service.Runtime != "" {
		cOpts = append(cOpts, containerd.WithRuntime(service.Runtime, nil))
	}
	cOpts = append(cOpts,
		containerd.WithImage(image),
		containerd.WithSnapshotter(snapshotter),
		containerd.WithNewSnapshot(id, image),
		containerd.WithNewSpec(opts...),
		containerd.WithContainerLabels(map[string]string{
			stellar.StellarApplicationLabel: req.Application,
			stellar.StellarNetworkLabel:     "true",
		}),
		restart.WithStatus(containerd.Running),
	)

	container, err := client.NewContainer(ctx, id, cOpts...)
	if err != nil {
		return empty, err
	}

	// log
	logPath, err := s.logPath(id)
	if err != nil {
		return empty, err
	}
	task, err := container.NewTask(ctx, cio.LogFile(logPath))
	if err != nil {
		return empty, err
	}
	if err := task.Start(ctx); err != nil {
		return empty, err
	}

	// update dns
	c, err := s.client()
	if err != nil {
		return empty, err
	}
	defer c.Close()

	// TODO: make domain configurable
	var records []*nameserverapi.Record
	recordName := id + ".stellar"
	records = append(records, &nameserverapi.Record{
		Type:  nameserverapi.RecordType_A,
		Name:  recordName,
		Value: ip,
	})
	records = append(records, &nameserverapi.Record{
		Type:  nameserverapi.RecordType_TXT,
		Name:  recordName,
		Value: fmt.Sprintf("node=%s; updated=%s", s.nodeName(), time.Now().Format(time.RFC3339)),
	})
	// endpoints
	for _, ep := range service.Endpoints {
		o := &types.SRVOptions{
			Service:  ep.Service,
			Protocol: strings.ToLower(ep.Protocol.String()),
			Priority: uint16(0),
			Weight:   uint16(0),
			Port:     uint16(ep.Port),
		}
		opts, err := typeurl.MarshalAny(o)
		if err != nil {
			return empty, err
		}
		records = append(records, &nameserverapi.Record{
			Type:    nameserverapi.RecordType_SRV,
			Name:    recordName,
			Value:   recordName,
			Options: opts,
		})
	}
	if err := c.Nameserver().CreateRecords(recordName, records); err != nil {
		return empty, err
	}
	return empty, nil
}

func (s *service) DeleteContainer(ctx context.Context, req *api.DeleteContainerRequest) (*ptypes.Empty, error) {
	c, err := s.containerd()
	if err != nil {
		return empty, err
	}
	defer c.Close()

	client, err := s.client()
	if err != nil {
		return empty, err
	}
	defer client.Close()

	wg := &sync.WaitGroup{}
	container, err := c.LoadContainer(ctx, req.ID)
	if err != nil {
		return empty, err
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return empty, err
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if task != nil {
			wait, err := task.Wait(ctx)
			if err != nil {
				logrus.Errorf("error waiting on task: %s", err)
				return
			}
			if err := task.Kill(ctx, unix.SIGTERM, containerd.WithKillAll); err != nil {
				logrus.Warnf("error killing container task: %s", err)
			}
			select {
			case <-wait:
				task.Delete(ctx)
				return
			case <-time.After(5 * time.Second):
				if err := task.Kill(ctx, unix.SIGKILL, containerd.WithKillAll); err != nil {
					logrus.Warnf("error force killing container task: %s", err)
				}
				return
			}
		}
	}()

	wg.Wait()

	networkEnabled, err := s.networkEnabled(ctx, container)
	if err != nil {
		return empty, err
	}

	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return empty, err
	}

	// delete data dir
	cpath, err := s.getContainerDataDir(req.ID)
	if err != nil {
		return empty, err
	}
	_ = os.RemoveAll(cpath)

	if networkEnabled {
		netPath, err := s.getNetPath(req.ID)
		if err != nil {
			return empty, err
		}
		netConfData, err := s.getNetworkConf()
		if err != nil {
			return empty, err
		}
		network, err := s.getNetwork(gocni.WithConf(netConfData))
		if err != nil {
			return empty, err
		}

		if err := network.Remove(req.ID, netPath); err != nil {
			return empty, err
		}

		if err := deleteNetNS(netPath); err != nil {
			logrus.Errorf("error deleting network namespace (%s) for %s: %s", netPath, req.ID, err)
		}
	}

	return empty, nil
}

func (s *service) networkEnabled(ctx context.Context, container containerd.Container) (bool, error) {
	labels, err := container.Labels(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := labels[stellar.StellarNetworkLabel]; ok {
		return true, nil
	}

	return false, nil
}

func (s *service) containersToProto(containers []containerd.Container) ([]*api.Container, error) {
	var c []*api.Container
	for _, container := range containers {
		conv, err := s.containerToProto(container)
		if err != nil {
			return nil, err
		}

		c = append(c, conv)
	}

	return c, nil
}

func (s *service) containerToProto(container containerd.Container) (*api.Container, error) {
	info, err := container.Info(context.Background())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pid := uint32(0)

	// attempt to find task pid
	task, _ := container.Task(ctx, nil)
	if task != nil {
		pid = task.Pid()
	}

	exts, err := container.Extensions(ctx)
	if err != nil {
		return nil, err
	}

	ctr := &api.Container{
		ID:     container.ID(),
		Image:  info.Image,
		Labels: info.Labels,
		Spec: &ptypes.Any{
			TypeUrl: info.Spec.TypeUrl,
			Value:   info.Spec.Value,
		},
		Snapshotter: info.Snapshotter,
		Task: &api.Container_Task{
			Pid: pid,
		},
		Runtime:    info.Runtime.Name,
		Extensions: make(map[string]*ptypes.Any),
	}
	for k, ext := range exts {
		ctr.Extensions[k] = &ext
	}

	return ctr, nil
}

func (s *service) getNetworkConf() ([]byte, error) {
	t := template.New("cni")
	tmpl, err := t.Parse(cniConfTemplate)
	if err != nil {
		return nil, err
	}
	peerAddr, err := s.peerAddr()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, cniConf{NodeName: s.nodeName(), PeerAddr: peerAddr, Bridge: s.bridge}); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (s *service) getNetwork(opts ...gocni.CNIOpt) (gocni.CNI, error) {
	network, err := gocni.New(
		gocni.WithPluginDir([]string{"/opt/containerd/bin", "/opt/cni/bin"}))
	if err != nil {
		return nil, err
	}
	// network confs
	if err := network.Load(opts...); err != nil {
		return nil, err
	}
	return network, nil
}

func (s *service) getContainerDataDir(id string) (string, error) {
	p := filepath.Join(s.dataDir, "containers", id)
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", err
	}

	return p, nil
}

func (s *service) logPath(id string) (string, error) {
	cpath, err := s.getContainerDataDir(id)
	if err != nil {
		return "", err
	}
	lp := filepath.Join(cpath, "log")
	if err := os.MkdirAll(filepath.Dir(lp), 0755); err != nil {
		return "", err
	}

	return lp, nil
}

func (s *service) getNetPath(id string) (string, error) {
	netPath := filepath.Join(s.stateDir, "ns", id, "net")
	if err := os.MkdirAll(filepath.Dir(netPath), 0700); err != nil {
		return "", err
	}

	return netPath, nil
}

func createNetNS(path string) error {
	cmd := exec.Command("stellar", "network", "create", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: unix.CLONE_NEWNET,
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, string(out))

	}
	return nil
}

func deleteNetNS(path string) error {
	cmd := exec.Command("stellar", "network", "delete", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: unix.CLONE_NEWNET,
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, string(out))

	}
	return nil
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

func (s *service) withStellarHosts(ctx context.Context, _ oci.Client, c *containers.Container, spec *oci.Spec) error {
	cpath, err := s.getContainerDataDir(c.ID)
	if err != nil {
		return err
	}
	t := template.New("hosts")
	tmpl, err := t.Parse(hostsTemplate)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, struct{ ID string }{ID: c.ID}); err != nil {
		return err
	}
	hostsPath := filepath.Join(cpath, "hosts")
	f, err := os.Create(hostsPath)
	if err != nil {
		return err
	}
	f.Write(b.Bytes())
	f.Close()

	spec.Mounts = append(spec.Mounts, specs.Mount{
		Type:        "bind",
		Source:      hostsPath,
		Destination: "/etc/hosts",
		Options:     []string{"rbind", "ro"},
	})
	return nil
}

func (s *service) withStellarResolvConf(ctx context.Context, _ oci.Client, c *containers.Container, spec *oci.Spec) error {
	spec.Mounts = append(spec.Mounts, specs.Mount{
		Type:        "bind",
		Source:      filepath.Join(s.dataDir, "resolv.conf"),
		Destination: "/etc/resolv.conf",
		Options:     []string{"rbind", "ro"},
	})
	return nil
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
