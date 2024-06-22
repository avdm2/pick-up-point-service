package packaging

import "homework-1/internal/models"

type Box struct{}

func (b Box) ValidateWeight(weight models.Kilo) error {
	if weight >= 30 {
		return errWeightExceededErr
	}
	return nil
}

func (b Box) GetCost() models.Rub {
	return 20
}
