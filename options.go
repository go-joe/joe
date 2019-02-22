package joe

type Option func(*Bot) error

func WithBrain(brain Brain) Option {
	return func(b *Bot) error {
		b.Brain = brain
		return nil
	}
}
