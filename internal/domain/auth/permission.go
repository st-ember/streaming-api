package auth

const (
	PermissionVideoUpload  = "video:upload"
	PermissionVideoUpdate  = "video:update"
	PermissionVideoArchive = "video:archive"
)

// AllPermissions returns a slice containing all defined permission slugs.
func AllPermissions() []string {
	return []string{
		PermissionVideoUpload,
		PermissionVideoUpdate,
		PermissionVideoArchive,
	}
}
