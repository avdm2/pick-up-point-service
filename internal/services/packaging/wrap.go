package packaging

import "homework-1/internal/models"

type Wrap struct{}

func (w Wrap) ValidateWeight(weight models.Kilo) error {
	return nil
}

func (w Wrap) GetCost() models.Rub {
	return 1
}
