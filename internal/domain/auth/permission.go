package auth

const (
	PermissionVideoUpload  = "video:upload"
	PermissionVideoUpdate  = "video:update"
	PermissionVideoArchive = "video:archive"
)

// AllPermissions returns a slice containing all defined permissions.
func AllPermissions() []string {
	return []string{
		PermissionVideoUpload,
		PermissionVideoUpdate,
		PermissionVideoArchive,
	}
}
