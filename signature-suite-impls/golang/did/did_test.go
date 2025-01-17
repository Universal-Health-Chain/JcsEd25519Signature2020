package did

import (
	"testing"

	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"

	"github.com/Universal-Health-Chain/JcsEd25519Signature2020/signature-suite-impls/golang/proof"
)

var (
	keySeed       = []byte("12345678901234567890123456789012")
	issuerPrivKey = ed25519.NewKeyFromSeed(keySeed) // this matches the public key in the DID Doc
	issuerPubKey  = issuerPrivKey.Public().(ed25519.PublicKey)
)

func TestGenerateDIDDocForIssuerWithServices(t *testing.T) {
	did := GenerateDID(&issuerPubKey)
	keyRef := did + "#" + InitialKey
	publicKeys := make(map[string]*ed25519.PublicKey)
	publicKeys[InitialKey] = &issuerPubKey
	issuer := "fooIssuer"
	schemaID := "schemaID"
	serviceDef := []ServiceDef{{
		ID:              schemaID,
		Type:            "schema",
		ServiceEndpoint: schemaID,
	}}

	input := generateDIDDocInput{
		DID:                  did,
		FullyQualifiedKeyRef: keyRef,
		SigningKey:           issuerPrivKey,
		PublicKeys:           publicKeys,
		Issuer:               issuer,
		Services:             serviceDef,
	}

	didDoc, err := generateDIDDoc(input)
	assert.NoError(t, err, "Error was not expected when creating did doc")
	assert.Equal(t, didDoc.ID, did)
	assert.Equal(t, didDoc.UnsignedDIDDoc.PublicKey[0].Controller, issuer)
	assert.Equal(t, didDoc.Service[0].ID, schemaID)
	assert.Equal(t, didDoc.PublicKey[0].PublicKeyBase58, base58.Encode(issuerPubKey))

	// uncomment me and set a break point when you need to generate a new did doc
	//didDocBytes, _ := json.Marshal(didDoc)
	//didDocString := string(didDocBytes)
	//assert.NotEmpty(t, didDocString)

	verifyDIDDoc(t, *didDoc, issuerPubKey)
}

func verifyDIDDoc(t *testing.T, doc DIDDoc, pubKey ed25519.PublicKey) {
	assert.Len(t, doc.PublicKey, 1)
	assert.NoError(t, ValidateDIDDocProof(doc, pubKey))

	pubK1Bytes, _ := base58.Decode(doc.PublicKey[0].PublicKeyBase58)

	assert.Equal(t, doc.ID, "did:work:"+base58.Encode(pubK1Bytes[:16]))
}

type generateDIDDocInput struct {
	DID                  string
	FullyQualifiedKeyRef string
	SigningKey           ed25519.PrivateKey
	PublicKeys           map[string]*ed25519.PublicKey
	Issuer               string
	Services             []ServiceDef
}

func generateDIDDoc(input generateDIDDocInput) (*DIDDoc, error) {
	var didPubKeys []KeyDef
	for k, v := range input.PublicKeys {
		keyEntry := KeyDef{
			ID:              input.DID + "#" + k,
			Type:            proof.JCSVerificationType,
			Controller:      input.Issuer,
			PublicKeyBase58: base58.Encode(*v),
		}
		didPubKeys = append(didPubKeys, keyEntry)
	}

	doc := UnsignedDIDDoc{
		ID:        input.DID,
		PublicKey: didPubKeys,
		Service:   input.Services,
	}
	signedDoc, err := SignDIDDoc(doc, input.SigningKey, input.FullyQualifiedKeyRef)
	if err != nil {
		return nil, err
	}

	return signedDoc, nil
}
