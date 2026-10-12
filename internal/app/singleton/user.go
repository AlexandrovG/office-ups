package singleton

const CurrentUserID uint = 1

func CurrentUser() uint {
	return CurrentUserID
}