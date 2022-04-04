package vov2

// UserRole 用户身份枚举.
type UserRole int32

const (
	USER_ROLE_MASTER_TEACHER UserRole = 3
	USER_ROLE_STUDENT        UserRole = 4

	// Undefined ...
	Undefined UserRole = -1
	// USER_ROLE_GUEST 游客模式.
	USER_ROLE_GUEST UserRole = 0
	// USER_ROLE_ORDINARY 普通用户.
	USER_ROLE_ORDINARY UserRole = 1
	// USER_ROLE_VIP VIP.
	USER_ROLE_VIP UserRole = 2
)

var (
	UserRoleMap = map[UserRole]string{
		0: "USER_ROLE_GUEST",
		1: "USER_ROLE_ORDINARY",
		2: "USER_ROLE_VIP",
	}

	UserRoleMapToInt = map[string]UserRole{
		"USER_ROLE_GUEST":    0,
		"USER_ROLE_ORDINARY": 1,
		"USER_ROLE_VIP":      2,
	}
)

//UserRoleToInt32 ...
func UserRoleToInt32(input string) int32 {
	if UserRoleToInt, ok := UserRoleMapToInt[input]; ok {
		return int32(UserRoleToInt)
	}
	return 0
}

// FromString ...
func FromString(input int32) string {
	if courseType, ok := UserRoleMap[UserRole(input)]; ok {
		return courseType
	}
	return ""
}
