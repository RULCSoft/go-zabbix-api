package zabbix_test

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	zapi "github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/go-version"
)

var (
	_host string
	_api  *zapi.API
)

func init() {
	rand.Seed(time.Now().UnixNano())

	var err error
	_host, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	_host += "-testing"

	if os.Getenv("TEST_ZABBIX_URL") == "" {
		log.Fatal("Set environment variables TEST_ZABBIX_URL (and optionally TEST_ZABBIX_USER and TEST_ZABBIX_PASSWORD)")
	}
}

func testGetHost() string {
	return _host
}

func testGetAPI(t *testing.T) *zapi.API {
	if _api != nil {
		return _api
	}

	url, username, password := os.Getenv("TEST_ZABBIX_URL"), os.Getenv("TEST_ZABBIX_USER"), os.Getenv("TEST_ZABBIX_PASSWORD")
	_api = zapi.NewAPI(url)
	_api.SetClient(http.DefaultClient)
	v := os.Getenv("TEST_ZABBIX_VERBOSE")
	if v != "" && v != "0" {
		_api.Logger = log.New(os.Stderr, "[zabbix] ", 0)
	}

	if username != "" {
		auth, err := _api.Login(username, password)
		if err != nil {
			t.Fatal(err)
		}
		if auth == "" {
			t.Fatal("Login failed")
		}
	}

	return _api
}

func compVersion(t *testing.T, comparedVersion string) (int, string) {
	api := testGetAPI(t)
	serverVersion, err := api.Version()
	if err != nil {
		t.Fatal(err)
	}
	sVersion, _ := version.NewVersion(serverVersion)
	cVersion, _ := version.NewVersion(comparedVersion)
	return sVersion.Compare(cVersion), serverVersion
}

func isVersionLessThan(t *testing.T, comparedVersion string) (bool, string) {
	comp, serverVersion := compVersion(t, comparedVersion)
	return comp < 0, serverVersion
}

func isVersionGreaterThanOrEqual(t *testing.T, comparedVersion string) (bool, string) {
	comp, serverVersion := compVersion(t, comparedVersion)
	return comp >= 0, serverVersion
}

func skipTestIfVersionGreaterThanOrEqual(t *testing.T, comparedVersion, msg string) {
	if compGreaterThanOrEqual, serverVersion := isVersionGreaterThanOrEqual(t, comparedVersion); compGreaterThanOrEqual {
		t.Skipf("Zabbix version %s is greater than or equal to %s which %s, skipping test.", serverVersion, comparedVersion, msg)
	}
}

func skipTestIfVersionLessThan(t *testing.T, comparedVersion, msg string) {
	if compGreaterThanOrEqual, serverVersion := isVersionLessThan(t, comparedVersion); compGreaterThanOrEqual {
		t.Skipf("Zabbix version %s is less than to %s which %s, skipping test.", serverVersion, comparedVersion, msg)
	}
}

func TestBadCalls(t *testing.T) {
	api := testGetAPI(t)
	res, err := api.Call("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Error.Code != -32600 && res.Error.Code != -32602 {
		t.Errorf("Expected code -32600 or -32602 depending on Zabbix Server version, got %s", res.Error)
	}
}

func TestVersion(t *testing.T) {
	api := testGetAPI(t)
	v, err := api.Version()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Zabbix version %s", v)
	if !regexp.MustCompile(`^\d\.\d\.\d+$`).MatchString(v) {
		t.Errorf("Unexpected version: %s", v)
	}
}
