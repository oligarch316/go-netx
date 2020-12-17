package grpcx

import (
	"context"
	"net"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/multi"
	"github.com/oligarch316/go-netx/serverx"
	"google.golang.org/grpc"
)

func buildContextDialerOption(md multi.Dialer) grpc.DialOption {
	return grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		if local, h, err := rsvParseHash(addr); local {
			if err != nil {
				return nil, Error{ Component: "dialer", err: err }
			}
			return md.DialContext(ctx, h)
		}

		// TODO: Manage network string better than a hardcoded "tcp"
		//
		// Both the interface and implementation for custom dialers limit our
		// ability here to do the correct thing in terms of network. Given that
		// stock dialing behavior deals only with "tcp" and "unix" networks,
		// hardcoding "tcp" here only precludes dialing to a unix socket on the
		// same machine but not one used by this same app. The use case for
		// running multiple server processes on the same machine that may want
		// to dial themselves or dial each other is esoteric enough to be the
		// lesser of all evils.
		//
		// Related issue:                    https://github.com/grpc/grpc-go/issues/3990
		// Better-but-not-great related fix: https://github.com/grpc/grpc-go/pull/4021/

		return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	})
}

// DialerParams TODO.
type DialerParams struct {
	Resolver        ResolverParams
	GRPCDialOptions []grpc.DialOption
}

func defaultDialerParams() DialerParams {
	var (
		defaultSchemeName = "localapp"
		defaultNetworkPriority = []string{ netx.NetworkInternal, "unix", "tcp" }
	)

	return DialerParams{
		Resolver: ResolverParams{
			SchemeName: &defaultSchemeName,
			DNSHostName: nil,
			NetworkPriority: defaultNetworkPriority,
		},
		GRPCDialOptions: nil,
	}
}

// Dialer TODO.
type Dialer struct {
	identity

	commonOpts []grpc.DialOption
}

// NewDialer TODO.
func NewDialer(s serverx.Server, opts ...DialerOption) (*Dialer, error) {
	md, err := s.Dialer(ID)
	if err != nil {
		return nil, err
	}

	params := defaultDialerParams()
	for _, opt := range opts {
		opt(&params)
	}

	commonOpts := append(
		params.GRPCDialOptions,
		buildResolversOption(md, params.Resolver),
		buildContextDialerOption(md),
	)

	return &Dialer{commonOpts: commonOpts}, nil
}

// Dial TODO.
func (d *Dialer) Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(target, append(d.commonOpts, opts...)...)
}

// DialContext TODO.
func (d *Dialer) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target, append(d.commonOpts, opts...)...)
}
