package config

type SMS struct {
	Recipient string
}

func (s *SMS) Send() error {
	return nil
}
