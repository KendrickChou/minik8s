package couchdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/klog/v2"
	"minik8s.com/minik8s/pkg/aqualake/apis/config"
	"minik8s.com/minik8s/pkg/aqualake/apis/constants"
)

func PutDatabase(ctx context.Context, id string) error {
	req, err := http.NewRequest("PUT", constants.CouchPutDBRequest(id), nil)

	if err != nil {
		klog.Errorf("Create Data Base Failed: %s", err.Error())
		return err
	}

	req.SetBasicAuth(config.CouchDBUser, config.CouchDBPasswd)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		klog.Errorf("Create Data Base Failed: %s", err.Error())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Errorf("Create Data Base Failed: %s", err.Error())
		return err
	}

	var res map[string]interface{}
	json.Unmarshal(body, &res)

	if _, ok := res["ok"]; !ok {
		errInfo := fmt.Sprintf("Create Data Base Failed: %s", string(body))
		klog.Error(errInfo)
		return errors.New(errInfo)
	}

	resp.Body.Close()

	return nil
}

func GetDoc(ctx context.Context, db, id string) ([]byte, error) {
	req, err := http.NewRequest("GET", constants.CouchGetDocRequest(db, id), nil)

	if err != nil {
		klog.Errorf("Get Doc Failed: %s", err.Error())
		return nil, err
	}

	req.SetBasicAuth(config.CouchDBUser, config.CouchDBPasswd)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		klog.Errorf("Get Doc Failed: %s", err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Errorf("Get Doc Failed: %s", err.Error())
		return nil, err
	}

	var res map[string]interface{}
	json.Unmarshal(body, &res)

	if _, ok := res["error"]; ok {
		errInfo := fmt.Sprintf("Get Doc Failed: %s", string(body))
		klog.Error(errInfo)
		return nil, errors.New(errInfo)
	}

	return body, nil
}

// return reversion, error
func PutDoc(ctx context.Context, db, id, doc string) (string, error) {
	// var reader io.Reader

	req, err := http.NewRequest("PUT", constants.CouchPutDocRequest(db, id), strings.NewReader(doc))

	if err != nil {
		klog.Errorf("Put Doc Failed: %s", err.Error())
		return "", err
	}

	req.SetBasicAuth(config.CouchDBUser, config.CouchDBPasswd)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		klog.Errorf("Put Doc Failed: %s", err.Error())
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Errorf("Put Doc Failed: %s", err.Error())
		return "", err
	}

	var res map[string]interface{}
	json.Unmarshal(body, &res)

	if _, ok := res["error"]; ok {
		errInfo := fmt.Sprintf("Put Doc Failed: %s", string(body))
		klog.Error(errInfo)
		return "", errors.New(errInfo)
	}

	return fmt.Sprintf("%v", res["rev"]), nil
}

func DelDoc(ctx context.Context, db, id, rev string) error {
	req, err := http.NewRequest("DELETE", constants.CouchDelDocRequest(db, id, rev), nil)

	if err != nil {
		klog.Errorf("Del Doc Failed: %s", err.Error())
		return err
	}

	req.SetBasicAuth(config.CouchDBUser, config.CouchDBPasswd)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		klog.Errorf("Del Doc Failed: %s", err.Error())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Errorf("Del Doc Failed: %s", err.Error())
		return err
	}

	var res map[string]interface{}
	json.Unmarshal(body, &res)

	if _, ok := res["error"]; ok {
		errInfo := fmt.Sprintf("Del Doc Failed: %s", string(body))
		klog.Error(errInfo)
		return errors.New(errInfo)
	}

	return nil
}

func PutFile(ctx context.Context, db, docId, fileId, rev, data string) error {
	req, err := http.NewRequest("PUT", constants.CouchPutFileRequest(db, docId, fileId, rev), strings.NewReader(data))

	if err != nil {
		klog.Errorf("Put File Failed: %s", err.Error())
		return err
	}

	req.SetBasicAuth(config.CouchDBUser, config.CouchDBPasswd)
	req.Header.Set("ContentType", "text/plain")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		klog.Errorf("Put File Failed: %s", err.Error())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Errorf("Put File Failed: %s", err.Error())
		return err
	}

	var res map[string]interface{}
	json.Unmarshal(body, &res)

	if _, ok := res["error"]; ok {
		errInfo := fmt.Sprintf("Put File Failed: %s", string(body))
		klog.Error(errInfo)
		return errors.New(errInfo)
	}

	return nil
}

func GetFile(ctx context.Context, db, docId, fileId string) (string, error) {
	req, err := http.NewRequest("GET", constants.CouchGetFileRequest(db, docId, fileId), nil)

	if err != nil {
		klog.Errorf("Get File Failed: %s", err.Error())
		return "", err
	}

	req.SetBasicAuth(config.CouchDBUser, config.CouchDBPasswd)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		klog.Errorf("Get File Failed: %s", err.Error())
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Errorf("Get File Failed: %s", err.Error())
		return "", err
	}

	return string(body), nil
}
