package authbe

import (
	ed "crypto/ed25519"
	pb "golang.conradwood.net/apis/auth"
	//	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/common"
	"golang.conradwood.net/go-easyops/utils"
	"testing"
	"time"
)

const (
	writefile = false
)

var (
	// Variable "authpubkey": byte contents of file '/tmp/auth-public.key' converted on 2020-04-16 13:03:48
	authpubkey = []byte{
		0xC7, 0x99, 0x37, 0x2D, 0x37, 0x54, 0x9F, 0x69, 0x0F, 0x13, 0x3A,
		0xCE, 0x35, 0xBF, 0x9C, 0x54, 0x92, 0x8F, 0xE3, 0x06, 0xB5, 0x9B,
		0xF0, 0x55, 0xD5, 0x34, 0xE4, 0x68, 0x21, 0x44, 0x65, 0x18,
	}
)

type fn func(*pb.User) string

// it's important to test if the public key changes because if it does it will break EEEEVERYTHING. Whilst unlikely, the impact is high.
func TestKey(t *testing.T) {
	*private_seed = "mae2iu3kaiFee0ahwoo5aim4AN8ach7F" // made up, random, insecure seed
	s := signPrivateKey()
	expect := 32
	if len(s) != expect {
		t.Logf("private key len does not match expected key length (%d != %d)", expect, len(s))
	}
	b := signPublicKey()
	if writefile {
		utils.WriteFile("/tmp/auth-public.key", b)
	}
	if !equalSlice(authpubkey, b) {
		t.Errorf("Invalid key")
	}
}
func equalSlice(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range a {
		if b[i] != x {
			return false
		}
	}
	return true
}

func TestSign(t *testing.T) {
	myUser := &pb.User{
		ID: "11",
	}
	sign(myUser)
	sig := myUser.SignatureID
	time.Sleep(2 * time.Second)
	sign(myUser)
	if equalSlice(myUser.SignatureID, sig) {
		t.Errorf("same signatures for same user at different times\n")
	}
}
func TestSigLen(t *testing.T) {
	myUser := &pb.User{ID: "15"}
	sign(myUser)
	if len(myUser.SignatureID) != 64 {
		t.Errorf("Signature not 64 Bytes long (len==%d)", len(myUser.SignatureID))
	}
}

func detect(t *testing.T, foo fn) {
	myUser := &pb.User{ID: "12311"}
	sign(myUser)
	v := ed.Verify(signPublicKey(), common.SignAllbytes(myUser), myUser.SignatureFull)
	if !v {
		t.Errorf("verify() failed\n")
	}
	ch := foo(myUser)
	v = ed.Verify(signPublicKey(), common.SignAllbytes(myUser), myUser.SignatureFull)
	if v {
		t.Errorf("signature verification did not detect change of \"%s\"", ch)
	}

}
func TestDetectFirstName(t *testing.T) {
	detect(t, func(in *pb.User) string {
		in.FirstName = "foo"
		return "FirstName"
	})
}
func TestDetectLastName(t *testing.T) {
	detect(t, func(in *pb.User) string {
		in.LastName = "foo"
		return "LastName"
	})
}
func TestDetectAbbrev(t *testing.T) {
	detect(t, func(in *pb.User) string {
		in.Abbrev = "foo"
		return "Abbrev"
	})
}
func TestDetectServiceAccount(t *testing.T) {
	detect(t, func(in *pb.User) string {
		in.ServiceAccount = !in.ServiceAccount
		return "ServiceAccount"
	})
}
func TestDetectActive(t *testing.T) {
	detect(t, func(in *pb.User) string {
		in.Active = !in.Active
		return "Active"
	})
}
func TestDetectEmailVerified(t *testing.T) {
	detect(t, func(in *pb.User) string {
		in.EmailVerified = !in.EmailVerified
		return "EmailVerified"
	})
}
func TestDetectEmail(t *testing.T) {
	detect(t, func(in *pb.User) string {
		in.Email = "foo"
		return "Email"
	})
}

func TestVerify(t *testing.T) {
	myUser := &pb.User{
		ID: "11",
	}
	sign(myUser)
	v := ed.Verify(signPublicKey(), signbytes(myUser), myUser.SignatureID)
	if !v {
		t.Errorf("verify() failed\n")
	}

	u2 := &pb.User{}
	*u2 = *myUser
	u2.ID = u2.ID + "foo"
	v = ed.Verify(signPublicKey(), signbytes(u2), u2.SignatureID)
	if v {
		t.Errorf("verify() did not detect id change\n")
	}

	*u2 = *myUser
	u2.ID = "12"
	v = ed.Verify(signPublicKey(), signbytes(u2), u2.SignatureID)
	if v {
		t.Errorf("verify() did not detect id increment\n")
	}

	*u2 = *myUser
	u2.SignedAt = 0
	v = ed.Verify(signPublicKey(), signbytes(u2), u2.SignatureID)
	if v {
		t.Errorf("verify() did not detect timestamp change\n")
	}
}
func TestFullVerify(t *testing.T) {
	myUser := &pb.User{
		ID: "11",
	}
	sign(myUser)
	v := ed.Verify(signPublicKey(), common.SignAllbytes(myUser), myUser.SignatureFull)
	if !v {
		t.Errorf("verify() failed\n")
	}

	u2 := &pb.User{}
	*u2 = *myUser
	u2.ID = u2.ID + "foo"
	v = ed.Verify(signPublicKey(), common.SignAllbytes(u2), u2.SignatureFull)
	if v {
		t.Errorf("verify() did not detect id change\n")
	}

	*u2 = *myUser
	u2.ID = "12"
	v = ed.Verify(signPublicKey(), common.SignAllbytes(u2), u2.SignatureFull)
	if v {
		t.Errorf("verify() did not detect id increment\n")
	}

	*u2 = *myUser
	u2.SignedAt = 0
	v = ed.Verify(signPublicKey(), common.SignAllbytes(u2), u2.SignatureFull)
	if v {
		t.Errorf("verify() did not detect timestamp change\n")
	}

	*u2 = *myUser
	u2.Email = "bademail"
	v = ed.Verify(signPublicKey(), common.SignAllbytes(u2), u2.SignatureFull)
	if v {
		t.Errorf("verify() did not detect email change\n")
	}
}

func TestVerifyMethod(t *testing.T) {
	myUser := &pb.User{
		ID: "11",
	}
	sign(myUser)
	sat := myUser.SignedAt
	v := VerifySignature(myUser)
	if !v {
		t.Errorf("verify()-method failed\n")
	}

	u2 := &pb.User{}
	*u2 = *myUser
	u2.ID = u2.ID + "foo"
	v = VerifySignature(u2)
	if v {
		t.Errorf("verify()-method did not detect id change\n")
	}

	*u2 = *myUser
	u2.ID = "12"
	v = VerifySignature(u2)
	if v {
		t.Errorf("verify()-method did not detect id increment\n")
	}

	*u2 = *myUser
	u2.SignedAt = 0
	v = VerifySignature(u2)
	if v {
		t.Errorf("verify()-method did not detect timestamp change (signed at before: %d, after: %d)\n", sat, u2.SignedAt)
	}

	*u2 = *myUser
	u2.Email = "bademail"
	v = VerifySignature(u2)
	if v {
		t.Errorf("verify()-method did not detect email change\n")
	}
}
