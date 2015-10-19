package render_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

var (
	serverHostID     = "host1"
	serverHostNodeID = report.MakeHostNodeID(serverHostID)

	randomIP             = "3.4.5.6"
	randomPort           = "56789"
	randomEndpointNodeID = report.MakeEndpointNodeID(serverHostID, randomIP, randomPort)

	serverIP             = "192.168.1.1"
	serverPort           = "80"
	serverEndpointNodeID = report.MakeEndpointNodeID(serverHostID, serverIP, serverPort)

	containerID     = "a1b2c3d4e5"
	containerIP     = "192.168.0.1"
	containerName   = "foo"
	containerNodeID = report.MakeContainerNodeID(serverHostID, containerID)

	rpt = report.Report{
		Endpoint: report.Topology{
			Nodes: report.Nodes{
				randomEndpointNodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:        randomIP,
					endpoint.Port:        randomPort,
					endpoint.Conntracked: "true",
				}).WithAdjacent(serverEndpointNodeID),

				serverEndpointNodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:        serverIP,
					endpoint.Port:        serverPort,
					endpoint.Conntracked: "true",
				}),
			},
		},
		Container: report.Topology{
			Nodes: report.Nodes{
				containerNodeID: report.MakeNode().WithMetadata(map[string]string{
					docker.ContainerID:    containerID,
					docker.ContainerName:  containerName,
					docker.ContainerIPs:   containerIP,
					docker.ContainerPorts: fmt.Sprintf("%s:%s->%s/tcp", serverIP, serverPort, serverPort),
					report.HostNodeID:     serverHostNodeID,
				}),
			},
		},
		Host: report.Topology{
			Nodes: report.Nodes{
				serverHostNodeID: report.MakeNodeWith(map[string]string{
					"local_networks":  "192.168.0.0/16",
					report.HostNodeID: serverHostNodeID,
				}),
			},
		},
	}

	want = (render.RenderableNodes{
		render.IncomingInternetID: {
			ID:         render.IncomingInternetID,
			LabelMajor: render.TheInternetMajor,
			LabelMinor: render.Incoming,
			Pseudo:     true,
			Node:       report.MakeNode().WithAdjacent(containerID),
			Origins:    report.MakeIDList(randomEndpointNodeID),
		},
		containerID: {
			ID:         containerID,
			LabelMajor: containerName,
			LabelMinor: serverHostID,
			Rank:       "",
			Pseudo:     false,
			Origins:    report.MakeIDList(containerNodeID, serverEndpointNodeID, serverHostNodeID),
			Node:       report.MakeNode(),
		},
	}).Prune()
)

func TestShortLivedInternetNodeConnections(t *testing.T) {
	have := (render.ContainerWithImageNameRenderer.Render(rpt)).Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
