package grpcx

import (
	"context"
	"fmt"
	"net"

	"github.com/oligarch316/go-netx/internal/servicex"
	"github.com/oligarch316/go-netx/multi/addrsort"
	"github.com/oligarch316/go-netx/serverx"
	"google.golang.org/grpc"
)

// DialerParams TODO.
type DialerParams struct {
	Resolver        ResolverParams
	GRPCDialOptions []grpc.DialOption
}

func defaultDialerParams() DialerParams {
	var (
		key = servicex.DefaultKey
		cmp = addrsort.ByPriorityNetwork(servicex.DefaultNetworkPriority...) // TODO: have servicex export a cmp directly
	)

	return DialerParams{
		Resolver: ResolverParams{
			SchemeName:   &key,
			DNSHostName:  nil,
			AddressOrder: addrsort.CompareList{cmp},
		},
		GRPCDialOptions: nil,
	}
}

func (dp DialerParams) buildContextDialer(dSet *serverx.DialSet) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		if local, h, err := rsvParseHash(addr); local {
			if err != nil {
				return nil, fmt.Errorf("grpcx: dialer: %w", err)
			}
			return dSet.DialContext(ctx, h)
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
	}
}

func (dp DialerParams) build(dSet *serverx.DialSet) []grpc.DialOption {
	return append(
		dp.GRPCDialOptions,
		grpc.WithResolvers(dp.Resolver.build(dSet)...),
		grpc.WithContextDialer(dp.buildContextDialer(dSet)),
	)
}

// Dialer TODO.
type Dialer struct{ commonOpts []grpc.DialOption }

// LoadDialer TODO.
func LoadDialer(svr serverx.Server, opts ...DialerOption) (*Dialer, error) {
	set, err := svr.DialSet(ID)
	if err != nil {
		return nil, err
	}
	return NewDialer(set, opts...), nil
}

// NewDialer TODO.
func NewDialer(set *serverx.DialSet, opts ...DialerOption) *Dialer {
	params := defaultDialerParams()
	for _, opt := range opts {
		opt(&params)
	}
	return &Dialer{commonOpts: params.build(set)}
}

// Dial TODO.
func (d *Dialer) Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(target, append(d.commonOpts, opts...)...)
}

// DialContext TODO.
func (d *Dialer) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target, append(d.commonOpts, opts...)...)
}
