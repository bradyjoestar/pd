package watch

import (
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"context"
	"reflect"
	"runtime"
	"strings"
)

// isHaltErr returns true if the given error and context indicate no forward
// progress can be made, even after reconnecting.
func isHaltErr(ctx context.Context, err error) bool {
	if ctx != nil && ctx.Err() != nil {
		return true
	}
	if err == nil {
		return false
	}
	ev, _ := status.FromError(err)
	// Unavailable codes mean the system will be right back.
	// (e.g., can't connect, lost leader)
	// Treat Internal codes as if something failed, leaving the
	// system in an inconsistent state, but retrying could make progress.
	// (e.g., failed in middle of send, corrupted frame)
	// TODO: are permanent Internal errors possible from grpc?
	return ev.Code() != codes.Unavailable && ev.Code() != codes.Internal
}

func toErr(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	err = rpctypes.Error(err)
	if _, ok := err.(rpctypes.EtcdError); ok {
		return err
	}
	if ev, ok := status.FromError(err); ok {
		code := ev.Code()
		switch code {
		case codes.DeadlineExceeded:
			fallthrough
		case codes.Canceled:
			if ctx.Err() != nil {
				err = ctx.Err()
			}
		case codes.Unavailable:
		case codes.FailedPrecondition:
			err = grpc.ErrClientConnClosing
		}
	}
	return err
}

// isUnavailableErr returns true if the given error is an unavailable error
func isUnavailableErr(ctx context.Context, err error) bool {
	if ctx != nil && ctx.Err() != nil {
		return false
	}
	if err == nil {
		return false
	}
	ev, _ := status.FromError(err)
	// Unavailable codes mean the system will be right back.
	// (e.g., can't connect, lost leader)
	return ev.Code() == codes.Unavailable
}

// Check if the provided function is being called in the op options.
func isOpFuncCalled(op string, opts []OpOption) bool {
	for _, opt := range opts {
		v := reflect.ValueOf(opt)
		if v.Kind() == reflect.Func {
			if opFunc := runtime.FuncForPC(v.Pointer()); opFunc != nil {
				if strings.Contains(opFunc.Name(), op) {
					return true
				}
			}
		}
	}
	return false
}

