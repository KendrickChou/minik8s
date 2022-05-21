package main

import (
	"context"
	"encoding/json"
	"fmt"

	"minik8s.com/minik8s/pkg/aqualake/apis/couchmeta"
	"minik8s.com/minik8s/pkg/aqualake/couchdb"
)

func main() {
	dbName := "test-db"
	docName := "test-doc"
	fileName := "demo.txt"

	err := couchdb.CreateDatabase(context.TODO(), dbName)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	_, err = couchdb.PutDoc(context.TODO(), dbName, docName, "{\"name\": \"kendrick\"}")

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	doc, err := couchdb.GetDoc(context.TODO(), dbName, docName)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Print(string(doc))

	var meta couchmeta.CouchMeta
	json.Unmarshal(doc, &meta)

	err = couchdb.PutFile(context.TODO(), dbName, docName, fileName, meta.Reversion, "Hello CouchDB !")

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	file, err := couchdb.GetFile(context.TODO(), dbName, docName, fileName)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Print(file)

	doc, err = couchdb.GetDoc(context.TODO(), dbName, docName)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Print(string(doc))

	json.Unmarshal(doc, &meta)

	err = couchdb.DelDoc(context.TODO(), dbName, docName, meta.Reversion)

	if err != nil {
		fmt.Print(err.Error())
		return
	}
}
