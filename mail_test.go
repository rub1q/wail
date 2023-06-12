package wail

import "testing"

var m = NewMail(nil)

var invalidEmails = [...]string{
	"////", "*(())*", "1234",
	"<>", "^@", "####", "-_-",
	"++=1", "$%_", ";", "'/asd",
	"i am hero",
}

const veryLongEmail = `owBXheRtZT3c37SCAKT8BVcx6guSJRy
guptnxkKxE6jWahc9LmcOJ1jisAeOD6kZUundp
lK9U8dKJ7ymKgjSu1FvZQ9F9FOMgXD9XXWSQK
TwlNzSg5EcvgvWkVZEGBn1S4gCjSOQ5Ex2fDV
LUoXNYpTcFCb1AI9sxMmegKVxyOIx2ViWTMzS
yUi3oVMV3chUqd4Pa0NRIw4VzQIdJJSR4PJWc
EYzB6tzlflv37AhEDAeJ7jxCppcMFwaVV@example.com`

func univEmailAddressesTest(f func(emails ...string) error, t *testing.T) {
	if err := f(""); err == nil {
		t.Error("To adresses should not be empty")
	}

	for _, v := range invalidEmails {
		if err := f(v); err == nil {
			t.Error("Provided email address should be invalid")
		}
	}

	if err := f(veryLongEmail); err == nil {
		t.Error("Email address is too long and should not go further")
	}
}

func TestTo(t *testing.T) {
	univEmailAddressesTest(m.To, t)
}

func TestCopyTo(t *testing.T) {
	univEmailAddressesTest(m.CopyTo, t)
}

func TestBlindCopyTo(t *testing.T) {
	univEmailAddressesTest(m.BlindCopyTo, t)
}
