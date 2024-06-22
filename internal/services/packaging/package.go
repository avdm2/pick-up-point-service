package packaging

import (
	"errors"
	"homework-1/internal/models"
)

var (
	errWeightExceededErr = errors.New("weight exceeded")
	invalidPackageErr    = errors.New("invalid package")
)

type Package interface {
	ValidateWeight(weight models.Kilo) error
	GetCost() models.Rub
}

func ParsePackage(p models.PackageType) (Package, error) {
	switch p {
	case "bag":
		return Bag{}, nil
	case "box":
		return Box{}, nil
	case "wrap":
		return Wrap{}, nil
	}
	return nil, invalidPackageErr
}
