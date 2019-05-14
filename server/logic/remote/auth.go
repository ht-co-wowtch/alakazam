package remote

import "gitlab.com/jetfueltw/cpw/alakazam/server/business"

func Renew(token string) (string, string, int) {
	if token == "0" {
		return "", "", business.Blockade
	}
	if token == "1" {
		return "82ea16cd2d6a49d887440066ef739669", "test", business.PlayDefaultPermission - 2
	}
	return "82ea16cd2d6a49d887440066ef739669", "test", business.PlayDefaultPermission
}
