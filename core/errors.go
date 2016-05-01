package bots

type AuthFailedError string

func (e AuthFailedError) Error() string {
	return string(e)
}
