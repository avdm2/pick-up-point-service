package packaging

import "homework-1/internal/models"

type Bag struct{}

func (b Bag) ValidateWeight(weight models.Kilo) error {
	if weight >= 10 {
		return errWeightExceededErr
	}
	return nil
}

func (b Bag) GetCost() models.Rub {
	return 5
}
