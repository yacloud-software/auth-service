package authbe

/*

This signs a userobjects' fields.
Note that the "public" part is in the common go-easyops "auth" package.
Amongst other things it contains the conversion from proto to bytes

*/
import (
	//	"crypto"
	ed "crypto/ed25519"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	pb "golang.conradwood.net/apis/auth"
	//	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/cmdline"
	"golang.conradwood.net/go-easyops/common"
	"time"
)

var (
	private_seed = flag.String("private_key_seed", "", "private key seed for signing users. must match on all auth-servers")
)

func signPublicKey() ed.PublicKey {
	pk := signPrivateKey().Public().(ed.PublicKey)
	return pk
}
func signPrivateKey() ed.PrivateKey {
	if len(cmdline.OptEnvString(*private_seed, "AUTH_PRIVATE_KEY_SEED")) != 32 {
		panic(fmt.Sprintf("Private key seed has invalid length (%d != 32)", len(cmdline.OptEnvString(*private_seed, "AUTH_PRIVATE_KEY_SEED"))))
	}
	pk := ed.NewKeyFromSeed([]byte(cmdline.OptEnvString(*private_seed, "AUTH_PRIVATE_KEY_SEED")))
	return pk
}

// sign a user object
func sign(in *pb.User) {
	in.SignedAt = uint32(time.Now().Unix())
	b := signbytes(in)
	s := ed.Sign(signPrivateKey(), b)
	in.SignatureID = s

	b = common.SignAllbytes(in)
	s = ed.Sign(signPrivateKey(), b)
	in.SignatureFull = s
}
func userToSignedUser(user *pb.User) (*pb.SignedUser, error) {
	data, err := proto.Marshal(user)
	if err != nil {
		return nil, err
	}
	res := &pb.SignedUser{User: data}
	res.Signature = ed.Sign(signPrivateKey(), res.User)
	return res, nil
}
