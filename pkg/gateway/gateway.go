package gateway

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/vishvananda/netlink"
	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/gateway/apis/config"
	"minik8s.com/minik8s/pkg/gateway/apis/constants"
	"minik8s.com/minik8s/pkg/gateway/apis/httpresponse"
)

func NewGateway() {
	nodeChangeRaw := make(chan []byte)
	errChan := make(chan string)

	handler, err := netlink.NewHandle()

	if err != nil {
		klog.Fatalf("Init Netlink Failed: %s", err.Error())
		os.Exit(0)
	}

	go watchingNodes(context.TODO(), handler, nodeChangeRaw, errChan)

	for {
		select {
		case e := <-errChan:
			klog.Errorf("Gateway listen failed: %s", e)
			return
		case rawBytes := <-nodeChangeRaw:
			req := &httpresponse.WatchNodeResponse{}
			err := json.Unmarshal(rawBytes, req)

			if err != nil {
				klog.Error("Unmarshal APIServer Data Failed: %s", err.Error())
			} else {
				err = handleNodeChangeRequest(handler, req)
				if err != nil {
					klog.Errorln(err)
				}
			}
		}
	}

}

func watchingNodes(ctx context.Context, handler *netlink.Handle, nodeChange chan []byte, errChan chan string) {
	resp, err := http.Get(config.ApiServerAddress + constants.WatchNodeRequest)

	if err != nil {
		klog.Errorf("Watching Nodes Failed: %s", err.Error())
		errChan <- err.Error()
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			reader := bufio.NewReader(resp.Body)
			buf, err := reader.ReadBytes(byte(constants.EOF))

			if err != nil {
				klog.Errorf("Watch Nodes Error: %s", err)
				errChan <- err.Error()
				return
			}

			nodeChange <- buf
		}
	}
}

func handleNodeChangeRequest(handler *netlink.Handle, req *httpresponse.WatchNodeResponse) error {
	var err error

	switch req.Type {
	case "PUT":
		err = addRoute(handler, &req.Node)
	case "DELETE":
		err = deleteRoute(handler, &req.Node)
	default:
		errInfo := "Unknown Pod Change Request Type: " + req.Type
		klog.Errorln(errInfo)
		err = errors.New(errInfo)
	}

	return err
}

func addRoute(handler *netlink.Handle, node *v1.Node) error {
	if node.Spec.CIDR == "" || node.Spec.IP == "" {
		errInfo := fmt.Sprintf("Add route failed: node %s, CIDR %s, IP %s", node.UID, node.Spec.CIDR, node.Spec.IP)
		return errors.New(errInfo)
	}

	_, dst, err := net.ParseCIDR(node.Spec.CIDR)

	if err != nil {
		return err
	}

	err = netlink.RouteAdd(&netlink.Route{
		Dst: dst,
		Gw:  net.ParseIP(node.Spec.IP),
	})

	return err
}

func deleteRoute(handler *netlink.Handle, node *v1.Node) error {
	if node.Spec.CIDR == "" || node.Spec.IP == "" {
		errInfo := fmt.Sprintf("Add route failed: node %s, CIDR %s, IP %s", node.UID, node.Spec.CIDR, node.Spec.IP)
		return errors.New(errInfo)
	}

	_, dst, err := net.ParseCIDR(node.Spec.CIDR)

	if err != nil {
		return err
	}

	err = netlink.RouteDel(&netlink.Route{
		Dst: dst,
		Gw:  net.ParseIP(node.Spec.IP),
	})

	return err
}
