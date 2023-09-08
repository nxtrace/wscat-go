package pow

import (
	"net/url"

	"github.com/tsosunchia/powclient"
)

const (
	protocol = "https"
	hostSNI  = "api.leo.moe"
	baseURL  = "/v3/challenge"
)

func GetToken() (string, error) {
	p := powclient.NewGetTokenParams()

	u := url.URL{
		Scheme: protocol,
		Host:   hostSNI,
		Path:   baseURL,
	}

	p.BaseUrl = u.String()
	p.SNI = hostSNI
	p.Host = hostSNI

	for i := 0; i < 3; i++ {
		token, err := powclient.RetToken(p)
		if err != nil {
			continue
		}
		return token, err
	}

	return "", nil
}
