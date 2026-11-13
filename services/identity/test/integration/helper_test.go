package integration

import "isletalk/identity/domain/valueobject"

func newID(s string) (valueobject.UserId, error) { return valueobject.NewUserId(s) }
func newName(s string) (valueobject.PublicName, error) {
	return valueobject.NewPublicName(s)
}
func newPhone(s string) (valueobject.PhoneHash, error) { return valueobject.NewPhoneHash(s) }
