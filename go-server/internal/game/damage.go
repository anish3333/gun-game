package game

func DamageForWeapon(w string) int {
	switch w {
	case "pistol":
		return 12
	case "shotgun":
		return 18
	case "sniper":
		return 40
	case "smg":
		return 7
	default:
		return 10
	}
}