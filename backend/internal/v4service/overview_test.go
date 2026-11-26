package v4service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func okJSON(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestGetOverviewParsesRealFirmwarePayload(t *testing.T) {
	client := ikuaiv4.NewClientWithHTTP("https://192.0.2.1", "tok", &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			switch r.URL.Path {
			case "/api/v4.0/monitoring/system":
				return okJSON(`{"code":0,"message":"Success","results":{"sysinfo":{"cpu":["0.50%","0.00%"],"cputemp":[],"freq":["3187","3187"],"gwid":"gw1","hostname":"iKuai","ip_addr":"192.168.9.1","link_status":0,"memory":{"total":1902852,"available":1144556,"free":1151284,"cached":343116,"buffers":71184,"used":"39%"},"online_user":{"count":3,"count_2g":1,"count_5g":1,"count_wired":1,"count_wireless":2},"stream":{"connect_num":11,"tcp_connect_num":4,"udp_connect_num":4,"icmp_connect_num":2,"upload":1024,"download":2048,"total_up":1082047046,"total_down":11440414},"uptime":3158145,"verinfo":{"verstring":"4.0.314-beta x64 Build202610131658","version":"4.0.314","build_date":202610131658,"arch":"x86","sysbit":"x64"}}}}`), nil
			case "/api/v4.0/monitoring/interfaces-status":
				return okJSON(`{"code":0,"message":"Success","results":{"iface_check":[{"id":1,"interface":"wan1","ip_addr":"198.51.100.10","gateway":"198.51.100.1","internet":"静态IP","result":"success","errmsg":"线路检测成功"}],"iface_stream":[{"interface":"lan1","ip_addr":"192.168.9.1","connect_num":"--","upload":0,"download":0,"total_up":100,"total_down":200},{"interface":"wan1","ip_addr":"198.51.100.10","connect_num":"11","upload":1024,"download":2048,"total_up":300,"total_down":400}]}}`), nil
			case "/api/v4.0/monitoring/interfaces-traffic":
				return okJSON(`{"code":0,"message":"Success","results":{"wans_stat_history":[{"timestamp":1795490700,"interface":"wan1","avg_upload":"0.50","avg_download":"1.20","avg_rtt":4.99,"max_rtt":6.79,"min_rtt":4.43,"drop_rate":"0.00"}]}}`), nil
			case "/api/v4.0/interfaces/physical":
				return okJSON(`{"code":0,"message":"Success","results":{"ether_info":{"eth0":{"driver":"virtio_net","type":"TP","mac":"bc:24:11:18:79:dc","link":1,"speed":1000,"duplex":2,"model":"Virtio","interface":"wan1"}}}}`), nil
			default:
				return okJSON(`{"code":0,"message":"Success","results":{"data":[],"total":0}}`), nil
			}
		}),
	})

	router := config.RouterConfig{ID: "r1", Name: "Test V4", URL: "https://192.0.2.1", Token: "tok"}
	ov, err := GetOverview(context.Background(), router, client)
	if err != nil {
		t.Fatalf("GetOverview failed: %v", err)
	}
	if ov.Router.Hostname != "iKuai" || ov.Router.Version != "4.0.314" || ov.CPU.AvgPercent != 0.5 || ov.Memory.UsedPercent != 39 {
		t.Fatalf("unexpected overview summary: %+v", ov)
	}
	if len(ov.WanChecks) != 1 || ov.WanChecks[0].AvgRTTMs != 4.99 {
		t.Fatalf("expected wan check with RTT 4.99, got %+v", ov.WanChecks)
	}
	if len(ov.PhysicalPorts) != 1 || ov.PhysicalPorts[0].Name != "eth0" || ov.PhysicalPorts[0].Interface != "wan1" {
		t.Fatalf("expected eth0 physical port, got %+v", ov.PhysicalPorts)
	}
	if len(ov.InterfaceTraffic) != 2 || ov.InterfaceTraffic[0].Connections != 0 || ov.InterfaceTraffic[1].Connections != 11 {
		t.Fatalf("unexpected interface traffic: %+v", ov.InterfaceTraffic)
	}

	legacy := ToLegacyInterfaceData(ov)
	if legacy.Summary.UploadSpeed != 1024 || legacy.Summary.DownloadSpeed != 2048 || len(legacy.WanStatus) != 1 {
		t.Fatalf("unexpected legacy interface mapping: %+v", legacy)
	}
}

func TestGetOverviewCachesPhysicalAndTrafficBriefly(t *testing.T) {
	var mu sync.Mutex
	counts := map[string]int{}

	client := ikuaiv4.NewClientWithHTTP("https://192.0.2.2", "tok", &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			mu.Lock()
			counts[r.URL.Path]++
			mu.Unlock()
			switch r.URL.Path {
			case "/api/v4.0/monitoring/system":
				return okJSON(`{"code":0,"message":"Success","results":{"sysinfo":{"cpu":["1.00%"],"memory":{"total":1000,"used":"50%"},"stream":{"upload":10,"download":20},"online_user":{"count":1}}}}`), nil
			case "/api/v4.0/monitoring/interfaces-status":
				return okJSON(`{"code":0,"message":"Success","results":{"iface_check":[],"iface_stream":[]}}`), nil
			case "/api/v4.0/monitoring/interfaces-traffic":
				return okJSON(`{"code":0,"message":"Success","results":{"wans_stat_history":[{"timestamp":1795490700,"interface":"wan1","avg_upload":"1","avg_download":"2","avg_rtt":5}]}}`), nil
			case "/api/v4.0/interfaces/physical":
				return okJSON(`{"code":0,"message":"Success","results":{"ether_info":{"eth0":{"link":1,"speed":1000}}}}`), nil
			default:
				return okJSON(`{"code":0,"message":"Success","results":{"data":[]}}`), nil
			}
		}),
	})

	router := config.RouterConfig{ID: "cache-test", Name: "Cache Test", URL: "https://192.0.2.2", Token: "tok"}
	if _, err := GetOverview(context.Background(), router, client); err != nil {
		t.Fatalf("first GetOverview failed: %v", err)
	}
	if _, err := GetOverview(context.Background(), router, client); err != nil {
		t.Fatalf("second GetOverview failed: %v", err)
	}

	if counts["/api/v4.0/monitoring/system"] != 2 {
		t.Fatalf("expected 2 live /monitoring/system calls, got %d", counts["/api/v4.0/monitoring/system"])
	}
	if counts["/api/v4.0/monitoring/interfaces-status"] != 2 {
		t.Fatalf("expected 2 live /monitoring/interfaces-status calls, got %d", counts["/api/v4.0/monitoring/interfaces-status"])
	}
	if counts["/api/v4.0/monitoring/interfaces-traffic"] != 1 {
		t.Fatalf("expected /monitoring/interfaces-traffic to be cached once, got %d", counts["/api/v4.0/monitoring/interfaces-traffic"])
	}
	if counts["/api/v4.0/interfaces/physical"] != 1 {
		t.Fatalf("expected /interfaces/physical to be cached once, got %d", counts["/api/v4.0/interfaces/physical"])
	}
}

func TestGetOverviewHandlesEmptyFreqAndV40317TrafficKeys(t *testing.T) {
	client := ikuaiv4.NewClientWithHTTP("https://192.0.2.3", "tok", &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			switch r.URL.Path {
			case "/api/v4.0/monitoring/system":
				return okJSON(`{"code":0,"message":"Success","results":{"sysinfo":{"cpu":["6.73%","1.01%","9.90%","7.07%","7.84%"],"cputemp":[],"freq":["","","",""],"gwid":"gw1","hostname":"iKuai","ip_addr":"192.0.2.1","link_status":0,"memory":{"total":495856,"available":200068,"free":169168,"cached":41108,"buffers":0,"used":"59%"},"online_user":{"count":36,"count_2g":0,"count_5g":0,"count_wired":36,"count_wireless":0},"stream":{"connect_num":1410,"tcp_connect_num":867,"udp_connect_num":504,"icmp_connect_num":15,"upload":274147,"download":51068,"total_up":72092856000,"total_down":60227197523},"uptime":24971,"verinfo":{"verstring":"4.0.317 x64 Build202611101439","version":"4.0.317","arch":"arm","sysbit":"x64"}}}}`), nil
			case "/api/v4.0/monitoring/interfaces-status":
				return okJSON(`{"code":0,"message":"Success","results":{"iface_check":[{"id":1,"interface":"wan1","parent_interface":"wan1","ip_addr":"203.0.113.10","gateway":"203.0.113.1","internet":"PPPOE","result":"success","errmsg":"线路检测成功"}],"iface_stream":[{"interface":"lan1","ip_addr":"192.0.2.1","connect_num":"--","upload":100,"download":200,"total_up":300,"total_down":400}]}}`), nil
			case "/api/v4.0/monitoring/interfaces-traffic":
				return okJSON(`{"code":0,"message":"Success","results":{"wans_stat_history":[{"timestamp":1795658220,"interface":"wan1","upload":"476.69","download":"18.41","rtt":6.61,"drop_rate":0},{"timestamp":1795658220,"interface":"wan2","upload":"1126.19","download":"46.42","rtt":6.07,"drop_rate":0}]}}`), nil
			case "/api/v4.0/interfaces/physical":
				return okJSON(`{"code":0,"message":"Success","results":{"ether_info":{"veth1":{"link":1,"speed":1000,"interface":"wan1"}}}}`), nil
			default:
				return okJSON(`{"code":0,"message":"Success","results":{"data":[]}}`), nil
			}
		}),
	})

	router := config.RouterConfig{ID: "arm-40317", Name: "ARM V4.0.317", URL: "https://192.0.2.3", Token: "tok"}
	ov, err := GetOverview(context.Background(), router, client)
	if err != nil {
		t.Fatalf("GetOverview failed: %v", err)
	}
	if ov.CPU.AvgPercent != 6.73 || len(ov.CPU.Cores) != 4 || ov.CPU.Cores[0] != 1.01 {
		t.Fatalf("expected 4 core metrics plus 6.73 avg, got %+v", ov.CPU)
	}
	if ov.CPU.FreqMHz == nil || ov.CPU.Temps == nil {
		t.Fatalf("expected non-nil empty slices for FreqMHz and Temps, got %+v", ov.CPU)
	}
	raw, err := json.Marshal(ov)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	jsonStr := string(raw)
	if strings.Contains(jsonStr, `"freq_mhz":null`) || strings.Contains(jsonStr, `"temps":null`) || strings.Contains(jsonStr, `"cores":null`) {
		t.Fatalf("serialized overview must not contain null CPU arrays: %s", jsonStr)
	}
	if len(ov.WanHistory) != 2 || ov.WanHistory[0].AvgDownloadKBps != 18.41 || ov.WanHistory[0].AvgUploadKBps != 476.69 {
		t.Fatalf("expected v4.0.317 upload/download fallback parsed, got %+v", ov.WanHistory)
	}
	if len(ov.WanChecks) != 1 || ov.WanChecks[0].AvgRTTMs != 6.61 {
		t.Fatalf("expected WanCheck AvgRTTMs 6.61 from rtt fallback, got %+v", ov.WanChecks)
	}
}

func TestGetInterfaceOptions(t *testing.T) {
	mockOpts, err := GetInterfaceOptions(context.Background(), config.RouterConfig{ID: "hq-v4", Mock: true}, nil)
	if err != nil {
		t.Fatalf("mock GetInterfaceOptions failed: %v", err)
	}
	if len(mockOpts.Wan) != 2 || len(mockOpts.Lan) == 0 || len(mockOpts.Physical) != 4 {
		t.Fatalf("unexpected mock options: %+v", mockOpts)
	}

	client := ikuaiv4.NewClientWithHTTP("https://192.0.2.4", "tok", &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			switch r.URL.Path {
			case "/api/v4.0/interfaces/wan-config":
				return okJSON(`{"code":0,"message":"Success","results":{"data":[{"id":1,"tagname":"wan1","internet":2,"ip_mask":"192.0.2.250/24","pppoe_ip_addr":"203.0.113.10","comment":"主线"},{"id":2,"tagname":"wan2","internet":1,"dhcp_ip_addr":"198.51.100.20"}]}}`), nil
			case "/api/v4.0/monitoring/interfaces-status":
				return okJSON(`{"code":0,"message":"Success","results":{"iface_check":[{"interface":"wan1","ip_addr":"203.0.113.11","internet":"PPPOE"},{"interface":"adsl1","ip_addr":"203.0.113.99","internet":"PPPOE"}],"iface_stream":[{"interface":"lan1","ip_addr":"192.0.2.1"}]}}`), nil
			case "/api/v4.0/interfaces/wan-vlan-config":
				return okJSON(`{"code":0,"message":"Success","results":{"vlan_data":[],"vlan_total":0}}`), nil
			case "/api/v4.0/interfaces/lan-config":
				return okJSON(`{"code":0,"message":"Success","results":{"data":[{"id":1,"tagname":"lan1","ip_mask":"192.0.2.1\/255.255.255.0"}]}}`), nil
			case "/api/v4.0/network/vlan":
				return okJSON(`{"code":0,"message":"Success","results":{"data":[{"id":1,"vlan_name":"vlan100","ip_addr":"198.51.100.1","comment":"IoT"}]}}`), nil
			case "/api/v4.0/interfaces/physical":
				return okJSON(`{"code":0,"message":"Success","results":{"ether_info":{"veth1":{"link":1,"speed":1000,"interface":"wan1"},"veth2":{"link":0,"speed":0,"interface":"Null"}}}}`), nil
			default:
				return okJSON(`{"code":0,"message":"Success","results":{"data":[]}}`), nil
			}
		}),
	})

	opts, err := GetInterfaceOptions(context.Background(), config.RouterConfig{ID: "real-v4", URL: "https://192.0.2.4", Token: "tok"}, client)
	if err != nil {
		t.Fatalf("real GetInterfaceOptions failed: %v", err)
	}
	if len(opts.Wan) != 3 || opts.Wan[0].Name != "wan1" || opts.Wan[0].Proto != "PPPoE" || opts.Wan[0].IP != "203.0.113.11" || opts.Wan[2].Name != "adsl1" {
		t.Fatalf("unexpected WAN options: %+v", opts.Wan)
	}
	if len(opts.Lan) != 2 || opts.Lan[0].Name != "lan1" || opts.Lan[0].IP != "192.0.2.1/255.255.255.0" || opts.Lan[1].Name != "vlan100" {
		t.Fatalf("unexpected LAN options: %+v", opts.Lan)
	}
	if len(opts.Physical) != 2 || opts.Physical[0].Name != "veth1" {
		t.Fatalf("unexpected physical options: %+v", opts.Physical)
	}
}

func TestGetLoadHistoryUsesRouterClockByDefault(t *testing.T) {
	var capturedQuery string
	client := ikuaiv4.NewClientWithHTTP("https://192.0.2.5", "tok", &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			capturedQuery = r.URL.RawQuery
			return okJSON(`{"code":0,"message":"Success","results":{"cpu":[{"timestamp":1795661820,"cpu":6}]}}`), nil
		}),
	})

	router := config.RouterConfig{ID: "load-test", Name: "Load Test", URL: "https://192.0.2.5", Token: "tok"}
	res, err := GetLoadHistory(context.Background(), router, client, "cpu", "hour", "avg")
	if err != nil {
		t.Fatalf("GetLoadHistory failed: %v", err)
	}
	if !strings.Contains(capturedQuery, "datetype=hour") || !strings.Contains(capturedQuery, "math=avg") {
		t.Fatalf("expected datetype=hour and math=avg in query, got %q", capturedQuery)
	}
	if strings.Contains(capturedQuery, "start_time=") || strings.Contains(capturedQuery, "end_time=") {
		t.Fatalf("expected no host start_time/end_time on primary load query, got %q", capturedQuery)
	}
	if len(res.Series) != 1 || res.Metric != "cpu" || res.TimeRange != "hour" {
		t.Fatalf("unexpected load history response: %+v", res)
	}
}
