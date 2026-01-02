package inbound

type MePermissionsResponse struct {
	Permissions map[string][]string `json:"permissions"`
}
