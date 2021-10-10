package authbe

import (
	ed "crypto/ed25519"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/go-easyops/common"
)

// get the bytes from a proto that ought to be signed
func signbytes(in *pb.User) []byte {
	b := []byte(in.ID)
	ts := in.SignedAt
	x := ts
	for i := 0; i < 4; i++ {
		b = append(b, byte(x&0xFF))
		x = x << 8
	}
	return b
}

/*
user has 2 signatures, one for the ID only and one "full" over all fields.
this one verifies the "full" signature (e.g. true indicates that the all fields in the user object
have been created by a 'real' auth-service and can be trusted)
*/
func VerifySignature(u *pb.User) bool {
	v := ed.Verify(signPublicKey(), common.SignAllbytes(u), u.SignatureFull)
	return v
}
