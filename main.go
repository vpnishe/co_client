package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"github.com/denisbrodbeck/machineid"
	"github.com/vpnishe/anyvalue"
	"github.com/vpnishe/elog"
	core "github.com/vpnishe/co_core"
)

var plog *elog.EasyLogger
var Config *anyvalue.AnyValue
var configPath string

const (
	CLIENT_STOPPED  = 0
	CLIENT_STARTING = 1
	CLIENT_STARTED  = 2
)


func GetDeviceId() string {
	id, err := machineid.ID()
	if err != nil {
		return "11111111111111111111111111111111"
	}

	h := md5.New()
	h.Write([]byte(id))
	result := hex.EncodeToString(h.Sum(nil))

	return result
}


func init() {
	flag.StringVar(&configPath, "config", "./config.json", "config file path")
	plog = elog.GetLogger()

}

func signalHandler(pc *core.PoleVpnClient) {

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				plog.Info("receive exit signal,exit")
				if pc != nil {
					pc.Stop()
				}
				plog.Flush()
				os.Exit(0)
			default:
			}
		}
	}()
}

var networkmgr core.NetworkManager
var device *core.TunDevice

func eventHandler(event int, client *core.PoleVpnClient, av *anyvalue.AnyValue) {

	switch event {
	case core.CLIENT_EVENT_ADDRESS_ALLOCED:
		{
			var err error
			var routes []string
			if Config.Get("use_remote_route").AsBool() {
				routes = append(routes, av.Get("route").AsStrArr()...)
			}

			if len(Config.Get("route_networks").AsStrArr()) > 0 {
				routes = append(routes, Config.Get("route_networks").AsStrArr()...)
			}

			if len(Config.Get("proxy_domains").AsStrArr()) > 0 {
				ips := GetRouteIpsFromDomain(Config.Get("proxy_domains").AsStrArr())
				routes = append(routes, ips...)
			}

			elog.Info("route=", routes, ",allocated ip=", av.Get("ip").AsStr(), ",dns=", av.Get("dns").AsStr())

			if runtime.GOOS == "windows" {
				err = device.GetInterface().SetTunNetwork(av.Get("ip").AsStr() + "/30")
				if err != nil {
					plog.Error("set tun network fail,", err)
					client.Stop()
				}
			}

			if !Config.Get("use_remote_dns").AsBool() {
				av.Set("dns", "")
			}

			err = networkmgr.SetNetwork(device.GetInterface().Name(), av.Get("ip").AsStr(), client.GetRemoteIP(), av.Get("dns").AsStr(), routes)
			if err != nil {
				plog.Error("set network fail,", err)
				client.Stop()
			}
		}
	case core.CLIENT_EVENT_STOPPED:
		{
			plog.Info("client stoped")
			networkmgr.RestoreNetwork()
		}
	case core.CLIENT_EVENT_RECONNECTED:
		plog.Info("client reconnected")
	case core.CLIENT_EVENT_RECONNECTING:
		err := networkmgr.RefreshDefaultGateway()
		if err != nil {
			plog.Error("refresh default gateway fail,", err)
		}
		plog.Info("client reconnecting")
	case core.CLIENT_EVENT_STARTED:
		plog.Info("client started")
	case core.CLIENT_EVENT_ERROR:
		plog.Info("client error ", av.Get("error").AsStr())
	default:
		plog.Error("invalid event=", event)
	}

}

func main() {

	flag.Parse()
	defer plog.Flush()
	var err error

	Config, err = GetConfig(configPath)

	if err != nil {
		elog.Fatal("load config fail", err)
	}

	device, err = core.NewTunDevice()
	if err != nil {
		elog.Fatal("create device fail,", err)
		return
	}

	if runtime.GOOS == "darwin" {
		networkmgr = core.NewDarwinNetworkManager()
	} else if runtime.GOOS == "linux" {
		networkmgr = core.NewLinuxNetworkManager()
	} else if runtime.GOOS == "windows" {
		networkmgr = core.NewWindowsNetworkManager()
	} else {
		plog.Fatal("os platform not support")
	}

	client, err := core.NewPoleVpnClient()

	if err != nil {
		plog.Fatal("new polevpn client fail,", err)
	}


	client.SetEventHandler(eventHandler)
	
	client.AttachTunDevice(device)

	deviceType := "Unknown"
	if runtime.GOOS == "darwin" {
		networkmgr = core.NewDarwinNetworkManager()
	} else if runtime.GOOS == "linux" {
		networkmgr = core.NewLinuxNetworkManager()
	} else if runtime.GOOS == "windows" {
		networkmgr = core.NewWindowsNetworkManager()
	} else {
		plog.Fatal("os platform not support")
	}

	deviceId := GetDeviceId()
	err = client.Start(Config.Get("endpoint").AsStr(), Config.Get("user").AsStr(), Config.Get("password").AsStr(), Config.Get("sni").AsStr(), Config.Get("skipVerifySSL").AsBool(), deviceType, deviceId)
	if err != nil {
		plog.Fatal("start polevpn client fail,", err)
	}

	signalHandler(client)

	client.WaitStop()
}
