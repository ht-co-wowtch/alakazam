package remote

import (
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/business"
)

func Renew(token string) (string, string, int) {
	return "82ea16cd2d6a49d887440066ef739669", "test", business.PlayDefaultPermission
}
