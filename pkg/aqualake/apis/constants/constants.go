package constants

import "minik8s.com/minik8s/pkg/aqualake/apis/config"

func CouchPutDBRequest(name string) string {
	return config.CouchDBAddr + "/" + name
}

func CouchGetDocRequest(db, id string) string {
	return config.CouchDBAddr + "/" + db + "/" + id
}

func CouchPutDocRequest(db, id string) string {
	return config.CouchDBAddr + "/" + db + "/" + id
}

func CouchDelDocRequest(db, id, rev string) string {
	return config.CouchDBAddr + "/" + db + "/" + id + "?rev=" + rev
}

func CouchPutFileRequest(db, docId, fileId, rev string) string {
	return config.CouchDBAddr + "/" + db + "/" + docId + "/" + fileId + "?rev=" + rev
}

func CouchGetFileRequest(db, docId, fileId string) string {
	return config.CouchDBAddr + "/" + db + "/" + docId + "/" + fileId
}

const (
	FunctionDBId string = "aqualake-function"
	ActionDBId   string = "aqualake-actionchain"
)
