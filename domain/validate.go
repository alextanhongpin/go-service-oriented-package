package domain

type validatable interface {
	Validate() error
}

func Validate(vs ...validatable) error {
	for _, v := range vs {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}
