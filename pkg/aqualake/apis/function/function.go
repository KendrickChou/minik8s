package function

import (
	"minik8s.com/minik8s/pkg/aqualake/apis/couchmeta"
)

type Function struct {
	couchmeta.CouchMeta

	ReturnType string
}
