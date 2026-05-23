package service

import (
	"context"
	"errors"
	"testing"

	ikuaiapi "github.com/zy84338719/ikuai-api"
)

func TestRetryV3CallReloginsOnceWhenAuthenticationExpired(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	relogins := 0

	err := retryV3CallWithRelogin(ctx, func() error {
		attempts++
		if attempts == 1 {
			return ikuaiapi.NewSDKError(ikuaiapi.ErrCodeRequestFailed, "no login authentication", nil)
		}
		return nil
	}, func(context.Context) error {
		relogins++
		return nil
	})

	if err != nil {
		t.Fatalf("retryV3CallWithRelogin() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if relogins != 1 {
		t.Fatalf("relogins = %d, want 1", relogins)
	}
}

func TestRetryV3CallDoesNotReloginForCanceledContext(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	relogins := 0

	err := retryV3CallWithRelogin(ctx, func() error {
		attempts++
		return ikuaiapi.NewSDKError(ikuaiapi.ErrCodeRequestFailed, "failed to send request", context.Canceled)
	}, func(context.Context) error {
		relogins++
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("retryV3CallWithRelogin() error = %v, want context.Canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
	if relogins != 0 {
		t.Fatalf("relogins = %d, want 0", relogins)
	}
}
