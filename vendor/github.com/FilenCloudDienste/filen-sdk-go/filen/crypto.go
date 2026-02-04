package filen

import (
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"strconv"
	"strings"
)

// HashFileName hashes a file name, this is used for file and directory names
// and is dependent on the auth version, version 1 and 2 use the crypto.V2Hash
// function, version 3 uses the HMACKey
func (api *Filen) HashFileName(name string) string {
	name = strings.ToLower(name)
	switch api.AuthVersion {
	case 1, 2:
		return crypto.V2Hash([]byte(name))
	default:
		return api.HMACKey.Hash([]byte(name))
	}
}

// EncryptMeta encrypts metadata, this is dependent on the auth version
// version 1 is unimplemented, 2 uses the MasterKeys, and 3 uses the DEK
func (api *Filen) EncryptMeta(metadata string) crypto.EncryptedString {
	switch api.MetadataEncryptionVersion {
	case 1:
		panic("unsupported version")
	case 2:
		return api.MasterKeys.EncryptMeta(metadata)
	case 3:
		return api.DEK.EncryptMeta(metadata)
	default:
		panic("unsupported version")
	}
}

// DecryptMeta decrypts metadata, this reads the encrypted string to determine
// whether to use the MasterKeys or DEK
func (api *Filen) DecryptMeta(encrypted crypto.EncryptedString) (string, error) {
	if encrypted[0:8] == "U2FsdGVk" {
		decrypted, err := api.MasterKeys.DecryptMetaV1(encrypted)
		return decrypted, err
	}
	switch encrypted[0:3] {
	case "002":
		decrypted, err := api.MasterKeys.DecryptMetaV2(encrypted)
		return decrypted, err
	case "003":
		decrypted, err := api.DEK.DecryptMeta(encrypted)
		return decrypted, err
	default:
		panic("unsupported version")
	}
}

func (api *Filen) GetMetaCrypterFromKeyString(keyStr string, v crypto.MetadataEncryptionVersion) (crypto.MetaCrypter, error) {
	if v == -1 {
		v = api.MetadataEncryptionVersion
	}

	if v == 3 {
		_, err := strconv.ParseUint(keyStr, 16, 64)
		if err != nil || len(keyStr) != 64 {
			v = 2
		}
	}
	switch v {
	case 1:
		panic("unsupported version")
	case 2:
		return crypto.NewMasterKey([]byte(keyStr))
	case 3:
		return crypto.MakeEncryptionKeyFromStr(keyStr)
	default:
		panic("unsupported version")
	}
}
