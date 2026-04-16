package docker

type ContainerSpec struct {
	Image        string
	ExposedPorts []string
	ReadOnlyRoot bool
	NetworkMode  string
	Seccomp      string
}

func DefaultCoreSpec() ContainerSpec {
	return ContainerSpec{
		Image:        "honeytrap-core:phase1",
		ExposedPorts: []string{"2222/tcp", "8080/tcp", "2121/tcp", "9161/udp"},
		ReadOnlyRoot: true,
		NetworkMode:  "bridge",
		Seccomp:      "default",
	}
}
