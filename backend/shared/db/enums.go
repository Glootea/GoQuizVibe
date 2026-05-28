package db

type GroupRole string

const (
	GroupRoleAdmin  GroupRole = "admin"
	GroupRoleMember GroupRole = "member"
)

type PermissionType string

const (
	PermissionTypeRead  PermissionType = "read"
	PermissionTypeWrite PermissionType = "write"
	PermissionTypeOwner PermissionType = "owner"
)

type RecipientType string

const (
	RecipientTypeUser  RecipientType = "user"
	RecipientTypeGroup RecipientType = "group"
)

type AssetType string

const (
	AssetTypeQuiz             AssetType = "quiz"
	AssetTypeLearningMaterial AssetType = "learning_material"
)

type StudentPermission string

const (
	StudentPermissionOpenToAll StudentPermission = "open_to_all"
	StudentPermissionAssigned  StudentPermission = "assigned"
	StudentPermissionPrivate   StudentPermission = "private"
)
