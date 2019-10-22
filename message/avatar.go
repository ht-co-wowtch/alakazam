package message

const (
	avatarFemale = "female"
	avatarMale   = "male"
	avatarOther  = "other"
	avatarRoot   = "root"
)

func toAvatarName(code int) string {
	switch code {
	case 0:
		return avatarFemale
	case 1:
		return avatarMale
	case 2:
		return avatarOther
	case 99:
		return avatarRoot
	}
	return avatarOther
}