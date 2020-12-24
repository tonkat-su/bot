package mclookup

import (
	"context"
	"net"
	"strconv"
)

// Server is host port
type Server struct {
	Host string
	Port uint16
}

func (srv *Server) String() string {
	return net.JoinHostPort(srv.Host, strconv.FormatUint(uint64(srv.Port), 10))
}

func ResolveMinecraftHostPort(ctx context.Context, resolver *net.Resolver, server string) (servers []Server, err error) {
	_, addrs, err := resolver.LookupSRV(ctx, "minecraft", "tcp", server)
	if err != nil {
		if e, ok := err.(*net.DNSError); ok {
			if e.IsNotFound {
				_, err = resolver.LookupHost(ctx, server)
				if err != nil {
					if e, ok := err.(*net.DNSError); ok {
						if e.IsNotFound {
							return nil, nil
						}
					}
					return nil, err
				}
				return []Server{{Host: server, Port: 25565}}, nil
			}
		}
		return nil, err
	}

	servers = make([]Server, len(addrs))
	for i := range servers {
		servers[i] = Server{
			Host: addrs[i].Target,
			Port: addrs[i].Port,
		}
	}
	return servers, nil
}
