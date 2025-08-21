package secret

type Generator interface {
	Generate(length int) (string, error)
	GenerateWithPrefix(prefix string, length int) (string, error)
}
