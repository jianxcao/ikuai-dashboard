package v4service

import (
	"bytes"
	"context"
	"io"
	"net/http"
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
