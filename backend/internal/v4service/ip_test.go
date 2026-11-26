package v4service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

func TestGetWanIPMockMode(t *testing.T) {
	router := config.RouterConfig{ID: "hq-v4", Name: "总部核心网关", Mock: true}

	res, err := GetWanIP(context.Background(), router, nil, "")
	if err != nil {
		t.Fatalf("GetWanIP mock default failed: %v", err)
	}
	if res.Interface != "wan1" || res.IP != "203.0.113.10" {
		t.Fatalf("unexpected default mock wan result: %+v", res)
	}

	res2, err := GetWanIP(context.Background(), router, nil, "wan2")
	if err != nil {
		t.Fatalf("GetWanIP mock wan2 failed: %v", err)
	}
	if res2.Interface != "wan2" || res2.IP != "198.51.100.22" {
		t.Fatalf("unexpected wan2 mock result: %+v", res2)
	}

	_, err = GetWanIP(context.Background(), router, nil, "wan9")
	if !errors.Is(err, ErrWanNotFound) {
		t.Fatalf("expected ErrWanNotFound for wan9, got %v", err)
	}
}

func TestGetWanIPRealFirmwareCases(t *testing.T) {
	router := config.RouterConfig{ID: "main", Name: "主路由", URL: "https://192.0.2.1", Token: "test-token"}

	t.Run("interfaces-status matches wan1 and case-insensitive WAN2", func(t *testing.T) {
		client := ikuaiv4.NewClientWithHTTP("https://192.0.2.1", "test-token", &http.Client{
			Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/api/v4.0/monitoring/interfaces-status" {
					return okJSON(`{"code":0,"message":"Success","results":{"iface_check":[{"id":1,"interface":"wan1","ip_addr":"203.0.113.10","result":"success"},{"id":2,"interface":"wan2","ip_addr":"198.51.100.20","result":"success"}],"iface_stream":[{"interface":"lan1","ip_addr":"192.168.1.1"}]}}`), nil
				}
				return okJSON(`{"code":0,"message":"Success","results":{"data":[]}}`), nil
			}),
		})

		got1, err := GetWanIP(context.Background(), router, client, "wan1")
		if err != nil || got1.IP != "203.0.113.10" {
			t.Fatalf("expected wan1 203.0.113.10, got %+v err=%v", got1, err)
		}
		got2, err := GetWanIP(context.Background(), router, client, "WAN2")
		if err != nil || got2.Interface != "wan2" || got2.IP != "198.51.100.20" {
			t.Fatalf("expected wan2 198.51.100.20, got %+v err=%v", got2, err)
		}
	})

	t.Run("iface_stream fallback without iface_check and ignores lan1", func(t *testing.T) {
		client := ikuaiv4.NewClientWithHTTP("https://192.0.2.1", "test-token", &http.Client{
			Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.Path {
				case "/api/v4.0/monitoring/interfaces-status":
					return okJSON(`{"code":0,"message":"Success","results":{"iface_check":[],"iface_stream":[{"interface":"lan1","ip_addr":"192.168.1.1"},{"interface":"wan2","ip_addr":"198.51.100.33"}]}}`), nil
				default:
					return okJSON(`{"code":0,"message":"Success","results":{"data":[]}}`), nil
				}
			}),
		})

		gotDefault, err := GetWanIP(context.Background(), router, client, "")
		if err != nil || gotDefault.Interface != "wan2" || gotDefault.IP != "198.51.100.33" {
			t.Fatalf("expected wan2 (never lan1), got %+v err=%v", gotDefault, err)
		}
		_, err = GetWanIP(context.Background(), router, client, "lan1")
		if !errors.Is(err, ErrWanNotFound) {
			t.Fatalf("expected ErrWanNotFound when requesting lan1, got %v", err)
		}
	})

	t.Run("pppoe child interface resolves by parent_interface or child name", func(t *testing.T) {
		client := ikuaiv4.NewClientWithHTTP("https://192.0.2.1", "test-token", &http.Client{
			Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
				return okJSON(`{"code":0,"message":"Success","results":{"iface_check":[{"id":1,"interface":"adsl1","parent_interface":"wan1","ip_addr":"203.0.113.77","result":"success"}],"iface_stream":[]}}`), nil
			}),
		})

		byParent, err := GetWanIP(context.Background(), router, client, "wan1")
		if err != nil || byParent.Interface != "adsl1" || byParent.IP != "203.0.113.77" {
			t.Fatalf("expected adsl1 via parent wan1, got %+v err=%v", byParent, err)
		}
		byChild, err := GetWanIP(context.Background(), router, client, "adsl1")
		if err != nil || byChild.IP != "203.0.113.77" {
			t.Fatalf("expected adsl1 directly, got %+v err=%v", byChild, err)
		}
	})

	t.Run("fallback to /interfaces/wan-config strips CIDR suffix", func(t *testing.T) {
		client := ikuaiv4.NewClientWithHTTP("https://192.0.2.1", "test-token", &http.Client{
			Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.Path {
				case "/api/v4.0/monitoring/interfaces-status":
					return okJSON(`{"code":0,"message":"Success","results":{"iface_check":[],"iface_stream":[]}}`), nil
				case "/api/v4.0/interfaces/wan-config":
					return okJSON(`{"code":0,"message":"Success","results":{"data":[{"id":1,"tagname":"wan1","ip_mask":"203.0.113.55/24"}]}}`), nil
				default:
					return okJSON(`{"code":0,"message":"Success","results":{"data":[]}}`), nil
				}
			}),
		})

		got, err := GetWanIP(context.Background(), router, client, "wan1")
		if err != nil || got.IP != "203.0.113.55" {
			t.Fatalf("expected stripped IP 203.0.113.55, got %+v err=%v", got, err)
		}
	})

	t.Run("disconnected wan returns ErrWanNoIP", func(t *testing.T) {
		client := ikuaiv4.NewClientWithHTTP("https://192.0.2.1", "test-token", &http.Client{
			Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.Path {
				case "/api/v4.0/monitoring/interfaces-status":
					return okJSON(`{"code":0,"message":"Success","results":{"iface_check":[{"id":1,"interface":"wan1","ip_addr":"--","result":"fail"}],"iface_stream":[]}}`), nil
				case "/api/v4.0/interfaces/wan-config":
					return okJSON(`{"code":0,"message":"Success","results":{"data":[{"id":1,"tagname":"wan1","ip_mask":"0.0.0.0/0"}]}}`), nil
				default:
					return okJSON(`{"code":0,"message":"Success","results":{"data":[]}}`), nil
				}
			}),
		})

		_, err := GetWanIP(context.Background(), router, client, "wan1")
		if !errors.Is(err, ErrWanNoIP) {
			t.Fatalf("expected ErrWanNoIP, got %v", err)
		}
	})
}
