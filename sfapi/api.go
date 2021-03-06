package sfapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/alecthomas/units"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	SVIP              string
	Endpoint          string
	DefaultAPIPort    int
	DefaultVolSize    int64 //bytes
	DefaultAccountID  int64
	DefaultTenantName string
	VolumeTypes       *[]VolType
	Config            *Config
}

type Config struct {
	TenantName     string
	EndPoint       string
	DefaultVolSz   int64 //Default volume size in GiB
	MountPoint     string
	SVIP           string
	InitiatorIFace string //iface to use of iSCSI initiator
	Types          *[]VolType
}

type VolType struct {
	Type string
	QOS  QoS
}

var (
	endpoint          string
	svip              string
	configFile        string
	defaultTenantName string
	defaultSizeGiB    int64
	cfg               Config
)

func init() {
	if os.Getenv("SF_CONFIG_FILE") != "" {
		conf, _ := ProcessConfig(configFile)
		cfg = conf
		endpoint = conf.EndPoint
		svip = conf.SVIP
		configFile = os.Getenv("SF_CONFIG_FILE")
		defaultSizeGiB = conf.DefaultVolSz
		defaultTenantName = conf.TenantName
	} else {
		endpoint = os.Getenv("SF_ENDPOINT")
		svip = os.Getenv("SF_SVIP")
		configFile = os.Getenv("SF_CONFIG_FILE")
		defaultSizeGiB, _ = strconv.ParseInt(os.Getenv("SF_DEFAULT_VSIZE"), 10, 64)
		defaultTenantName = os.Getenv("SF_DEFAULT_TENANT_NAME")
	}
}

func ProcessConfig(fname string) (Config, error) {
	content, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal("Error processing config file: ", err)
	}
	var conf Config
	err = json.Unmarshal(content, &conf)
	if err != nil {
		log.Fatal("Error parsing config file: ", err)
	}
	return conf, nil
}

func NewFromConfig(configFile string) (c *Client, err error) {
	conf, err := ProcessConfig(configFile)
	if err != nil {
		log.Fatal("Error initializing client from Config file: ", configFile, "(", err, ")")
	}
	cfg = conf
	endpoint = conf.EndPoint
	svip = conf.SVIP
	configFile = os.Getenv("SF_CONFIG_FILE")
	defaultSizeGiB = conf.DefaultVolSz
	defaultTenantName = conf.TenantName
	return New()
}

func New() (c *Client, err error) {
	rand.Seed(time.Now().UTC().UnixNano())
	defSize := defaultSizeGiB * int64(units.GiB)
	SFClient := &Client{
		Endpoint:          endpoint,
		DefaultVolSize:    defSize,
		SVIP:              svip,
		Config:            &cfg,
		DefaultAPIPort:    443,
		VolumeTypes:       cfg.Types,
		DefaultTenantName: defaultTenantName,
	}
	return SFClient, nil
}

func (c *Client) Request(method string, params interface{}, id int) (response []byte, err error) {
	log.Debug("Issue request to SolidFire Endpoint...")
	if c.Endpoint == "" {
		log.Error("Endpoint is not set, unable to issue requests")
		err = errors.New("Unable to issue json-rpc requests without specifying Endpoint")
		return nil, err
	}
	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"id":     id,
		"params": params,
	})

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	log.Debugf("POST request to: %+v", c.Endpoint)
	Http := &http.Client{Transport: tr}
	resp, err := Http.Post(c.Endpoint,
		"json-rpc",
		strings.NewReader(string(data)))
	if err != nil {
		log.Errorf("Error encountered posting request: %v", err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}

	var prettyJson bytes.Buffer
	_ = json.Indent(&prettyJson, body, "", "  ")
	log.WithField("", prettyJson.String()).Debug("request:", id, " method:", method, " params:", params)

	errresp := APIError{}
	json.Unmarshal([]byte(body), &errresp)
	if errresp.Error.Code != 0 {
		err = errors.New("Received error response from API request")
		return body, err
	}
	return body, nil
}

func newReqID() int {
	return rand.Intn(1000-1) + 1
}

func NewReqID() int {
	return rand.Intn(1000-1) + 1
}
