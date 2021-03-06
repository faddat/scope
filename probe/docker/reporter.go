package docker

import (
	"net"

	docker_client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/report"
)

// Keys for use in Node
const (
	ImageID   = "docker_image_id"
	ImageName = "docker_image_name"
)

// Reporter generate Reports containing Container and ContainerImage topologies
type Reporter struct {
	registry Registry
	hostID   string
}

// NewReporter makes a new Reporter
func NewReporter(registry Registry, hostID string) *Reporter {
	return &Reporter{
		registry: registry,
		hostID:   hostID,
	}
}

// Report generates a Report containing Container and ContainerImage topologies
func (r *Reporter) Report() (report.Report, error) {
	localAddrs, err := report.LocalAddresses()
	if err != nil {
		return report.MakeReport(), nil
	}

	result := report.MakeReport()
	result.Container = result.Container.Merge(r.containerTopology(localAddrs))
	result.ContainerImage = result.ContainerImage.Merge(r.containerImageTopology())
	return result, nil
}

func (r *Reporter) containerTopology(localAddrs []net.IP) report.Topology {
	result := report.MakeTopology()
	result.Controls.AddControl(report.Control{
		ID:    StopContainer,
		Human: "Stop",
		Icon:  "fa-stop",
	})
	result.Controls.AddControl(report.Control{
		ID:    StartContainer,
		Human: "Start",
		Icon:  "fa-play",
	})
	result.Controls.AddControl(report.Control{
		ID:    RestartContainer,
		Human: "Restart",
		Icon:  "fa-repeat",
	})
	result.Controls.AddControl(report.Control{
		ID:    PauseContainer,
		Human: "Pause",
		Icon:  "fa-pause",
	})
	result.Controls.AddControl(report.Control{
		ID:    UnpauseContainer,
		Human: "Unpause",
		Icon:  "fa-play",
	})

	r.registry.WalkContainers(func(c Container) {
		nodeID := report.MakeContainerNodeID(r.hostID, c.ID())
		result.AddNode(nodeID, c.GetNode(r.hostID, localAddrs))
	})

	return result
}

func (r *Reporter) containerImageTopology() report.Topology {
	result := report.MakeTopology()

	r.registry.WalkImages(func(image *docker_client.APIImages) {
		nmd := report.MakeNodeWith(map[string]string{
			ImageID: image.ID,
		})
		AddLabels(nmd, image.Labels)

		if len(image.RepoTags) > 0 {
			nmd.Metadata[ImageName] = image.RepoTags[0]
		}

		nodeID := report.MakeContainerNodeID(r.hostID, image.ID)
		result.AddNode(nodeID, nmd)
	})

	return result
}
