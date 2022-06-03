package couchobject

import (
	"minik8s.com/minik8s/pkg/aqualake/apis/actionchain"
	"minik8s.com/minik8s/pkg/aqualake/apis/couchmeta"
	"minik8s.com/minik8s/pkg/aqualake/apis/function"
)

type Function struct {
	couchmeta.CouchMeta

	function.Function
}

type ActionChain struct {
	couchmeta.CouchMeta

	actionchain.ActionChain
}
