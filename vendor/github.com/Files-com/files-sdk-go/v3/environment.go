package files_sdk

type Environment int64

const (
	Production Environment = iota
	Staging
	Development
)

func NewEnvironment(env string) Environment {
	switch env {
	case "staging":
		return Staging
	case "development":
		return Development
	default:
		return Production
	}
}

func (e Environment) String() string {
	switch e {
	case Staging:
		return "staging"
	case Development:
		return "development"
	default:
		return "production"
	}
}

const (
	ProductionEndpoint  = "https://{{SUBDOMAIN}}.files.com"
	developmentEndpoint = "https://{{SUBDOMAIN}}.filesrails.test"
	stagingEndpoint     = "https://{{SUBDOMAIN}}.filesstaging.av"
)

func (e Environment) Endpoint() string {
	switch e {
	case Staging:
		return stagingEndpoint
	case Development:
		return developmentEndpoint
	default:
		return ProductionEndpoint
	}
}
