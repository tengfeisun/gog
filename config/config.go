package config

import (
	"flag"
	"net"
)

// Config describes the config of the system.
type Config struct {
	// Net should be tcp4 or tcp6.
	Net string
	// AddrStr is the local address string.
	AddrStr string
	// LocalTCPAddr is TCP address parsed from
	// Net and AddrStr.
	LocalTCPAddr *net.TCPAddr
	// Fanin is the nodes we allow to have
	// us in their active view.
	Fanin int
	// Fanout is the number of nodes in our
	// active view.
	Fanout int
	// AViewSize is the size of the active view.
	// It is equal to Fanout.
	AViewSize int
	// PViewSize is the size of the passive view.
	PViewSize int
	// Ka is the number of nodes to choose from active view
	// when shuffling views.
	Ka int
	// Kp is the number of nodes to choose from passive view
	// when shuffling views.
	Kp int
	// Active Random Walk Length.
	ARWL int
	// Passive Random Walk Length.
	PRWL int
}

func ParseConfig() (*Config, error) {
	cfg := new(Config)

	flag.StringVar(&cfg.Net, "net", "tcp", "The network protocol")
	flag.StringVar(&cfg.AddrStr, "addr", "localhost:8424", "The address the agent listens on")

	flag.IntVar(&cfg.Fanin, "fanin", 5, "The fan-in")
	flag.IntVar(&cfg.Fanout, "fanout", 5, "The fan-out")

	flag.IntVar(&cfg.AViewSize, "active_view_size", 5, "The size of the active view")
	flag.IntVar(&cfg.PViewSize, "passive_view_size", 5, "The size of the passive view")

	flag.IntVar(&cfg.Ka, "ka", 1, "The number of active nodes to shuffle")
	flag.IntVar(&cfg.Kp, "kp", 3, "The number of passive nodes to shuffle")

	flag.IntVar(&cfg.ARWL, "arwl", 5, "The active random walk length")
	flag.IntVar(&cfg.PRWL, "prwl", 5, "The passive random walk length")

	flag.Parse()

	// TODO check config.
	tcpAddr, err := net.ResolveTCPAddr(cfg.Net, cfg.AddrStr)
	if err != nil {
		return nil, err
	}
	cfg.LocalTCPAddr = tcpAddr
	return cfg, nil
}
