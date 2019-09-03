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

func ToAvatarCode(name string) int {
	switch name {
	case avatarFemale:
		return 0
	case avatarMale:
		return 1
	case avatarOther:
		return 2
	case avatarRoot:
		return 99
	}
	return 2
}
