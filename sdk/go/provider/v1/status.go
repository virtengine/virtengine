package v1

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

type Storage map[string]*resource.Quantity
