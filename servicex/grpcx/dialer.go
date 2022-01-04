package grpcx

import (
	"context"
	"fmt"
	"net"

	"github.com/oligarch316/go-netx/listenerx/multi"
	"github.com/oligarch316/go-netx/serverx"
	"github.com/oligarch316/go-netx/servicex"
	"google.golang.org/grpc"
)

type hashDialer interface {
	DialHash(multi.SetHash) (net.Conn, error)
	DialContextHash(context.Context, multi.SetHash) (net.Conn, error)
}

// DialSet TODO.
type DialSet interface {
	hashResolver
	hashDialer
}

// DialerOption TODO.
type DialerOption func(*DialerParams)

// DialerParams TODO.
type DialerParams struct {
	Resolver        ResolverParams
	GRPCDialOptions []grpc.DialOption
}

func defaultDialerParams() DialerParams {
	schemeName := servicex.DefaultDialKey

	return DialerParams{
		Resolver: ResolverParams{
			SchemeName:  &schemeName,
			DNSHostName: nil,
		},
		GRPCDialOptions: nil,
	}
}

func (dp DialerParams) buildContextDialer(hDialer hashDialer) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		if local, h, err := rsvParseHash(addr); local {
			if err != nil {
				return nil, fmt.Errorf("grpcx: dialer: %w", err)
			}
			return hDialer.DialContextHash(ctx, h)
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

func (dp DialerParams) build(dialSet DialSet) []grpc.DialOption {
	return append(
		dp.GRPCDialOptions,
		grpc.WithResolvers(dp.Resolver.build(dialSet)...),
		grpc.WithContextDialer(dp.buildContextDialer(dialSet)),
	)
}

// Dialer TODO.
type Dialer struct{ commonOpts []grpc.DialOption }

// LoadDialer TODO.
func LoadDialer(svr *serverx.Server, opts ...DialerOption) (*Dialer, error) {
	dialSet, err := svr.Dialer(ID)
	if err != nil {
		return nil, err
	}
	return NewDialer(dialSet, opts...), nil
}

// NewDialer TODO.
func NewDialer(dialSet DialSet, opts ...DialerOption) *Dialer {
	params := defaultDialerParams()
	for _, opt := range opts {
		opt(&params)
	}
	return &Dialer{commonOpts: params.build(dialSet)}
}

// Dial TODO.
func (d *Dialer) Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(target, append(d.commonOpts, opts...)...)
}

// DialContext TODO.
func (d *Dialer) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target, append(d.commonOpts, opts...)...)
}
