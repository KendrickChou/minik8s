package couchobject

import (
	"minik8s.com/minik8s/pkg/aqualake/apis/actionchain"
	"minik8s.com/minik8s/pkg/aqualake/apis/couchmeta"
)

type Function struct {
	couchmeta.CouchMeta
}

type ActionChain struct {
	couchmeta.CouchMeta

	actionchain.ActionChain
}