package fbm_strongo_bot

type FbmPlatform struct {
}

func (p FbmPlatform) Id() string {
	return "fbm"
}

func (p FbmPlatform) Version() string {
	return "1"
}