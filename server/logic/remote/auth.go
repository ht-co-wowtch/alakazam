package remote

import "gitlab.com/jetfueltw/cpw/alakazam/server/business"

func Renew(token string) (string, string, int) {
	if token == "0" {
		return "", "", business.Blockade
	}
	return "82ea16cd2d6a49d887440066ef739669", "test", business.PlayDefaultPermission
}
